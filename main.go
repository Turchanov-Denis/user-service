package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const shardCount = 100

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar []byte `json:"avatar"`
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

	if n, ok := c.items[key]; ok {
		c.moveToFront(n)
		return n.value, true
	}
	return User{}, false
}

func (c *LRUCache) Put(key string, value User) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if n, ok := c.items[key]; ok {
		n.value = value
		c.moveToFront(n)
		return
	}

	n := &node{key: key, value: value}
	c.items[key] = n
	c.addToFront(n)

	if len(c.items) > c.capacity {
		c.removeOldest()
	}
}

func (c *LRUCache) moveToFront(n *node) {
	c.remove(n)
	c.addToFront(n)
}

func (c *LRUCache) addToFront(n *node) {
	n.prev = nil
	n.next = c.head
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

	n.prev = nil
	n.next = nil
}

func (c *LRUCache) removeOldest() {
	if c.tail == nil {
		return
	}
	oldest := c.tail
	c.remove(oldest)
	delete(c.items, oldest.key)
}

type ShardedLRU struct {
	shards [shardCount]*LRUCache
}

func NewShardedLRU(capacity int) *ShardedLRU {
	perShard := capacity / shardCount
	if perShard == 0 {
		perShard = 1
	}

	c := &ShardedLRU{}
	for i := 0; i < shardCount; i++ {
		c.shards[i] = NewLRUCache(perShard)
	}
	return c
}

func fnv32(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (c *ShardedLRU) getShard(key string) *LRUCache {
	return c.shards[fnv32(key)%shardCount]
}

func (c *ShardedLRU) Get(key string) (User, bool) {
	return c.getShard(key).Get(key)
}

func (c *ShardedLRU) Put(key string, value User) {
	c.getShard(key).Put(key, value)
}

var (
	db              *sql.DB
	cache           = NewShardedLRU(1000)
	stmtInsertUser  *sql.Stmt
	stmtGetUserByID *sql.Stmt
)

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "users.db")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, name TEXT, avatar BLOB)`)
	if err != nil {
		log.Fatal(err)
	}
	stmtGetUserByID, err = db.Prepare("SELECT id, name, avatar FROM users WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	stmtInsertUser, err = db.Prepare("INSERT OR REPLACE INTO users(id, name, avatar) VALUES(?, ?, ?)")
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
	if u.ID == "" || u.Name == "" || len(u.Avatar) == 0 {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if len(u.Avatar) > 20*1024 {
		http.Error(w, "avatar so large", http.StatusBadRequest)
		return
	}
	_, err := stmtInsertUser.Exec(u.ID, u.Name, u.Avatar)

	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	cache.Put(u.ID, u)
	w.WriteHeader(http.StatusCreated)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/user/")
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
	err := stmtGetUserByID.QueryRow(id).Scan(&u.ID, &u.Name, &u.Avatar)
	if err == sql.ErrNoRows {
		http.Error(w, "id not exist", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	cache.Put(id, u)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createUser(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getUser(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	return mux
}

//	func timingMiddleware(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			start := time.Now()
//
//			next.ServeHTTP(w, r)
//
//			log.Printf("%s %s took %v mc", r.Method, r.URL.Path, time.Since(start).Microseconds())
//		})
//	}

func preloadUsers() {
	rows, err := db.Query("SELECT id, name, avatar FROM users")
	if err != nil {
		log.Println("[preload] query error:", err)
		return
	}
	defer rows.Close()

	var count int

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Avatar); err != nil {
			continue
		}
		cache.Put(u.ID, u)
		count++
	}
	log.Printf("preload %d users into cache\n", count)
}
func seedFakeUsers(n int) {
	log.Printf("inserting %d fake users\n", n)

	for i := 1; i <= n; i++ {
		id := fmt.Sprintf("%d", i)
		name := fmt.Sprintf("user_%d", i)

		avatar := make([]byte, 512)
		for j := range avatar {
			avatar[j] = byte(i % 256)
		}

		_, err := stmtInsertUser.Exec(id, name, avatar)
		if err != nil {
			log.Println("[seed] insert error:", err)
			continue
		}
	}

	log.Println("[seed] done")
}
func main() {
	initDB()
	seedFakeUsers(1000)
	preloadUsers()
	defer db.Close()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router(),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	log.Println("Listening on " + srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
