package cache

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	capacity int
	list     *list.List
	items    map[interface{}]*list.Element
	mutex    sync.RWMutex
}

type entry struct {
	key   interface{}
	value interface{}
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		list:     list.New(),
		items:    make(map[interface{}]*list.Element),
		mutex:    sync.RWMutex{},
	}
}

func (c *LRUCache) Update(key interface{}) (ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var elem *list.Element
	if elem, ok = c.items[key]; ok {
		c.list.MoveToFront(elem)
		return
	}
	return
}

func (c *LRUCache) Add(key, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	ele := c.list.PushFront(&entry{key, value})
	c.items[key] = ele

	if c.list.Len() > c.capacity {
		lastElem := c.list.Back()
		if lastElem != nil {
			c.list.Remove(lastElem)
			delete(c.items, lastElem.Value.(*entry).key)
		}
	}
}

func (c *LRUCache) Remove(key interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ele, hit := c.items[key]; hit {
		c.list.Remove(ele)
		delete(c.items, key)
	}
}

func (c *LRUCache) Get(key interface{}, move bool) (value interface{}, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ele, hit := c.items[key]; hit {
		if move {
			c.list.MoveToFront(ele)
		}
		return ele.Value.(*entry).value, true
	}
	return nil, false
}

func (c *LRUCache) GetByIndex(index int) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if index < 0 || index > c.list.Len()-1 {
		return nil, false
	}
	e := c.list.Front()
	for i := 0; i < index; i++ {
		e = e.Next()
	}

	if e != nil {
		//c.list.MoveToFront(e)
		return e.Value.(*entry).value, true
	}

	return nil, false
}

func countUniqueValues(values []interface{}) int {
	uniqueValues := make(map[interface{}]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(uniqueValues)
}

func (c *LRUCache) GetList(limit int) []interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if limit <= 0 || limit > c.list.Len() {
		limit = c.list.Len()
	}

	all := make([]interface{}, 0, limit)

	for e, i := c.list.Front(), 0; e != nil && i < limit; e, i = e.Next(), i+1 {
		all = append(all, e.Value.(*entry).value)
	}
	return all
}
