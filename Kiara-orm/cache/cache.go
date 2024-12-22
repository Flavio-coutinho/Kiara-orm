package cache

import (
	"context"
	"sync"
	"time"
)

// Item representa um item no cache
type Item struct {
	Value      interface{}
	Expiration int64
}

// Cache implementa um cache em memória com expiração
type Cache struct {
	items sync.Map
	mu    sync.RWMutex
}

// NewCache cria uma nova instância do Cache
func NewCache() *Cache {
	cache := &Cache{}
	go cache.janitor()
	return cache
}

// Set adiciona um item ao cache
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	expiration := time.Now().Add(duration).UnixNano()
	c.items.Store(key, Item{
		Value:      value,
		Expiration: expiration,
	})
}

// Get obtém um item do cache
func (c *Cache) Get(key string) (interface{}, bool) {
	item, exists := c.items.Load(key)
	if !exists {
		return nil, false
	}
	
	cached := item.(Item)
	if time.Now().UnixNano() > cached.Expiration {
		c.items.Delete(key)
		return nil, false
	}
	
	return cached.Value, true
}

// Delete remove um item do cache
func (c *Cache) Delete(key string) {
	c.items.Delete(key)
}

// Clear limpa todo o cache
func (c *Cache) Clear() {
	c.items = sync.Map{}
}

// janitor limpa itens expirados periodicamente
func (c *Cache) janitor() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		now := time.Now().UnixNano()
		c.items.Range(func(key, value interface{}) bool {
			item := value.(Item)
			if now > item.Expiration {
				c.items.Delete(key)
			}
			return true
		})
	}
} 