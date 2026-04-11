package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	_ "modernc.org/sqlite"
)

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

func (c *LRUCache) moveToFront(n *node) {
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

var db *sql.DB
var cache = NewLRUCache(1000)

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "users.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, name TEXT, avatar BLOB)`)
	if err != nil {
		log.Fatal(err)
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u User

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if u.ID == "" || len(u.Avatar) == 0 {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if len(u.Avatar) > 20*1024 {
		http.Error(w, "avatar so large", http.StatusBadRequest)
	}
	_, err := db.Exec("INSERT OR REPLACE INTO users(id, name, avatar) VALUES(?, ?, ?)")

	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	cache.Put(u.ID, u)
	w.WriteHeader(http.StatusCreated)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	if u, ok := cache.Get(id); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
		return
	}

	var u User
	err := db.QueryRow("SELECT id, name, avatar FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name, &u.Avatar)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
	}
	cache.Put(id, u)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func main() {
	initDB()

}
