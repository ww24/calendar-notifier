package calendar

import (
	"sync"
)

type syncToken struct {
	mu    sync.RWMutex
	token string
}

func (c *syncToken) update(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

func (c *syncToken) get() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}
