package sql

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/tests"
)

func newTestStore(t *testing.T) *Store {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skipf("POSTGRES_TEST_DSN not defined, can't run Postgres tests")
	}

	s, err := NewStore(dsn)
	require.NoError(t, err)
	for _, q := range []string{
		"DROP TABLE transports_ack",
		"DROP TABLE transports",
	} {
		s.db.Exec(q)
	}

	require.NoError(t, s.Migrate(context.Background()))
	return s
}

func TestSQL(t *testing.T) {
	s := newTestStore(t)

	suite.Run(t, &tests.TransportSuite{Store: s})
}
