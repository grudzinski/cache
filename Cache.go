package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type Cache struct {
	max      int
	ttl      time.Duration
	getValue func(interface{}) (interface{}, error)

	keyToEntryLock sync.RWMutex
	keyToEntry     map[interface{}]*tEntry
	list           list.List
}

func New() *Cache {
	cache := new(Cache)
	cache.Init()
	return cache
}

func (cache *Cache) Init() {
	cache.keyToEntry = make(map[interface{}]*tEntry)
	cache.list.Init()
}

func (cache *Cache) remove(key interface{}, entry *tEntry) {
	keyToEntryLock := cache.keyToEntryLock
	keyToEntryLock.Lock()
	delete(cache.keyToEntry, key)
	keyToEntryLock.Unlock()
	cache.list.Remove(entry.element)
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
			cache.remove(key, entry)
			return
		}
		value, err := cache.getValue(key)
		if err != nil {
			cache.remove(key, entry)
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
	list := &cache.list
	if ok {
		list.MoveToFront(entry.element)
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
	if len(keyToEntry) == cache.max {
		back := list.Back()
		backKey := list.Remove(back)
		delete(keyToEntry, backKey)
	}
	keyToEntry[key] = entry
	keyToEntryLock.Unlock()
	entry.element = list.PushFront(key)
	go cache.scheduleUpdateOrRemove(key)
	return value, nil
}
