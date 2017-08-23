package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewCacheManager() (CacheFactory, error) {
	err := ensureFileExists(".data/cache.db")
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(".data/cache.db", 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, errors.WithMessage(err, "Can't open cache")
	}

	cacheLogger := logrus.StandardLogger()
	cacheLogger.SetLevel(logrus.DebugLevel)

	return &StandardCacheManager{
		db:     db,
		logger: cacheLogger,
	}, nil
}

type CacheFactory interface {
	Get(cacheName string, ttl time.Duration) Cache
}

type Cache interface {
	Save(key string, value interface{}) error
	Load(key string, value interface{}) error
}

type StandardCacheManager struct {
	logger *logrus.Logger
	db     *bolt.DB
}

func (scm *StandardCacheManager) Get(cacheName string, ttl time.Duration) Cache {
	return &boltCache{
		db:        scm.db,
		cacheName: cacheName,
		logger:    scm.logger,
	}
	// &expirableCache{
	// 	ttl:    ttl,
	// 	logger: scm.logger,
	// 	cache: &boltCache{
	// 		db:        scm.db,
	// 		cacheName: cacheName,
	// 		logger:    scm.logger,
	// 	},
	// }
}

type boltCache struct {
	db        *bolt.DB
	cacheName string
	logger    *logrus.Logger
}

// Save new value to the cache using provided key
func (c *boltCache) Save(key string, value interface{}) error {
	c.logger.Debugf("Saving value to cache [%s] using [%s] key", c.cacheName, key)

	return c.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(c.cacheName))
		if err != nil {
			c.logger.Errorln(err.Error())

			return fmt.Errorf("Cannot create bucket: %s", err)
		}

		// Marshal user data into bytes.
		buf, err := json.Marshal(value)
		if err != nil {
			return errors.WithMessage(err, "Cannot marshal value to cache")
		}

		// Persist bytes to users bucket.
		return bucket.Put([]byte(key), buf)
	})
}

// Load value by key from the specified cache
func (c *boltCache) Load(key string, value interface{}) (err error) {
	c.logger.Debugf("Load value from cache [%s] using [%s] key", c.cacheName, key)

	err = c.db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(c.cacheName))
		if bucket == nil {
			c.logger.Debugln("No bucket to read")
			return nil
		}
		cur := bucket.Cursor()

		key := []byte(key)
		for k, v := cur.Seek(key); bytes.Equal(k, key); k, v = cur.Next() {
			c.logger.Debugln("Unmarshaling cache value")
			return json.Unmarshal(v, value)
		}

		c.logger.Debugln("No element in cache with provided key")
		return nil
	})

	return
}

type expirableCache struct {
	ttl    time.Duration
	cache  Cache
	logger *logrus.Logger
}

type expirableCacheItem struct {
	Item      interface{} `json:"item,omitempty"`
	CreatedAt time.Time   `json:"created_at,omitempty"`
}

func (c *expirableCache) Save(key string, value interface{}) error {
	c.logger.Debugf("Saving expirable item. Key: %s Now: %s", key, time.Now())
	return c.cache.Save(key, &expirableCacheItem{Item: value, CreatedAt: time.Now()})
}

func (c *expirableCache) Load(key string, value interface{}) (err error) {
	copy := reflect.New(reflect.TypeOf(value)).Elem().Interface()

	item := &expirableCacheItem{Item: copy}
	err = c.cache.Load(key, item)
	if err != nil {
		return
	}

	if item.CreatedAt.Add(c.ttl).Before(time.Now()) {
		return
	}

	// c.logger.WithFields(logrus.Fields{
	// 	"createAt": item.CreatedAt,
	// 	"ttl":      c.ttl,
	// 	"now":      time.Now(),
	// }).Debugln("Expirable entry has been loaded from cache")

	c.logger.Debugf("Expirable entry has been loaded from cache: %v+", copy)
	value = copy
	return
}
