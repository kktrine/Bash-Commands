package cashe

import (
	"io"
	"sync"
)

type Cache struct {
	sync.RWMutex
	Pids map[int]Item
}

type Item struct {
	stdOut io.ReadCloser
	stdErr io.ReadCloser
}

func NewCache() *Cache {

	pids := make(map[int]Item)
	cache := Cache{
		Pids: pids,
	}

	return &cache
}

func (c *Cache) AddOne(pid int, stdOut io.ReadCloser, stdErr io.ReadCloser) {
	c.Lock()
	defer c.Unlock()
	c.Pids[pid] = Item{
		stdOut: stdOut,
		stdErr: stdErr,
	}
}

func (c *Cache) GetOne(pid int) (io.ReadCloser, io.ReadCloser) {
	proc, ok := c.Pids[pid]
	if !ok {
		return nil, nil
	}
	return proc.stdOut, proc.stdErr
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
