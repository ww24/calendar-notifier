package main

import (
	"sync/atomic"

	"github.com/ww24/calendar-notifier"
)

type itemsCache struct {
	value *atomic.Value
}

func newItemsCache() *itemsCache {
	return &itemsCache{
		value: &atomic.Value{},
	}
}

func (c *itemsCache) SetFromList(items []*calendar.EventItem) {
	m := make(map[string]*calendar.EventItem, len(items))
	for _, e := range items {
		m[e.ID] = e
	}
	c.value.Store(m)
}

func (c *itemsCache) Get() map[string]*calendar.EventItem {
	v := c.value.Load()
	if v == nil {
		return nil
	}
	return v.(map[string]*calendar.EventItem)
}
