package cashe

import (
	"os"
	"sync"
	"time"
)

type Cache struct {
	sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	Items             map[string]Item
}
type Item struct {
	Id        int64
	Pid       int64
	CreatedAt time.Time
	Content   string
}

func NewCache() *Cache {

	items := make(map[string]Item)
	exp, err := time.ParseDuration(os.Getenv("CACHE_EXPIRATION"))
	if err != nil {
		panic("Can't parse CACHE_EXPIRATION: " + err.Error())
	}
	clean, err := time.ParseDuration(os.Getenv("CACHE_CLEANUP_INTERVAL"))
	if err != nil {
		panic("Can't parse CACHE_CLEANUP_INTERVAL: " + err.Error())
	}
	cache := Cache{
		Items:             items,
		defaultExpiration: exp,
		cleanupInterval:   clean,
	}

	cache.startGC()

	return &cache
}

func (c *Cache) AddOne(command Item) {
	c.Lock()
	defer c.Unlock()

}
