package transport

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
)

// RoutingTable represents a routing table implementation.
type RoutingTable interface {

	// AddRule adds a new RoutingRules to the table and returns assigned RouteID.
	AddRule(rule RoutingRule) (routeID RouteID, err error)

	// SetRule updates RoutingRule with a given RouteID.
	SetRule(routeID RouteID, rule RoutingRule) error

	// Rule returns RoutingRule with a given RouteID.
	Rule(routeID RouteID) (rule RoutingRule, err error)

	// DeleteRule removes RoutingRule with a given a RouteID.
	DeleteRule(routeID RouteID) error

	// Count returns the number of RoutingRule entries stored.
	Count() int
}

type inMemoryRoutingTable struct {
	sync.RWMutex

	nextID uint32
	rules  map[RouteID]RoutingRule
}

// InMemoryRoutingTable return in-memory RoutingTable implementation.
func InMemoryRoutingTable() RoutingTable {
	return &inMemoryRoutingTable{
		rules: map[RouteID]RoutingRule{},
	}
}

func (rt *inMemoryRoutingTable) AddRule(rule RoutingRule) (routeID RouteID, err error) {
	if routeID == math.MaxUint32 {
		return 0, errors.New("no available routeIDs")
	}

	routeID = RouteID(atomic.AddUint32(&rt.nextID, 1))

	rt.Lock()
	rt.rules[routeID] = rule
	rt.Unlock()

	return routeID, nil
}

func (rt *inMemoryRoutingTable) SetRule(routeID RouteID, rule RoutingRule) error {
	if routeID == InitRouteID {
		return errors.New("reserved routeID")
	}

	rt.Lock()
	rt.rules[routeID] = rule
	rt.Unlock()

	return nil
}

func (rt *inMemoryRoutingTable) Rule(routeID RouteID) (rule RoutingRule, err error) {
	rt.RLock()
	rule = rt.rules[routeID]
	rt.RUnlock()

	return rule, nil
}

func (rt *inMemoryRoutingTable) DeleteRule(routeID RouteID) error {
	rt.Lock()
	delete(rt.rules, routeID)
	rt.Unlock()

	return nil
}

func (rt *inMemoryRoutingTable) Count() int {
	rt.RLock()
	count := len(rt.rules)
	rt.RUnlock()
	return count
}
