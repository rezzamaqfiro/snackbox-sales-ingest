package main

import "sync"

type StockCache struct {
	mu   sync.RWMutex
	data map[string]int
}

func NewStockCache() *StockCache {
	return &StockCache{
		data: make(map[string]int),
	}
}

func (c *StockCache) Get(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *StockCache) Set(key string, val int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = val
}
