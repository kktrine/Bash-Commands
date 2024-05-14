package cashe

import (
	"sync"
)

type Cache struct {
	sync.RWMutex
	Pids map[int]interface{}
}

func NewCache() *Cache {

	pids := make(map[int]interface{})
	cache := Cache{
		Pids: pids,
	}

	return &cache
}

func (c *Cache) AddOne(pid int) {
	c.Lock()
	defer c.Unlock()
	c.Pids[pid] = nil
}

func (c *Cache) Stop(pid int) {
	c.Lock()
	defer c.Unlock()
	delete(c.Pids, pid)
}

func (c *Cache) Check(pid int) bool {
	c.Lock()
	defer c.Unlock()
	_, ok := c.Pids[pid]
	return ok
}
