package main

import "sync"

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type node struct {
	key   string
	value User
	prev  *node
	next  *node
}

type LRUCache struct {
	capacity int
	items    map[string]*node
	head     *node
	tail     *node
	mutex    sync.Mutex
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*node),
	}
}

func (c *LRUCache) Get(key string) (User, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, ok := c.items[key]; ok {
		c.moveToFront(node)
		return node.value, true
	}
	return User{}, false
}

func (c *LRUCache) Put(key string, value User) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, ok := c.items[key]; ok {
		node.value = value
		c.moveToFront(node)
		return
	}
	node := &node{key: key, value: value}
	c.items[key] = node
	c.addToFront(node)

	if len(c.items) == c.capacity {
		c.removeOldest()
	}
}

func (c *LRUCache) moveToFront(node *node) {
	c.remove(n)
	c.addToFront(n)
}

func (c *LRUCache) addToFront(n *node) {
	n.next = c.head
	n.prev = nil

	if c.head != nil {
		c.head.prev = n
	}
	c.head = n

	if c.tail == nil {
		c.tail = n
	}
}

func (c *LRUCache) remove(n *node) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		c.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		c.tail = n.prev
	}
}
func (c *LRUCache) removeOldest() {
	if c.tail == nil {
		return
	}
	oldest := c.tail
	c.remove(oldest)
	delete(c.items, oldest.key)
}
func main() {

}
