# Go语言中间件与工具集成

## 前言

本文档将全面介绍Go语言的中间件与工具集成，帮助读者更好地理解和使用Go语言进行中间件开发和工具集成。通过学习本文档，你将掌握Go语言中间件与工具集成的核心概念和常用技巧，包括HTTP中间件、日志中间件、认证中间件、数据库集成、缓存集成等。

## 第一章：HTTP中间件

### 1.1 中间件基础

中间件是一种在HTTP请求处理链中插入处理逻辑的模式。

**基本中间件**

```go
package main

import (
    "fmt"
    "net/http"
    "time"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Printf("[%s] %s %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    handler := loggingMiddleware(mux)
    
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", handler)
}
```

### 1.2 日志中间件

日志中间件用于记录HTTP请求和响应。

**请求日志**

```go
package main

import (
    "log"
    "net/http"
    "time"
)

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(rw, r)
        
        duration := time.Since(start)
        log.Printf("%s %s %d %v", r.Method, r.URL.Path, rw.status, duration)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    handler := loggingMiddleware(mux)
    
    log.Println("Server starting on :8080")
    http.ListenAndServe(":8080", handler)
}
```

### 1.3 认证中间件

认证中间件用于验证用户身份。

**API Key认证**

```go
package main

import (
    "net/http"
)

type APIKeyMiddleware struct {
    apiKeys map[string]bool
}

func NewAPIKeyMiddleware(apiKeys []string) *APIKeyMiddleware {
    keys := make(map[string]bool)
    for _, key := range apiKeys {
        keys[key] = true
    }
    return &APIKeyMiddleware{apiKeys: keys}
}

func (am *APIKeyMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            http.Error(w, "API Key required", http.StatusUnauthorized)
            return
        }

        if !am.apiKeys[apiKey] {
            http.Error(w, "Invalid API Key", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
    auth := NewAPIKeyMiddleware([]string{"key1", "key2", "key3"})
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Protected data"))
    })

    handler := auth.Middleware(mux)
    
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", handler)
}
```

### 1.4 CORS中间件

CORS中间件用于处理跨域请求。

**基本CORS**

```go
package main

import (
    "net/http"
)

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Data"))
    })

    handler := corsMiddleware(mux)
    
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", handler)
}
```

### 1.5 小结

本章介绍了Go语言的HTTP中间件，包括中间件基础、日志中间件、认证中间件、CORS中间件等。我们学习了Go语言如何创建中间件、构建中间件链、实现常见的中间件功能等。掌握好Go语言的中间件对于编写可维护的HTTP服务非常重要。

下一章我们将学习Go语言的数据库集成，包括SQL数据库、ORM、连接池等，以及它们的使用方法和注意事项。

## 第二章：数据库集成

### 2.1 SQL数据库

Go语言提供了database/sql包来操作SQL数据库。

**基本数据库操作**

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    _ "github.com/lib/pq"
)

type User struct {
    ID       int
    Name     string
    Email    string
    Age      int
    CreateAt time.Time
}

func main() {
    db, err := sql.Open("postgres", "user=postgres dbname=test sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }
    
    user := User{
        Name:  "John Doe",
        Email: "john@example.com",
        Age:   30,
    }
    
    err = createUser(db, &user)
    if err != nil {
        log.Fatal(err)
    }
    
    found, err := getUser(db, user.ID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found user: %+v\n", found)
}

func createUser(db *sql.DB, user *User) error {
    query := `INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id, created_at`
    
    err := db.QueryRow(query, user.Name, user.Email, user.Age).Scan(&user.ID, &user.CreateAt)
    if err != nil {
        return err
    }
    
    return nil
}

func getUser(db *sql.DB, id int) (*User, error) {
    query := `SELECT id, name, email, age, created_at FROM users WHERE id = $1`
    
    var user User
    err := db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.CreateAt)
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

**批量操作**

```go
func batchInsertUsers(db *sql.DB, users []User) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    stmt, err := tx.Prepare(`INSERT INTO users (name, email, age) VALUES ($1, $2, $3)`)
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, user := range users {
        _, err = stmt.Exec(user.Name, user.Email, user.Age)
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}

func listUsers(db *sql.DB) ([]User, error) {
    query := `SELECT id, name, email, age, created_at FROM users ORDER BY id`
    
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age, &user.CreateAt)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    
    return users, nil
}
```

### 2.2 ORM框架

GORM是一个流行的Go语言ORM框架。

**GORM基础**

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Product struct {
    ID        uint
    Name      string
    Price     float64
    Stock     int
    CreatedAt time.Time
    UpdatedAt time.Time
}

func main() {
    dsn := "host=localhost user=postgres dbname=test sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    err = db.AutoMigrate(&Product{})
    if err != nil {
        log.Fatal(err)
    }
    
    product := Product{
        Name:  "Laptop",
        Price: 999.99,
        Stock: 100,
    }
    
    err = createProduct(db, &product)
    if err != nil {
        log.Fatal(err)
    }
    
    found, err := getProduct(db, product.ID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found product: %+v\n", found)
}

func createProduct(db *gorm.DB, product *Product) error {
    return db.Create(product).Error
}

func getProduct(db *gorm.DB, id uint) (*Product, error) {
    var product Product
    err := db.First(&product, id).Error
    return &product, err
}

func updateProduct(db *gorm.DB, product *Product) error {
    return db.Save(product).Error
}

func deleteProduct(db *gorm.DB, id uint) error {
    return db.Delete(&Product{}, id).Error
}
```

**关联关系**

```go
type Order struct {
    ID        uint
    UserID    uint
    Total     float64
    Status    string
    User      User      `gorm:"foreignKey:UserID"`
    Items     []OrderItem `gorm:"foreignKey:OrderID"`
    CreatedAt time.Time
}

type OrderItem struct {
    ID        uint
    OrderID   uint
    ProductID uint
    Quantity  int
    Price     float64
    Order     Order   `gorm:"foreignKey:OrderID"`
    Product   Product `gorm:"foreignKey:ProductID"`
}

func createOrder(db *gorm.DB, order *Order) error {
    return db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(order).Error; err != nil {
            return err
        }
        
        for i := range order.Items {
            order.Items[i].OrderID = order.ID
            if err := tx.Create(&order.Items[i]).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
}

func getOrderWithItems(db *gorm.DB, id uint) (*Order, error) {
    var order Order
    err := db.Preload("Items").Preload("Items.Product").First(&order, id).Error
    return &order, err
}
```

### 2.3 连接池

连接池管理数据库连接，提高性能。

**配置连接池**

```go
func setupDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(5 * time.Minute)
    
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    return db, nil
}
```

**连接池监控**

```go
type DBStats struct {
    MaxOpenConnections int
    OpenConnections   int
    InUse             int
    Idle              int
    WaitCount         int64
    WaitDuration      time.Duration
    MaxIdleClosed     int64
    MaxLifetimeClosed int64
}

func getDBStats(db *sql.DB) DBStats {
    stats := db.Stats()
    return DBStats{
        MaxOpenConnections: stats.MaxOpenConnections,
        OpenConnections:    stats.OpenConnections,
        InUse:             stats.InUse,
        Idle:              stats.Idle,
        WaitCount:         stats.WaitCount,
        WaitDuration:      stats.WaitDuration,
        MaxIdleClosed:     stats.MaxIdleClosed,
        MaxLifetimeClosed: stats.MaxLifetimeClosed,
    }
}
```

### 2.4 事务处理

事务确保数据一致性。

**基本事务**

```go
func transferMoney(db *sql.DB, fromID, toID int, amount float64) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    _, err = tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromID)
    if err != nil {
        return err
    }
    
    _, err = tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toID)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

**GORM事务**

```go
func transferMoneyGORM(db *gorm.DB, fromID, toID uint, amount float64) error {
    return db.Transaction(func(tx *gorm.DB) error {
        var fromAccount, toAccount Account
        
        if err := tx.First(&fromAccount, fromID).Error; err != nil {
            return err
        }
        
        if err := tx.First(&toAccount, toID).Error; err != nil {
            return err
        }
        
        if fromAccount.Balance < amount {
            return errors.New("insufficient balance")
        }
        
        fromAccount.Balance -= amount
        toAccount.Balance += amount
        
        if err := tx.Save(&fromAccount).Error; err != nil {
            return err
        }
        
        if err := tx.Save(&toAccount).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### 2.5 小结

本章介绍了Go语言的数据库集成，包括SQL数据库、ORM框架、连接池、事务处理等。我们学习了Go语言如何使用database/sql包操作数据库、使用GORM进行ORM操作、配置连接池、处理事务等。掌握好Go语言的数据库集成对于构建数据驱动的应用程序非常重要。

下一章我们将学习Go语言的缓存集成，包括Redis、Memcached、本地缓存等，以及它们的使用方法和注意事项。

## 第三章：缓存集成

### 3.1 Redis缓存

Redis是一个高性能的键值存储系统，常用于缓存。

**基本Redis操作**

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: "",
        DB:       0,
    })
    
    return &RedisCache{client: rdb}
}

func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    return rc.client.Set(ctx, key, value, expiration).Err()
}

func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
    return rc.client.Get(ctx, key).Result()
}

func (rc *RedisCache) Delete(ctx context.Context, keys ...string) error {
    return rc.client.Del(ctx, keys...).Err()
}

func (rc *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
    return rc.client.Exists(ctx, keys...).Result()
}

func (rc *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
    return rc.client.SetNX(ctx, key, value, expiration).Result()
}

func main() {
    cache := NewRedisCache("localhost:6379")
    ctx := context.Background()
    
    err := cache.Set(ctx, "user:1", "John Doe", time.Hour)
    if err != nil {
        log.Fatal(err)
    }
    
    value, err := cache.Get(ctx, "user:1")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Cached value: %s\n", value)
}
```

**Redis哈希操作**

```go
func (rc *RedisCache) HSet(ctx context.Context, key, field string, value interface{}) error {
    return rc.client.HSet(ctx, key, field, value).Err()
}

func (rc *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
    return rc.client.HGet(ctx, key, field).Result()
}

func (rc *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
    return rc.client.HGetAll(ctx, key).Result()
}

func (rc *RedisCache) HDel(ctx context.Context, key string, fields ...string) error {
    return rc.client.HDel(ctx, key, fields...).Err()
}

func (rc *RedisCache) HExists(ctx context.Context, key, field string) (bool, error) {
    return rc.client.HExists(ctx, key, field).Result()
}
```

**Redis列表操作**

```go
func (rc *RedisCache) LPush(ctx context.Context, key string, values ...interface{}) error {
    return rc.client.LPush(ctx, key, values...).Err()
}

func (rc *RedisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
    return rc.client.RPush(ctx, key, values...).Err()
}

func (rc *RedisCache) LPop(ctx context.Context, key string) (string, error) {
    return rc.client.LPop(ctx, key).Result()
}

func (rc *RedisCache) RPop(ctx context.Context, key string) (string, error) {
    return rc.client.RPop(ctx, key).Result()
}

func (rc *RedisCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
    return rc.client.LRange(ctx, key, start, stop).Result()
}

func (rc *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
    return rc.client.LLen(ctx, key).Result()
}
```

### 3.2 Memcached缓存

Memcached是一个高性能的分布式内存对象缓存系统。

**基本Memcached操作**

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/bradfitz/gomemcache/memcache"
)

type MemcachedCache struct {
    client *memcache.Client
}

func NewMemcachedCache(servers []string) *MemcachedCache {
    return &MemcachedCache{
        client: memcache.New(servers...),
    }
}

func (mc *MemcachedCache) Set(key string, value []byte, expiration int32) error {
    item := &memcache.Item{
        Key:        key,
        Value:      value,
        Expiration: expiration,
    }
    return mc.client.Set(item)
}

func (mc *MemcachedCache) Get(key string) ([]byte, error) {
    item, err := mc.client.Get(key)
    if err != nil {
        return nil, err
    }
    return item.Value, nil
}

func (mc *MemcachedCache) Delete(key string) error {
    return mc.client.Delete(key)
}

func (mc *MemcachedCache) DeleteAll() error {
    return mc.client.DeleteAll()
}

func main() {
    cache := NewMemcachedCache([]string{"localhost:11211"})
    
    err := cache.Set("user:1", []byte("John Doe"), 3600)
    if err != nil {
        log.Fatal(err)
    }
    
    value, err := cache.Get("user:1")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Cached value: %s\n", string(value))
}
```

### 3.3 本地缓存

本地缓存使用内存存储数据，速度最快。

**简单本地缓存**

```go
package main

import (
    "sync"
    "time"
)

type LocalCache struct {
    items map[string]*cacheItem
    mu    sync.RWMutex
}

type cacheItem struct {
    value      interface{}
    expiration int64
}

func NewLocalCache() *LocalCache {
    lc := &LocalCache{
        items: make(map[string]*cacheItem),
    }
    go lc.cleanup()
    return lc
}

func (lc *LocalCache) Set(key string, value interface{}, ttl time.Duration) {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    
    var expiration int64
    if ttl > 0 {
        expiration = time.Now().Add(ttl).UnixNano()
    }
    
    lc.items[key] = &cacheItem{
        value:      value,
        expiration: expiration,
    }
}

func (lc *LocalCache) Get(key string) (interface{}, bool) {
    lc.mu.RLock()
    defer lc.mu.RUnlock()
    
    item, ok := lc.items[key]
    if !ok {
        return nil, false
    }
    
    if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
        return nil, false
    }
    
    return item.value, true
}

func (lc *LocalCache) Delete(key string) {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    
    delete(lc.items, key)
}

func (lc *LocalCache) cleanup() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        lc.mu.Lock()
        now := time.Now().UnixNano()
        for key, item := range lc.items {
            if item.expiration > 0 && now > item.expiration {
                delete(lc.items, key)
            }
        }
        lc.mu.Unlock()
    }
}
```

**LRU缓存**

```go
package main

import (
    "container/list"
    "sync"
)

type LRUCache struct {
    capacity int
    items    map[string]*list.Element
    list     *list.List
    mu       sync.Mutex
}

type cacheEntry struct {
    key   string
    value interface{}
}

func NewLRUCache(capacity int) *LRUCache {
    return &LRUCache{
        capacity: capacity,
        items:    make(map[string]*list.Element),
        list:     list.New(),
    }
}

func (lc *LRUCache) Get(key string) (interface{}, bool) {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    
    if elem, ok := lc.items[key]; ok {
        lc.list.MoveToFront(elem)
        return elem.Value.(*cacheEntry).value, true
    }
    
    return nil, false
}

func (lc *LRUCache) Set(key string, value interface{}) {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    
    if elem, ok := lc.items[key]; ok {
        lc.list.MoveToFront(elem)
        elem.Value.(*cacheEntry).value = value
        return
    }
    
    if lc.list.Len() >= lc.capacity {
        oldest := lc.list.Back()
        if oldest != nil {
            lc.list.Remove(oldest)
            delete(lc.items, oldest.Value.(*cacheEntry).key)
        }
    }
    
    elem := lc.list.PushFront(&cacheEntry{key: key, value: value})
    lc.items[key] = elem
}

func (lc *LRUCache) Delete(key string) {
    lc.mu.Lock()
    defer lc.mu.Unlock()
    
    if elem, ok := lc.items[key]; ok {
        lc.list.Remove(elem)
        delete(lc.items, key)
    }
}
```

### 3.4 缓存模式

**缓存穿透**

```go
func (s *UserService) GetUserWithCache(id string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", id)
    
    if cached, ok := s.cache.Get(cacheKey); ok {
        if cached == nil {
            return nil, nil
        }
        return cached.(*User), nil
    }
    
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    if user == nil {
        s.cache.Set(cacheKey, nil, 5*time.Minute)
        return nil, nil
    }
    
    s.cache.Set(cacheKey, user, time.Hour)
    return user, nil
}
```

**缓存击穿**

```go
func (s *UserService) GetUserWithLock(id string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", id)
    
    if cached, ok := s.cache.Get(cacheKey); ok {
        return cached.(*User), nil
    }
    
    lockKey := fmt.Sprintf("lock:%s", id)
    locked, err := s.cache.SetNX(lockKey, "1", 10*time.Second)
    if err != nil {
        return nil, err
    }
    
    if locked {
        defer s.cache.Delete(lockKey)
        
        user, err := s.repo.FindByID(id)
        if err != nil {
            return nil, err
        }
        
        s.cache.Set(cacheKey, user, time.Hour)
        return user, nil
    }
    
    time.Sleep(100 * time.Millisecond)
    return s.GetUserWithLock(id)
}
```

**缓存雪崩**

```go
func (s *UserService) GetUserWithRandomTTL(id string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", id)
    
    if cached, ok := s.cache.Get(cacheKey); ok {
        return cached.(*User), nil
    }
    
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    ttl := time.Hour + time.Duration(rand.Intn(300))*time.Second
    s.cache.Set(cacheKey, user, ttl)
    return user, nil
}
```

### 3.5 小结

本章介绍了Go语言的缓存集成，包括Redis缓存、Memcached缓存、本地缓存、缓存模式等。我们学习了Go语言如何使用Redis进行缓存、使用Memcached进行缓存、实现本地缓存、处理缓存穿透、缓存击穿、缓存雪崩等问题。掌握好Go语言的缓存集成对于提高应用程序性能非常重要。

下一章我们将学习Go语言的更多中间件，包括限流、压缩、超时、重试等，以及它们的使用方法和注意事项。

## 第四章：更多中间件

### 4.1 限流中间件

限流中间件用于控制请求速率，防止系统过载。

**令牌桶限流**

```go
package main

import (
    "net/http"
    "sync"
    "time"
)

type TokenBucket struct {
    capacity int
    tokens   int
    rate     int
    mu       sync.Mutex
    lastTime time.Time
}

func NewTokenBucket(capacity, rate int) *TokenBucket {
    return &TokenBucket{
        capacity: capacity,
        tokens:   capacity,
        rate:     rate,
        lastTime: time.Now(),
    }
}

func (tb *TokenBucket) Allow() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()
    
    now := time.Now()
    elapsed := now.Sub(tb.lastTime).Seconds()
    
    tb.tokens += int(elapsed * float64(tb.rate))
    if tb.tokens > tb.capacity {
        tb.tokens = tb.capacity
    }
    
    tb.lastTime = now
    
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }
    
    return false
}

func rateLimitMiddleware(tb *TokenBucket) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !tb.Allow() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func main() {
    tb := NewTokenBucket(100, 10)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Data"))
    })

    handler := rateLimitMiddleware(tb)(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

**滑动窗口限流**

```go
type SlidingWindow struct {
    windowSize time.Duration
    maxCount   int
    requests   []time.Time
    mu         sync.Mutex
}

func NewSlidingWindow(windowSize time.Duration, maxCount int) *SlidingWindow {
    return &SlidingWindow{
        windowSize: windowSize,
        maxCount:   maxCount,
        requests:   make([]time.Time, 0),
    }
}

func (sw *SlidingWindow) Allow() bool {
    sw.mu.Lock()
    defer sw.mu.Unlock()
    
    now := time.Now()
    cutoff := now.Add(-sw.windowSize)
    
    for i, t := range sw.requests {
        if t.After(cutoff) {
            sw.requests = sw.requests[i:]
            break
        }
    }
    
    if len(sw.requests) >= sw.maxCount {
        return false
    }
    
    sw.requests = append(sw.requests, now)
    return true
}
```

### 4.2 压缩中间件

压缩中间件用于压缩响应数据，减少网络传输。

**Gzip压缩**

```go
package main

import (
    "compress/gzip"
    "net/http"
    "strings"
)

type gzipResponseWriter struct {
    http.ResponseWriter
    writer *gzip.Writer
}

func (gzw *gzipResponseWriter) Write(b []byte) (int, error) {
    return gzw.writer.Write(b)
}

func (gzw *gzipResponseWriter) Close() error {
    return gzw.writer.Close()
}

func gzipMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }
        
        w.Header().Set("Content-Encoding", "gzip")
        
        gz := gzip.NewWriter(w)
        defer gz.Close()
        
        gzw := &gzipResponseWriter{ResponseWriter: w, writer: gz}
        next.ServeHTTP(gzw, r)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"message": "Hello, World!"}`))
    })

    handler := gzipMiddleware(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### 4.3 超时中间件

超时中间件用于设置请求超时时间。

**请求超时**

```go
package main

import (
    "context"
    "net/http"
    "time"
)

func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            done := make(chan struct{})
            
            go func() {
                next.ServeHTTP(w, r.WithContext(ctx))
                close(done)
            }()
            
            select {
            case <-done:
                return
            case <-ctx.Done():
                http.Error(w, "Request timeout", http.StatusRequestTimeout)
            }
        })
    }
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(2 * time.Second)
        w.Write([]byte("Data"))
    })

    handler := timeoutMiddleware(1 * time.Second)(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### 4.4 重试中间件

重试中间件用于在请求失败时自动重试。

**指数退避重试**

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

type RetryConfig struct {
    MaxRetries int
    InitialDelay time.Duration
    MaxDelay     time.Duration
}

func exponentialBackoff(attempt int, config RetryConfig) time.Duration {
    delay := config.InitialDelay * time.Duration(1<<uint(attempt))
    if delay > config.MaxDelay {
        delay = config.MaxDelay
    }
    return delay
}

func retryMiddleware(config RetryConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var lastErr error
            
            for attempt := 0; attempt <= config.MaxRetries; attempt++ {
                if attempt > 0 {
                    delay := exponentialBackoff(attempt-1, config)
                    time.Sleep(delay)
                }
                
                rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
                next.ServeHTTP(rw, r)
                
                if rw.statusCode < 500 {
                    return
                }
                
                lastErr = fmt.Errorf("request failed with status %d", rw.statusCode)
            }
            
            http.Error(w, lastErr.Error(), http.StatusServiceUnavailable)
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func main() {
    config := RetryConfig{
        MaxRetries:   3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay:     5 * time.Second,
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Data"))
    })

    handler := retryMiddleware(config)(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### 4.5 熔断中间件

熔断中间件用于在服务不可用时快速失败。

**熔断器实现**

```go
package main

import (
    "net/http"
    "sync"
    "time"
)

type CircuitState int

const (
    StateClosed CircuitState = iota
    StateOpen
    StateHalfOpen
)

type CircuitBreaker struct {
    mu                sync.Mutex
    state             CircuitState
    failureCount      int
    failureThreshold  int
    successThreshold  int
    timeout           time.Duration
    lastFailureTime   time.Time
}

func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:            StateClosed,
        failureThreshold: failureThreshold,
        successThreshold: successThreshold,
        timeout:          timeout,
    }
}

func (cb *CircuitBreaker) Allow() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    switch cb.state {
    case StateClosed:
        return true
    case StateOpen:
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = StateHalfOpen
            return true
        }
        return false
    case StateHalfOpen:
        return true
    default:
        return false
    }
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.failureCount = 0
    
    if cb.state == StateHalfOpen {
        cb.state = StateClosed
    }
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.failureCount++
    cb.lastFailureTime = time.Now()
    
    if cb.failureCount >= cb.failureThreshold {
        cb.state = StateOpen
    }
}

func circuitBreakerMiddleware(cb *CircuitBreaker) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !cb.Allow() {
                http.Error(w, "Circuit breaker is open", http.StatusServiceUnavailable)
                return
            }
            
            rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            next.ServeHTTP(rw, r)
            
            if rw.statusCode >= 500 {
                cb.RecordFailure()
            } else {
                cb.RecordSuccess()
            }
        })
    }
}

func main() {
    cb := NewCircuitBreaker(5, 3, 30*time.Second)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Data"))
    })

    handler := circuitBreakerMiddleware(cb)(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### 4.6 小结

本章介绍了Go语言的更多中间件，包括限流中间件、压缩中间件、超时中间件、重试中间件、熔断中间件等。我们学习了Go语言如何实现令牌桶限流、滑动窗口限流、Gzip压缩、请求超时、指数退避重试、熔断器等。掌握好Go语言的这些中间件对于构建高可用、高性能的HTTP服务非常重要。

下一章我们将学习Go语言的消息队列集成，包括RabbitMQ、Kafka等，以及它们的使用方法和注意事项。

## 第五章：消息队列集成

### 5.1 RabbitMQ

RabbitMQ是一个开源的消息代理，支持多种消息协议。

**基本RabbitMQ操作**

```go
package main

import (
    "log"
    
    amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
    conn *amqp.Connection
    ch   *amqp.Channel
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }
    
    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }
    
    return &RabbitMQClient{conn: conn, ch: ch}, nil
}

func (c *RabbitMQClient) Close() error {
    if c.ch != nil {
        c.ch.Close()
    }
    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}

func (c *RabbitMQClient) DeclareQueue(name string) (amqp.Queue, error) {
    return c.ch.QueueDeclare(
        name,
        true,
        false,
        false,
        false,
        nil,
    )
}

func (c *RabbitMQClient) Publish(queue string, body []byte) error {
    return c.ch.Publish(
        "",
        queue,
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        body,
        },
    )
}

func (c *RabbitMQClient) Consume(queue string) (<-chan amqp.Delivery, error) {
    return c.ch.Consume(
        queue,
        "",
        true,
        false,
        false,
        false,
        nil,
    )
}

func main() {
    client, err := NewRabbitMQClient("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    queue, err := client.DeclareQueue("hello")
    if err != nil {
        log.Fatal(err)
    }
    
    err = client.Publish(queue.Name, []byte("Hello, RabbitMQ!"))
    if err != nil {
        log.Fatal(err)
    }
    
    msgs, err := client.Consume(queue.Name)
    if err != nil {
        log.Fatal(err)
    }
    
    for msg := range msgs {
        log.Printf("Received: %s", msg.Body)
    }
}
```

**工作队列**

```go
func (c *RabbitMQClient) PublishTask(queue string, task []byte) error {
    return c.ch.Publish(
        "",
        queue,
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        task,
        DeliveryMode: amqp.Persistent,
        },
    )
}

func (c *RabbitMQClient) SetQos(prefetchCount int) error {
    return c.ch.Qos(
        prefetchCount,
        0,
        false,
    )
}

func worker() {
    client, err := NewRabbitMQClient("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    queue, err := client.DeclareQueue("task_queue")
    if err != nil {
        log.Fatal(err)
    }
    
    err = client.SetQos(1)
    if err != nil {
        log.Fatal(err)
    }
    
    msgs, err := client.Consume(queue.Name)
    if err != nil {
        log.Fatal(err)
    }
    
    for msg := range msgs {
        log.Printf("Processing task: %s", msg.Body)
        
        time.Sleep(2 * time.Second)
        
        log.Printf("Task completed: %s", msg.Body)
    }
}
```

**发布订阅**

```go
func (c *RabbitMQClient) DeclareExchange(name, kind string) error {
    return c.ch.ExchangeDeclare(
        name,
        kind,
        true,
        false,
        false,
        false,
        nil,
    )
}

func (c *RabbitMQClient) PublishToExchange(exchange string, body []byte) error {
    return c.ch.Publish(
        exchange,
        "",
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        body,
        },
    )
}

func (c *RabbitMQClient) BindQueue(queue, exchange string) error {
    return c.ch.QueueBind(
        queue,
        "",
        exchange,
        false,
        nil,
    )
}

func publisher() {
    client, err := NewRabbitMQClient("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    err = client.DeclareExchange("logs", "fanout")
    if err != nil {
        log.Fatal(err)
    }
    
    err = client.PublishToExchange("logs", []byte("Hello, World!"))
    if err != nil {
        log.Fatal(err)
    }
}

func subscriber() {
    client, err := NewRabbitMQClient("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    err = client.DeclareExchange("logs", "fanout")
    if err != nil {
        log.Fatal(err)
    }
    
    queue, err := client.DeclareQueue("")
    if err != nil {
        log.Fatal(err)
    }
    
    err = client.BindQueue(queue.Name, "logs")
    if err != nil {
        log.Fatal(err)
    }
    
    msgs, err := client.Consume(queue.Name)
    if err != nil {
        log.Fatal(err)
    }
    
    for msg := range msgs {
        log.Printf("Received: %s", msg.Body)
    }
}
```

### 5.2 Kafka

Kafka是一个分布式流处理平台，用于处理实时数据流。

**基本Kafka操作**

```go
package main

import (
    "context"
    "log"
    
    "github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
    writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
    return &KafkaProducer{
        writer: &kafka.Writer{
            Addr:     kafka.TCP(brokers...),
            Topic:    topic,
            Balancer: &kafka.LeastBytes{},
        },
    }
}

func (p *KafkaProducer) Produce(ctx context.Context, key, value []byte) error {
    return p.writer.WriteMessages(ctx, kafka.Message{
        Key:   key,
        Value: value,
    })
}

func (p *KafkaProducer) Close() error {
    return p.writer.Close()
}

type KafkaConsumer struct {
    reader *kafka.Reader
}

func NewKafkaConsumer(brokers []string, topic, groupID string) *KafkaConsumer {
    return &KafkaConsumer{
        reader: kafka.NewReader(kafka.ReaderConfig{
            Brokers:  brokers,
            Topic:    topic,
            GroupID:  groupID,
            MinBytes: 10e3,
            MaxBytes: 10e6,
        }),
    }
}

func (c *KafkaConsumer) Consume(ctx context.Context) (kafka.Message, error) {
    return c.reader.ReadMessage(ctx)
}

func (c *KafkaConsumer) Close() error {
    return c.reader.Close()
}

func main() {
    producer := NewKafkaProducer([]string{"localhost:9092"}, "test-topic")
    defer producer.Close()
    
    err := producer.Produce(context.Background(), []byte("key"), []byte("Hello, Kafka!"))
    if err != nil {
        log.Fatal(err)
    }
    
    consumer := NewKafkaConsumer([]string{"localhost:9092"}, "test-topic", "test-group")
    defer consumer.Close()
    
    for {
        msg, err := consumer.Consume(context.Background())
        if err != nil {
            log.Fatal(err)
        }
        
        log.Printf("Received: key=%s, value=%s", string(msg.Key), string(msg.Value))
    }
}
```

**Kafka分区**

```go
func (p *KafkaProducer) ProduceToPartition(ctx context.Context, partition int, key, value []byte) error {
    return p.writer.WriteMessages(ctx, kafka.Message{
        Partition: partition,
        Key:      key,
        Value:    value,
    })
}

func (c *KafkaConsumer) ConsumePartition(ctx context.Context, partition int) (kafka.Message, error) {
    return c.reader.ReadMessage(ctx)
}
```

### 5.3 消息模式

**发布订阅模式**

```go
type EventBus struct {
    subscribers map[string][]chan []byte
    mu          sync.RWMutex
}

func NewEventBus() *EventBus {
    return &EventBus{
        subscribers: make(map[string][]chan []byte),
    }
}

func (eb *EventBus) Subscribe(topic string) chan []byte {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    ch := make(chan []byte, 100)
    eb.subscribers[topic] = append(eb.subscribers[topic], ch)
    return ch
}

func (eb *EventBus) Publish(topic string, message []byte) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    
    for _, ch := range eb.subscribers[topic] {
        select {
        case ch <- message:
        default:
            log.Printf("Channel full, dropping message")
        }
    }
}

func (eb *EventBus) Unsubscribe(topic string, ch chan []byte) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    subscribers := eb.subscribers[topic]
    for i, subscriber := range subscribers {
        if subscriber == ch {
            eb.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
            break
        }
    }
}
```

**请求响应模式**

```go
type RPCClient struct {
    conn    *amqp.Connection
    ch      *amqp.Channel
    replyTo string
    chMap   map[string]chan []byte
    mu      sync.Mutex
}

func NewRPCClient(url string) (*RPCClient, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }
    
    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }
    
    queue, err := ch.QueueDeclare("", false, false, true, false, nil)
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }
    
    return &RPCClient{
        conn:    conn,
        ch:      ch,
        replyTo: queue.Name,
        chMap:   make(map[string]chan []byte),
    }, nil
}

func (c *RPCClient) Call(queue string, request []byte) ([]byte, error) {
    corrID := generateUUID()
    
    replyCh := make(chan []byte, 1)
    
    c.mu.Lock()
    c.chMap[corrID] = replyCh
    c.mu.Unlock()
    
    defer func() {
        c.mu.Lock()
        delete(c.chMap, corrID)
        c.mu.Unlock()
    }()
    
    err := c.ch.Publish(
        "",
        queue,
        false,
        false,
        amqp.Publishing{
            ContentType:   "text/plain",
            CorrelationId: corrID,
            ReplyTo:       c.replyTo,
            Body:          request,
        },
    )
    
    if err != nil {
        return nil, err
    }
    
    select {
    case response := <-replyCh:
        return response, nil
    case <-time.After(5 * time.Second):
        return nil, errors.New("timeout")
    }
}

func (c *RPCClient) StartListening() error {
    msgs, err := c.ch.Consume(c.replyTo, "", true, false, false, false, nil)
    if err != nil {
        return err
    }
    
    go func() {
        for msg := range msgs {
            c.mu.Lock()
            if ch, ok := c.chMap[msg.CorrelationId]; ok {
                ch <- msg.Body
            }
            c.mu.Unlock()
        }
    }()
    
    return nil
}
```

### 5.4 小结

本章介绍了Go语言的消息队列集成，包括RabbitMQ、Kafka、消息模式等。我们学习了Go语言如何使用RabbitMQ进行消息传递、使用Kafka进行流处理、实现发布订阅模式、请求响应模式等。掌握好Go语言的消息队列集成对于构建分布式系统非常重要。

下一章我们将学习Go语言的配置管理，包括环境变量、配置文件、配置中心等，以及它们的使用方法和注意事项。

## 第六章：配置管理

### 6.1 环境变量

环境变量是配置管理的基础方式。

**基本环境变量操作**

```go
package main

import (
    "fmt"
    "os"
    "strconv"
)

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func getEnvInt(key string, defaultValue int) int {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    
    intValue, err := strconv.Atoi(value)
    if err != nil {
        return defaultValue
    }
    
    return intValue
}

func getEnvBool(key string, defaultValue bool) bool {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    
    boolValue, err := strconv.ParseBool(value)
    if err != nil {
        return defaultValue
    }
    
    return boolValue
}

type Config struct {
    DatabaseURL string
    Port        int
    Debug       bool
}

func LoadConfig() *Config {
    return &Config{
        DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/db"),
        Port:        getEnvInt("PORT", 8080),
        Debug:       getEnvBool("DEBUG", false),
    }
}

func main() {
    config := LoadConfig()
    fmt.Printf("DatabaseURL: %s\n", config.DatabaseURL)
    fmt.Printf("Port: %d\n", config.Port)
    fmt.Printf("Debug: %v\n", config.Debug)
}
```

**使用Viper库**

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func LoadConfigWithViper() (*Config, error) {
    viper.SetDefault("database.url", "postgres://localhost:5432/db")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.debug", false)
    
    viper.AutomaticEnv()
    viper.SetEnvPrefix("APP")
    viper.BindEnv("database.url", "DATABASE_URL")
    viper.BindEnv("server.port", "PORT")
    viper.BindEnv("server.debug", "DEBUG")
    
    config := &Config{}
    if err := viper.Unmarshal(config); err != nil {
        return nil, err
    }
    
    return config, nil
}
```

### 6.2 配置文件

配置文件提供更灵活的配置管理方式。

**JSON配置文件**

```go
package main

import (
    "encoding/json"
    "os"
)

type Config struct {
    Database struct {
        URL string `json:"url"`
    } `json:"database"`
    Server struct {
        Port  int  `json:"port"`
        Debug bool `json:"debug"`
    } `json:"server"`
}

func LoadConfigFromFile(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func SaveConfigToFile(filename string, config *Config) error {
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, data, 0644)
}
```

**YAML配置文件**

```go
package main

import (
    "fmt"
    "os"
    
    "gopkg.in/yaml.v2"
)

type Config struct {
    Database struct {
        URL string `yaml:"url"`
    } `yaml:"database"`
    Server struct {
        Port  int  `yaml:"port"`
        Debug bool `yaml:"debug"`
    } `yaml:"server"`
}

func LoadConfigFromYAML(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func SaveConfigToYAML(filename string, config *Config) error {
    data, err := yaml.Marshal(config)
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, data, 0644)
}

func main() {
    config := &Config{
        Database: struct{ URL string }{URL: "postgres://localhost:5432/db"},
        Server: struct {
            Port  int
            Debug bool
        }{Port: 8080, Debug: false},
    }
    
    err := SaveConfigToYAML("config.yaml", config)
    if err != nil {
        fmt.Printf("Error saving config: %v\n", err)
        return
    }
    
    loaded, err := LoadConfigFromYAML("config.yaml")
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        return
    }
    
    fmt.Printf("Loaded config: %+v\n", loaded)
}
```

**TOML配置文件**

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/BurntSushi/toml"
)

type Config struct {
    Database struct {
        URL string `toml:"url"`
    } `toml:"database"`
    Server struct {
        Port  int  `toml:"port"`
        Debug bool `toml:"debug"`
    } `toml:"server"`
}

func LoadConfigFromTOML(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if _, err := toml.Decode(string(data), &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func SaveConfigToTOML(filename string, config *Config) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    return toml.NewEncoder(file).Encode(config)
}
```

### 6.3 配置中心

配置中心提供集中化的配置管理。

**Etcd配置中心**

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "go.etcd.io/etcd/client/v3"
)

type EtcdConfig struct {
    client *clientv3.Client
}

func NewEtcdConfig(endpoints []string) (*EtcdConfig, error) {
    client, err := clientv3.New(clientv3.Config{
        Endpoints:   endpoints,
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        return nil, err
    }
    
    return &EtcdConfig{client: client}, nil
}

func (ec *EtcdConfig) Get(ctx context.Context, key string) (string, error) {
    resp, err := ec.client.Get(ctx, key)
    if err != nil {
        return "", err
    }
    
    if len(resp.Kvs) == 0 {
        return "", fmt.Errorf("key not found")
    }
    
    return string(resp.Kvs[0].Value), nil
}

func (ec *EtcdConfig) Set(ctx context.Context, key, value string) error {
    _, err := ec.client.Put(ctx, key, value)
    return err
}

func (ec *EtcdConfig) Watch(ctx context.Context, key string) <-chan string {
    ch := make(chan string)
    
    watchChan := ec.client.Watch(ctx, key)
    
    go func() {
        for watchResp := range watchChan {
            for _, event := range watchResp.Events {
                ch <- string(event.Kv.Value)
            }
        }
        close(ch)
    }()
    
    return ch
}

func (ec *EtcdConfig) Close() error {
    return ec.client.Close()
}
```

**Consul配置中心**

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/hashicorp/consul/api"
)

type ConsulConfig struct {
    client *api.Client
}

func NewConsulConfig(addr string) (*ConsulConfig, error) {
    config := api.DefaultConfig()
    config.Address = addr
    
    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }
    
    return &ConsulConfig{client: client}, nil
}

func (cc *ConsulConfig) Get(key string) (string, error) {
    kv, _, err := cc.client.KV().Get(key, nil)
    if err != nil {
        return "", err
    }
    
    if kv == nil {
        return "", fmt.Errorf("key not found")
    }
    
    return string(kv.Value), nil
}

func (cc *ConsulConfig) Set(key, value string) error {
    kv := &api.KVPair{
        Key:   key,
        Value: []byte(value),
    }
    
    _, err := cc.client.KV().Put(kv, nil)
    return err
}

func (cc *ConsulConfig) Watch(key string) (<-chan string, error) {
    ch := make(chan string)
    
    opts := &api.QueryOptions{
        WaitIndex: 0,
    }
    
    go func() {
        for {
            kv, meta, err := cc.client.KV().Get(key, opts)
            if err != nil {
                close(ch)
                return
            }
            
            if kv != nil {
                ch <- string(kv.Value)
            }
            
            opts.WaitIndex = meta.LastIndex
        }
    }()
    
    return ch, nil
}
```

### 6.4 配置热更新

配置热更新允许在不重启应用的情况下更新配置。

**配置热更新实现**

```go
package main

import (
    "context"
    "log"
    "sync"
    "time"
    
    "github.com/fsnotify/fsnotify"
)

type ConfigManager struct {
    config  *Config
    mu      sync.RWMutex
    watchers []func(*Config)
}

func NewConfigManager(filename string) (*ConfigManager, error) {
    config, err := LoadConfigFromFile(filename)
    if err != nil {
        return nil, err
    }
    
    cm := &ConfigManager{
        config:   config,
        watchers: make([]func(*Config), 0),
    }
    
    go cm.watchConfig(filename)
    
    return cm, nil
}

func (cm *ConfigManager) GetConfig() *Config {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    return cm.config
}

func (cm *ConfigManager) WatchConfig(callback func(*Config)) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    
    cm.watchers = append(cm.watchers, callback)
}

func (cm *ConfigManager) watchConfig(filename string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()
    
    err = watcher.Add(filename)
    if err != nil {
        log.Fatal(err)
    }
    
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            
            if event.Op&fsnotify.Write == fsnotify.Write {
                config, err := LoadConfigFromFile(filename)
                if err != nil {
                    log.Printf("Error reloading config: %v", err)
                    continue
                }
                
                cm.mu.Lock()
                cm.config = config
                cm.mu.Unlock()
                
                for _, watcher := range cm.watchers {
                    watcher(config)
                }
                
                log.Println("Config reloaded")
            }
            
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Watcher error: %v", err)
        }
    }
}
```

### 6.5 小结

本章介绍了Go语言的配置管理，包括环境变量、配置文件、配置中心、配置热更新等。我们学习了Go语言如何使用环境变量、JSON/YAML/TOML配置文件、Etcd/Consul配置中心、实现配置热更新等。掌握好Go语言的配置管理对于构建可维护的应用程序非常重要。

下一章我们将学习Go语言的监控和追踪，包括Prometheus、Jaeger、日志聚合等，以及它们的使用方法和注意事项。

## 第七章：监控和追踪

### 7.1 Prometheus监控

Prometheus是一个开源的监控和警报工具包。

**Prometheus客户端**

```go
package main

import (
    "net/http"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    
    activeConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Number of active connections",
        },
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(activeConnections)
}

func prometheusMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        activeConnections.Inc()
        defer activeConnections.Dec()
        
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(rw, r)
        
        duration := time.Since(start).Seconds()
        
        httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, string(rw.statusCode)).Inc()
        httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Data"))
    })
    
    handler := prometheusMiddleware(mux)
    
    http.Handle("/metrics", promhttp.Handler())
    http.Handle("/", handler)
    
    http.ListenAndServe(":8080", nil)
}
```

**自定义指标**

```go
type Metrics struct {
    requestsTotal    *prometheus.CounterVec
    requestDuration *prometheus.HistogramVec
    errorsTotal     *prometheus.CounterVec
    activeUsers     prometheus.Gauge
}

func NewMetrics() *Metrics {
    return &Metrics{
        requestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "app_requests_total",
                Help: "Total number of requests",
            },
            []string{"endpoint", "method"},
        ),
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "app_request_duration_seconds",
                Help:    "Request duration in seconds",
                Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
            },
            []string{"endpoint"},
        ),
        errorsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "app_errors_total",
                Help: "Total number of errors",
            },
            []string{"type"},
        ),
        activeUsers: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "app_active_users",
            Help: "Number of active users",
        }),
    }
}

func (m *Metrics) Register() {
    prometheus.MustRegister(m.requestsTotal)
    prometheus.MustRegister(m.requestDuration)
    prometheus.MustRegister(m.errorsTotal)
    prometheus.MustRegister(m.activeUsers)
}
```

### 7.2 Jaeger分布式追踪

Jaeger是一个开源的分布式追踪系统。

**Jaeger客户端**

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
    jaegercfg "github.com/uber/jaeger-client-go/config"
)

func initTracer(serviceName string) (opentracing.Tracer, io.Closer, error) {
    cfg := jaegercfg.Configuration{
        ServiceName: serviceName,
        Sampler: &jaegercfg.SamplerConfig{
            Type:  jaeger.SamplerTypeConst,
            Param: 1,
        },
        Reporter: &jaegercfg.ReporterConfig{
            LogSpans: true,
        },
    }
    
    tracer, closer, err := cfg.NewTracer()
    if err != nil {
        return nil, nil, err
    }
    
    opentracing.SetGlobalTracer(tracer)
    return tracer, closer, nil
}

func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        wireContext, err := opentracing.GlobalTracer().Extract(
            opentracing.HTTPHeaders,
            opentracing.HTTPHeadersCarrier(r.Header),
        )
        
        var span opentracing.Span
        if err != nil {
            span = opentracing.StartSpan(r.URL.Path)
        } else {
            span = opentracing.StartSpan(r.URL.Path, opentracing.ChildOf(wireContext))
        }
        defer span.Finish()
        
        ctx := opentracing.ContextWithSpan(r.Context(), span)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func main() {
    tracer, closer, err := initTracer("my-service")
    if err != nil {
        log.Fatal(err)
    }
    defer closer.Close()
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
        span := opentracing.SpanFromContext(r.Context())
        span.SetTag("user.id", "123")
        
        w.Write([]byte("Data"))
    })
    
    handler := tracingMiddleware(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### 7.3 日志聚合

日志聚合将来自多个来源的日志集中管理。

**结构化日志**

```go
package main

import (
    "os"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger struct {
    *zap.Logger
}

func NewLogger(level string) (*Logger, error) {
    var zapLevel zapcore.Level
    switch level {
    case "debug":
        zapLevel = zapcore.DebugLevel
    case "info":
        zapLevel = zapcore.InfoLevel
    case "warn":
        zapLevel = zapcore.WarnLevel
    case "error":
        zapLevel = zapcore.ErrorLevel
    default:
        zapLevel = zapcore.InfoLevel
    }
    
    config := zap.Config{
        Level:            zap.NewAtomicLevelAt(zapLevel),
        Development:      false,
        Encoding:         "json",
        EncoderConfig:    zap.NewProductionEncoderConfig(),
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }
    
    logger, err := config.Build()
    if err != nil {
        return nil, err
    }
    
    return &Logger{Logger: logger}, nil
}

func (l *Logger) LogRequest(method, path string, statusCode int, duration time.Duration) {
    l.Info("HTTP Request",
        zap.String("method", method),
        zap.String("path", path),
        zap.Int("status", statusCode),
        zap.Duration("duration", duration),
    )
}

func (l *Logger) LogError(err error, fields ...zap.Field) {
    fields = append(fields, zap.Error(err))
    l.Error("Error occurred", fields...)
}
```

**日志中间件**

```go
func loggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            next.ServeHTTP(rw, r)
            
            duration := time.Since(start)
            
            logger.LogRequest(r.Method, r.URL.Path, rw.statusCode, duration)
        })
    }
}
```

### 7.4 健康检查

健康检查用于监控服务状态。

**健康检查端点**

```go
package main

import (
    "encoding/json"
    "net/http"
)

type HealthChecker struct {
    checks map[string]CheckFunc
}

type CheckFunc func() error

type HealthStatus struct {
    Status  string            `json:"status"`
    Checks  map[string]string `json:"checks"`
    Version string            `json:"version"`
}

func NewHealthChecker() *HealthChecker {
    return &HealthChecker{
        checks: make(map[string]CheckFunc),
    }
}

func (hc *HealthChecker) AddCheck(name string, check CheckFunc) {
    hc.checks[name] = check
}

func (hc *HealthChecker) Handler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        status := "healthy"
        checks := make(map[string]string)
        
        for name, check := range hc.checks {
            if err := check(); err != nil {
                status = "unhealthy"
                checks[name] = err.Error()
            } else {
                checks[name] = "ok"
            }
        }
        
        healthStatus := HealthStatus{
            Status:  status,
            Checks:  checks,
            Version: "1.0.0",
        }
        
        w.Header().Set("Content-Type", "application/json")
        
        if status == "healthy" {
            w.WriteHeader(http.StatusOK)
        } else {
            w.WriteHeader(http.StatusServiceUnavailable)
        }
        
        json.NewEncoder(w).Encode(healthStatus)
    }
}
```

**就绪检查**

```go
type ReadinessChecker struct {
    ready bool
    mu    sync.RWMutex
}

func NewReadinessChecker() *ReadinessChecker {
    return &ReadinessChecker{ready: false}
}

func (rc *ReadinessChecker) SetReady(ready bool) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    rc.ready = ready
}

func (rc *ReadinessChecker) IsReady() bool {
    rc.mu.RLock()
    defer rc.mu.RUnlock()
    return rc.ready
}

func (rc *ReadinessChecker) Handler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if rc.IsReady() {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("OK"))
        } else {
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte("Not Ready"))
        }
    }
}
```

### 7.5 小结

本章介绍了Go语言的监控和追踪，包括Prometheus监控、Jaeger分布式追踪、日志聚合、健康检查等。我们学习了Go语言如何使用Prometheus进行监控、使用Jaeger进行分布式追踪、实现结构化日志、实现健康检查和就绪检查等。掌握好Go语言的监控和追踪对于构建可观测的分布式系统非常重要。

## 总结

本文档全面介绍了Go语言的中间件与工具集成，包括：

1. **HTTP中间件**：中间件基础、日志中间件、认证中间件、CORS中间件
2. **数据库集成**：SQL数据库、ORM框架、连接池、事务处理
3. **缓存集成**：Redis缓存、Memcached缓存、本地缓存、缓存模式
4. **更多中间件**：限流中间件、压缩中间件、超时中间件、重试中间件、熔断中间件
5. **消息队列集成**：RabbitMQ、Kafka、消息模式
6. **配置管理**：环境变量、配置文件、配置中心、配置热更新
7. **监控和追踪**：Prometheus监控、Jaeger分布式追踪、日志聚合、健康检查

通过学习本文档，你应该对Go语言的中间件与工具集成有了全面的了解，能够更好地使用Go语言进行中间件开发和工具集成。希望本文档能够帮助你在Go语言的学习之路上更进一步！