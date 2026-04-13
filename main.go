package main

import (
	"context"
	"database/sql"
	"hash/fnv"
	"log"
	"net"
	"sync"

	"user-service/user-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	_ "modernc.org/sqlite"
)

const shardCount = 100

type User struct {
	ID     string
	Name   string
	Avatar []byte
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

	s := &ShardedLRU{}
	for i := 0; i < shardCount; i++ {
		s.shards[i] = NewLRUCache(perShard)
	}
	return s
}

func fnv32(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (s *ShardedLRU) getShard(key string) *LRUCache {
	return s.shards[fnv32(key)%shardCount]
}

func (s *ShardedLRU) Get(key string) (User, bool) {
	return s.getShard(key).Get(key)
}

func (s *ShardedLRU) Put(key string, value User) {
	s.getShard(key).Put(key, value)
}

var (
	db              *sql.DB
	cache           = NewShardedLRU(5000)
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

type GRPCServer struct {
	proto.UnimplementedUserServiceServer
}

func (s *GRPCServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	if req.User == nil || req.User.Id == "" || req.User.Name == "" || len(req.User.Avatar) == 0 {
		return nil, status.Error(codes.InvalidArgument, "id, name and avatar are required")
	}

	if len(req.User.Avatar) > 20*1024 {
		return nil, status.Error(codes.InvalidArgument, "avatar too large (max 20KB)")
	}

	internalUser := User{
		ID:     req.User.Id,
		Name:   req.User.Name,
		Avatar: req.User.Avatar,
	}

	_, err := stmtInsertUser.Exec(internalUser.ID, internalUser.Name, internalUser.Avatar)
	if err != nil {
		log.Printf("db error: %v", err)
		return nil, status.Error(codes.Internal, "database error")
	}

	cache.Put(internalUser.ID, internalUser)

	return &proto.CreateUserResponse{
		Ok:      true,
		Message: "user created successfully",
	}, nil
}

func (s *GRPCServer) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if u, ok := cache.Get(req.Id); ok {
		return &proto.GetUserResponse{
			User: &proto.ProtoUser{
				Id:     u.ID,
				Name:   u.Name,
				Avatar: u.Avatar,
			},
			Found: true,
		}, nil
	}

	var internalUser User
	err := stmtGetUserByID.QueryRow(req.Id).Scan(&internalUser.ID, &internalUser.Name, &internalUser.Avatar)
	if err == sql.ErrNoRows {
		return &proto.GetUserResponse{
			Found: false,
		}, nil
	}
	if err != nil {
		log.Printf("db error: %v", err)
		return nil, status.Error(codes.Internal, "database error")
	}

	cache.Put(req.Id, internalUser)

	return &proto.GetUserResponse{
		User: &proto.ProtoUser{
			Id:     internalUser.ID,
			Name:   internalUser.Name,
			Avatar: internalUser.Avatar,
		},
		Found: true,
	}, nil
}

func main() {
	initDB()

	empty, err := isUsersTableEmpty(db)
	if err != nil {
		log.Fatal(err)
	}
	if empty {
		seedFakeUsers(db, stmtInsertUser, 1000)
	}

	preloadUsers()
	defer db.Close()

	grpcServer := grpc.NewServer()
	//reflection.Register(grpcServer)
	proto.RegisterUserServiceServer(grpcServer, &GRPCServer{})

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("gRPC server listening on :8080")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
