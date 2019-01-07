package config

import "sync"

// CustomLock is a sync RWLock which adds an operation to check if it is locked
type CustomLock struct {
	locked bool
	sync.RWMutex
}

// IsLock returns true if CustomLock is locked
func (c *CustomLock) IsLock() bool {
	return c.locked
}

// Lock locks CustomLock
func (c *CustomLock) Lock() {
	c.RLock()
	c.locked = true
	c.RUnlock()
}

// Unlock unlocks CustomLock
func (c *CustomLock) Unlock() {
	c.RLock()
	c.locked = false
	c.RUnlock()
}
