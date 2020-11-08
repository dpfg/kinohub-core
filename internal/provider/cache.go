package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewCacheFactory(path string, logger *logrus.Logger) (CacheFactory, error) {
	err := ensureFileExists(path + "cache.db")
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(path+"cache.db", 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, errors.WithMessage(err, "Can't open cache")
	}

	return &StandardCacheManager{
		db:     db,
		logger: logger.WithFields(logrus.Fields{"prefix": "cache"}),
	}, nil
}

type CacheFactory interface {
	Get(cacheName string, ttl time.Duration) Cache
}

type Cache interface {
	Save(key string, value CacheEntry)
	Load(key string, value CacheEntry) bool
}

type CacheEntry interface {
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
}

type cacheable struct {
	entry interface{}
}

func (c cacheable) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c.entry)
}

func (c cacheable) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &c.entry)
}

func Cacheable(m interface{}) CacheEntry {
	return &cacheable{entry: &m}
}

type StandardCacheManager struct {
	logger *logrus.Entry
	db     *bolt.DB
}

func (scm *StandardCacheManager) Get(cacheName string, ttl time.Duration) Cache {
	return &expirableCache{
		ttl:    ttl,
		logger: scm.logger,
		cache: &multicastCache{
			caches: []Cache{
				&inMemoryCache{cache: make(map[string][]byte)},
				&boltCache{
					db:        scm.db,
					cacheName: cacheName,
					logger:    scm.logger,
				},
			},
		},
	}
}

type boltCache struct {
	db        *bolt.DB
	cacheName string
	logger    *logrus.Entry
}

// Save new value to the cache using provided key
func (c *boltCache) Save(key string, value CacheEntry) {
	c.logger.Debugf("Saving value to cache [%s] using [%s] key", c.cacheName, key)

	err := c.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(c.cacheName))
		if err != nil {
			c.logger.Errorln(err.Error())

			return fmt.Errorf("Cannot create bucket: %s", err)
		}

		// Marshal user data into bytes.
		buf, err := value.MarshalBinary()
		if err != nil {
			return errors.WithMessage(err, "Cannot marshal value to cache")
		}

		// Persist bytes to users bucket.
		return bucket.Put([]byte(key), buf)
	})

	if err != nil {
		panic(err.Error())
	}
}

// Load value by key from the specified cache
func (c *boltCache) Load(key string, value CacheEntry) bool {
	c.logger.Debugf("Load value from cache [%s] using [%s] key", c.cacheName, key)

	var loaded bool

	err := c.db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(c.cacheName))
		if bucket == nil {
			c.logger.Debugln("No bucket to read")
			return nil
		}
		cur := bucket.Cursor()

		key := []byte(key)
		for k, v := cur.Seek(key); bytes.Equal(k, key); k, v = cur.Next() {
			c.logger.Debugln("Unmarshaling cache value")
			loaded = true
			return value.UnmarshalBinary(v)
		}

		c.logger.Debugln("No element in cache with provided key")
		return nil
	})

	if err != nil {
		panic(err.Error())
	}

	return loaded
}

type expirableCache struct {
	ttl    time.Duration
	cache  Cache
	logger *logrus.Entry
}

func (c *expirableCache) Save(key string, value CacheEntry) {
	c.logger.Debugf("Saving expirable item. Key: %s Now: %s", key, time.Now())
	// save experation
	c.cache.Save("_CREATED_AT_:"+key, Cacheable(time.Now()))
	// save data
	c.cache.Save(key, value)
}

func (c *expirableCache) Load(key string, value CacheEntry) bool {
	c.logger.Debugln("Loading from expirable cache")

	createdAt := &time.Time{}
	c.cache.Load("_CREATED_AT_:"+key, Cacheable(createdAt))

	c.logger.Debugf("Created at: [%v]", createdAt)
	if createdAt.Add(c.ttl).Before(time.Now()) {
		c.logger.Debugln("Item has been expired.")
		return false
	}

	return c.cache.Load(key, value)
}

type inMemoryCache struct {
	cache map[string][]byte
}

func (c *inMemoryCache) Save(key string, value CacheEntry) {
	data, err := value.MarshalBinary()
	if err != nil {
		panic(err.Error())
	}

	c.cache[key] = data
}

func (c *inMemoryCache) Load(key string, value CacheEntry) bool {
	if data, ok := c.cache[key]; ok {
		err := value.UnmarshalBinary(data)
		if err != nil {
			return false
		}

		return true
	}

	return false
}

type multicastCache struct {
	caches []Cache
}

func (c *multicastCache) Save(key string, value CacheEntry) {
	for _, cache := range c.caches {
		cache.Save(key, value)
	}
}

func (c *multicastCache) Load(key string, value CacheEntry) bool {
	for index, cache := range c.caches {
		if cache.Load(key, value) {
			// persist data to the first cache
			if index == 0 && cache != c.caches[index] {
				c.caches[index].Save(key, value)
			}
			return true
		}
	}

	return false
}
