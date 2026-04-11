package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const shardCount = 128

type User struct {
	ID     string
	Name   string
	Avatar []byte
}

type shard struct {
	mu   sync.RWMutex
	m    map[string]User
	keys []string
	max  int
}

type ShardedCache struct {
	shards [shardCount]*shard
}

func NewCache(maxPerShard int) *ShardedCache {
	c := &ShardedCache{}

	for i := 0; i < shardCount; i++ {
		c.shards[i] = &shard{
			m:    make(map[string]User),
			max:  maxPerShard,
			keys: make([]string, 0, maxPerShard),
		}
	}

	return c
}

func fnv32(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func (c *ShardedCache) getShard(key string) *shard {
	return c.shards[fnv32(key)%shardCount]
}

func (c *ShardedCache) Get(key string) (User, bool) {
	s := c.getShard(key)

	s.mu.RLock()
	u, ok := s.m[key]
	s.mu.RUnlock()

	return u, ok
}

func (c *ShardedCache) Put(key string, u User) {
	s := c.getShard(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[key]; ok {
		s.m[key] = u
		return
	}

	s.m[key] = u
	s.keys = append(s.keys, key)

	if len(s.m) > s.max {
		oldKey := s.keys[0]
		s.keys = s.keys[1:]
		delete(s.m, oldKey)
	}
}

var (
	db              *sql.DB
	cache           = NewCache(1000)
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
	return timingMiddleware(mux)
}

func timingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s took %v mc", r.Method, r.URL.Path, time.Since(start).Microseconds())
	})
}

func main() {
	initDB()

	srv := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      router(),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	log.Println("Listening on " + srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
