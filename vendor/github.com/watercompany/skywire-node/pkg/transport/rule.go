package transport

import (
	"encoding/binary"

	"github.com/skycoin/skycoin/src/cipher"
)

// RoutingOperation represent operation type of the RoutingRule.
type RoutingOperation byte

const (
	// RoutingOperationLoopback defines loopback rule type.
	RoutingOperationLoopback RoutingOperation = iota
	// RoutingOperationForward defines forward rule type.
	RoutingOperationForward
)

// RoutingRule represents a rule in a RoutingTable.
type RoutingRule []byte

// LoopbackRule constructs a new loopback RoutingRule.
func LoopbackRule(port uint16) RoutingRule {
	rule := []byte{byte(RoutingOperationLoopback), 0, 0}
	binary.LittleEndian.PutUint16(rule[1:], port)
	return RoutingRule(rule)
}

// ForwardRule constructs a new forward RoutingRule.
func ForwardRule(remotePK cipher.PubKey, ruleID RouteID) RoutingRule {
	rule := []byte{byte(RoutingOperationForward)}
	rule = append(rule, remotePK[:]...)
	rule = append(rule, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(rule[34:], uint32(ruleID))
	return RoutingRule(rule)
}
