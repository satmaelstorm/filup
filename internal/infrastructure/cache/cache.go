package cache

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
)

type Cache struct {
	controller  *lru.Cache
	errorLogger logsEngine.ILogger
}

func ProvideMetaCache(
	cfg config.Configuration,
	logger logsEngine.ILogger,
) (*Cache, error) {
	c, err := lru.New(cfg.Caches.Parts.Size)
	if err != nil {
		return nil, err
	}
	return &Cache{controller: c, errorLogger: logger}, nil
}

func (c *Cache) Add(key string, value []byte) {
	c.controller.Add(key, value)
}

func (c *Cache) Get(key string) ([]byte, bool) {
	v, ok := c.controller.Get(key)
	if !ok {
		return nil, false
	}
	if nil == v {
		return nil, true
	}
	r, ok := v.([]byte)
	if !ok {
		c.errorLogger.Error().Println("not []byte in parts cache under key " + key)
		return nil, false
	}
	return r, true
}
