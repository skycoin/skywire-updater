package sql

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type Store struct {
	db *sql.DB
}

type Edges []cipher.PubKey

func (e *Edges) Scan(value interface{}) error {
	var edges []string
	if err := pq.Array(&edges).Scan(value); err != nil {
		return err
	}

	for _, hex := range edges {
		pk, err := cipher.PubKeyFromHex(hex)
		if err != nil {
			return err
		}

		*e = append(*e, pk)
	}

	return nil
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &Store{db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// TODO: it should recover from panics
func (s *Store) withinTx(fn func(tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// RegisterTransport creates a transport in the store and wait until the other node Register the transport.
// RegisterTransport is idempotent.
func (s *Store) RegisterTransport(ctx context.Context, t *store.Transport) error {
	var query string

	var edges = [2]string{
		t.Edges[0].Hex(),
		t.Edges[1].Hex(),
	}

	fn := func(tx *sql.Tx) error {
		// Find or Create transport (idempotency)
		query = `SELECT id, registered FROM transports WHERE edges @> ARRAY[$1, $2]::VARCHAR(66)[]`
		if err := tx.QueryRowContext(ctx, query, edges[0], edges[1]).Scan(&t.ID, &t.Registered); err != nil {
			if err != sql.ErrNoRows {
				return err
			}

			query = `INSERT INTO transports (edges) VALUES(ARRAY[$1, $2]) RETURNING id, registered`
			if err := tx.QueryRowContext(ctx, query, edges[0], edges[1]).Scan(&t.ID, &t.Registered); err != nil {
				return err
			}
		}

		// Add our ACK
		query = `INSERT INTO transports_ack VALUES($1, $2)`
		if _, err := tx.ExecContext(ctx, query, t.ID, edges[0]); err != nil {
			return err
		}

		return nil

	}

	if err := s.withinTx(fn); err != nil {
		return err
	}

	return s.waitForTransport(ctx, t.ID, 1*time.Second)
}

func (s *Store) waitForTransport(ctx context.Context, id store.ID, delay time.Duration) error {
	for {
		_, err := s.GetTransportByID(ctx, id)
		// No Error means that transport was found; thus, created.
		if err == nil {
			return nil
		}

		// Any error different from these should be reported
		if err != sql.ErrNoRows && err != store.ErrNotEnoughACKs {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

func (s *Store) GetTransportByID(ctx context.Context, id store.ID) (*store.Transport, error) {
	var query string
	var acks int

	var t = &store.Transport{ID: id}
	fn := func(tx *sql.Tx) error {
		var edges []string

		query = `SELECT edges, registered FROM transports WHERE id = $1`
		if err := tx.QueryRowContext(ctx, query, id).Scan(pq.Array(&edges), &t.Registered); err != nil {
			return err
		}

		// TODO: use a Scanner to de duplicate this from other places
		pk1, err := cipher.PubKeyFromHex(edges[0])
		if err != nil {
			return err
		}

		pk2, err := cipher.PubKeyFromHex(edges[1])
		if err != nil {
			return err
		}
		t.Edges = []cipher.PubKey{pk1, pk2}

		query = `SELECT COUNT(*) FROM transports_ack WHERE
			transport_id = $1 AND node in ($2, $3)
		`
		if err := tx.QueryRowContext(ctx, query, id, edges[0], edges[1]).Scan(&acks); err != nil {
			return err
		}

		return nil
	}

	if err := s.withinTx(fn); err != nil {
		return nil, err
	}

	if acks < 2 {
		return nil, store.ErrNotEnoughACKs
	}

	return t, nil
}

func (s *Store) GetTransportsByEdge(ctx context.Context, edge cipher.PubKey) ([]*store.Transport, error) {
	var query = ` SELECT id, edges, registered FROM transports WHERE edges @> ARRAY[$1]::VARCHAR(66)[]`
	rows, err := s.db.QueryContext(ctx, query, edge.Hex())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ts []*store.Transport
	for rows.Next() {
		var t store.Transport
		var edges = Edges{}

		if err := rows.Scan(&t.ID, &edges, &t.Registered); err != nil {
			return nil, err
		}

		t.Edges = edges

		ts = append(ts, &t)
	}

	return ts, nil
}

func (s *Store) DeregisterTransport(ctx context.Context, id store.ID) (*store.Transport, error) {
	var t store.Transport

	fn := func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM transports_ack WHERE transport_id = $1`, id); err != nil {
			return err
		}

		row := tx.QueryRowContext(ctx, `DELETE FROM transports WHERE id = $1 RETURNING registered, edges`, id)

		var edges []string
		if err := row.Scan(&t.Registered, pq.Array(&edges)); err != nil {
			return err
		}

		pk1, err := cipher.PubKeyFromHex(edges[0])
		if err != nil {
			return err
		}

		pk2, err := cipher.PubKeyFromHex(edges[1])
		if err != nil {
			return err
		}
		t.Edges = []cipher.PubKey{pk1, pk2}

		return nil
	}

	if err := s.withinTx(fn); err != nil {
		return nil, err
	}

	t.ID = id
	return &t, nil
}

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS transports (
		id SERIAL PRIMARY KEY NOT NULL,
		edges VARCHAR(66)[] NOT NULL,
		registered TIMESTAMP DEFAULT Now()
	)`,
	`CREATE INDEX IF NOT EXISTS
	  transports_edges_idx on transports USING GIN ("edges")`,

	`CREATE TABLE IF NOT EXISTS transports_ack (
		transport_id INTEGER REFERENCES transports(id),
		node VARCHAR(66),
		PRIMARY KEY (transport_id, node)
	)`,
}

func (s *Store) Migrate(ctx context.Context) error {
	for _, m := range migrations {
		if _, err := s.db.ExecContext(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
