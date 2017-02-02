package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

type Cache struct {
	ttl      time.Duration
	getValue func(interface{}) (interface{}, error)

	keyToEntryLock sync.RWMutex
	keyToEntry     map[interface{}]*tEntry
}

func New() *Cache {
	return &Cache{
		keyToEntry: make(map[interface{}]*tEntry),
	}
}

func (cache *Cache) scheduleUpdateOrRemove(key interface{}) {
	for {
		time.Sleep(cache.ttl)
		keyToEntry := cache.keyToEntry
		keyToEntryLock := cache.keyToEntryLock
		keyToEntryLock.RLock()
		entry := keyToEntry[key]
		keyToEntryLock.RUnlock()
		if atomic.LoadUint32(&entry.hit) == 0 {
			keyToEntryLock.Lock()
			delete(keyToEntry, key)
			keyToEntryLock.Unlock()
			return
		}
		value, err := cache.getValue(key)
		if err != nil {
			keyToEntryLock.Lock()
			delete(keyToEntry, key)
			keyToEntryLock.Unlock()
			return
		}
		entry.value.Store(value)
		atomic.StoreUint32(&entry.hit, 0)
	}
}

func (cache *Cache) Get(key interface{}) (interface{}, error) {
	keyToEntry := cache.keyToEntry
	keyToEntryLock := &cache.keyToEntryLock
	keyToEntryLock.RLock()
	entry, ok := keyToEntry[key]
	keyToEntryLock.RUnlock()
	if ok {
		atomic.StoreUint32(&entry.hit, 1)
		return entry.value.Load(), nil
	}
	value, err := cache.getValue(key)
	if err != nil {
		return nil, err
	}
	entry = new(tEntry)
	entry.value.Store(value)
	keyToEntryLock.Lock()
	keyToEntry[key] = entry
	keyToEntryLock.Unlock()
	go cache.scheduleUpdateOrRemove(key)
	return value, nil
}
