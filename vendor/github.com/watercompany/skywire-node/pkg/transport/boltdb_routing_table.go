package transport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	bolt "go.etcd.io/bbolt"
)

var boltDBBucket = []byte("routing")

// BoltDBRoutingTable implements RoutingTable on top of BoltDB.
type BoltDBRoutingTable struct {
	db *bolt.DB
}

// NewBoltDBRoutingTable consturcts a new BoldDBRoutingTable.
func NewBoltDBRoutingTable(path string) (*BoltDBRoutingTable, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucket(boltDBBucket); err != nil {
			return fmt.Errorf("failed to create bucket: %s", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &BoltDBRoutingTable{db}, nil
}

// AddRule adds routing to the tabled and returns assigned Route ID.
func (rt *BoltDBRoutingTable) AddRule(rule RoutingRule) (routeID RouteID, err error) {
	err = rt.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltDBBucket)
		nextID, _ := b.NextSequence() // nolint

		if nextID > math.MaxUint32 {
			return errors.New("no available routeIDs")
		}

		routeID = RouteID(nextID)
		return b.Put(binaryID(routeID), []byte(rule))
	})

	return routeID, err
}

// SetRule updates RoutingRule with a given RouteID.
func (rt *BoltDBRoutingTable) SetRule(routeID RouteID, rule RoutingRule) error {
	if routeID == InitRouteID {
		return errors.New("reserved routeID")
	}

	return rt.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltDBBucket)

		return b.Put(binaryID(routeID), []byte(rule))
	})
}

// Rule obtains a routing rule with a given route ID.
func (rt *BoltDBRoutingTable) Rule(routeID RouteID) (rule RoutingRule, err error) {
	err = rt.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltDBBucket)

		rule = b.Get(binaryID(routeID))
		return nil
	})

	return rule, err
}

// DeleteRule removes rule with a given a route ID.
func (rt *BoltDBRoutingTable) DeleteRule(routeID RouteID) error {
	return rt.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltDBBucket)

		return b.Delete(binaryID(routeID))
	})
}

// Count returns the number of routing rules stored.
func (rt *BoltDBRoutingTable) Count() (count int) {
	err := rt.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(boltDBBucket)

		stats := b.Stats()
		count = int(stats.KeyN)
		return nil
	})
	if err != nil {
		return 0
	}

	return count
}

// Close closes underlying BoltDB instance.
func (rt *BoltDBRoutingTable) Close() error {
	return rt.db.Close()
}

func binaryID(v RouteID) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	return b
}
