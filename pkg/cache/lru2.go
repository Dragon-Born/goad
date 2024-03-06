package cache

import (
	"sync"
)

type LRUCache2 struct {
	capacity int
	items    []*entry
	mapping  map[interface{}]int
	mutex    sync.RWMutex
}

func NewLRUCache2(capacity int) *LRUCache2 {
	return &LRUCache2{
		capacity: capacity,
		items:    make([]*entry, 0, capacity),
		mapping:  make(map[interface{}]int),
		mutex:    sync.RWMutex{},
	}
}

func (c *LRUCache2) Add(key, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if index, ok := c.mapping[key]; ok {
		c.items[index].value = value
		c.updateLRU(index)
		return
	}

	if len(c.items) >= c.capacity {
		// Deal with capacity overflow, remove last element
		delete(c.mapping, c.items[len(c.items)-1].key)
		c.items = c.items[:len(c.items)-1]
	}

	c.items = append([]*entry{{key, value}}, c.items...)
	c.mapping[key] = 0
}

func (c *LRUCache2) Get(key interface{}) (value interface{}, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if index, ok := c.mapping[key]; ok {
		c.updateLRU(index)
		return c.items[0].value, true
	}
	return nil, false
}

func (c *LRUCache2) GetByIndex(index int) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if index < 0 || index >= len(c.items) {
		return nil, false
	}

	c.updateLRU(index)
	return c.items[0].value, true
}

func (c *LRUCache2) updateLRU(index int) {
	updated := c.items[index]
	copy(c.items[1:], c.items[:index])
	c.items[0] = updated
	for i := 0; i <= index; i++ {
		c.mapping[c.items[i].key] = i
	}
}
