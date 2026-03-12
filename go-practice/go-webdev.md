# Go语言网络开发

## 前言

本文档将全面介绍Go语言的网络开发，帮助读者更好地理解和使用Go语言进行网络编程。通过学习本文档，你将掌握Go语言网络开发的核心概念和常用技巧，包括HTTP服务、TCP/UDP编程、WebSocket、RPC等。

## 第一章：HTTP服务

### 1.1 基础HTTP服务

Go语言的net/http包提供了强大的HTTP服务功能。

**简单的HTTP服务器**

```go
package main

import (
    "fmt"
    "net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func main() {
    http.HandleFunc("/", helloHandler)
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```

**使用http.Server**

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func main() {
    server := &http.Server{
        Addr:         ":8080",
        Handler:      http.HandlerFunc(handler),
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        fmt.Println("Server starting on :8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("Server error: %v\n", err)
        }
    }()

    gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) {
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        fmt.Printf("Server shutdown error: %v\n", err)
    }
    fmt.Println("Server stopped")
}
```

### 1.2 HTTP路由

Go语言支持多种路由方式。

**使用ServeMux**

```go
package main

import (
    "fmt"
    "net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Home Page")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "About Page")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Contact Page")
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", homeHandler)
    mux.HandleFunc("/about", aboutHandler)
    mux.HandleFunc("/contact", contactHandler)

    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", mux)
}
```

**使用Gin框架**

```go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello, World!",
        })
    })

    r.GET("/users/:id", func(c *gin.Context) {
        id := c.Param("id")
        c.JSON(200, gin.H{
            "user_id": id,
        })
    })

    r.POST("/users", func(c *gin.Context) {
        var user User
        if err := c.BindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        c.JSON(201, user)
    })

    r.Run(":8080")
}

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

### 1.3 HTTP客户端

Go语言的net/http包也提供了HTTP客户端功能。

**GET请求**

```go
package main

import (
    "fmt"
    "io"
    "net/http"
)

func main() {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Response: %s\n", body)
}
```

**POST请求**

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    user := User{
        Name:  "Alice",
        Email: "alice@example.com",
    }

    jsonData, err := json.Marshal(user)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    resp, err := http.Post(
        "https://api.example.com/users",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()

    fmt.Printf("Status: %s\n", resp.Status)
}
```

**使用http.Client**

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

func main() {
    client := &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()

    fmt.Printf("Status: %s\n", resp.Status)
}
```

### 1.4 中间件

中间件是HTTP服务中常用的模式。

**日志中间件**

```go
package main

import (
    "log"
    "net/http"
    "time"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        next.ServeHTTP(w, r)
        
        log.Printf(
            "%s %s %s %v",
            r.Method,
            r.RequestURI,
            r.RemoteAddr,
            time.Since(start),
        )
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

**认证中间件**

```go
package main

import (
    "net/http"
)

type AuthMiddleware struct {
    apiKey string
}

func NewAuthMiddleware(apiKey string) *AuthMiddleware {
    return &AuthMiddleware{apiKey: apiKey}
}

func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := r.Header.Get("X-API-Key")
        if key != am.apiKey {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func main() {
    auth := NewAuthMiddleware("secret-key")
    
    mux := http.NewServeMux()
    mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Protected resource"))
    })

    handler := auth.Middleware(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### 1.5 RESTful API

让我们创建一个完整的RESTful API。

```go
package main

import (
    "encoding/json"
    "net/http"
    "sync"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserStore struct {
    users map[string]*User
    mu    sync.RWMutex
}

func NewUserStore() *UserStore {
    return &UserStore{
        users: make(map[string]*User),
    }
}

func (s *UserStore) Create(user *User) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.users[user.ID]; exists {
        return fmt.Errorf("user already exists")
    }
    
    s.users[user.ID] = user
    return nil
}

func (s *UserStore) Get(id string) (*User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    user, exists := s.users[id]
    if !exists {
        return nil, fmt.Errorf("user not found")
    }
    
    return user, nil
}

func (s *UserStore) Update(user *User) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.users[user.ID]; !exists {
        return fmt.Errorf("user not found")
    }
    
    s.users[user.ID] = user
    return nil
}

func (s *UserStore) Delete(id string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.users[id]; !exists {
        return fmt.Errorf("user not found")
    }
    
    delete(s.users, id)
    return nil
}

func (s *UserStore) List() []*User {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    users := make([]*User, 0, len(s.users))
    for _, user := range s.users {
        users = append(users, user)
    }
    
    return users
}

type API struct {
    store *UserStore
}

func NewAPI(store *UserStore) *API {
    return &API{store: store}
}

func (api *API) handleCreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    if err := api.store.Create(&user); err != nil {
        http.Error(w, err.Error(), http.StatusConflict)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func (api *API) handleGetUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/users/"):]
    
    user, err := api.store.Get(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func (api *API) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/users/"):]
    
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    user.ID = id
    if err := api.store.Update(&user); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func (api *API) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/users/"):]
    
    if err := api.store.Delete(id); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}

func (api *API) handleListUsers(w http.ResponseWriter, r *http.Request) {
    users := api.store.List()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func main() {
    store := NewUserStore()
    api := NewAPI(store)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodPost:
            api.handleCreateUser(w, r)
        case http.MethodGet:
            api.handleListUsers(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })
    
    mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            api.handleGetUser(w, r)
        case http.MethodPut:
            api.handleUpdateUser(w, r)
        case http.MethodDelete:
            api.handleDeleteUser(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })
    
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", mux)
}
```

### 1.6 小结

本章介绍了Go语言的HTTP服务，包括基础HTTP服务、HTTP路由、HTTP客户端、中间件、RESTful API等。我们学习了Go语言如何创建HTTP服务、处理HTTP请求、使用中间件、构建RESTful API等。掌握好Go语言的HTTP服务对于编写Web应用非常重要。

下一章我们将学习Go语言的TCP/UDP编程，包括TCP服务器、UDP服务器、Socket编程等，以及它们的使用方法和注意事项。

## 第二章：TCP/UDP编程

### 2.1 TCP服务器

Go语言的net包提供了TCP网络编程功能。

**简单的TCP服务器**

```go
package main

import (
    "bufio"
    "fmt"
    "net"
)

func handleConnection(conn net.Conn) {
    defer conn.Close()
    
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        message := scanner.Text()
        fmt.Printf("Received: %s\n", message)
        
        response := fmt.Sprintf("Echo: %s\n", message)
        conn.Write([]byte(response))
    }
    
    if err := scanner.Err(); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer listener.Close()
    
    fmt.Println("TCP server listening on :8080")
    
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        
        go handleConnection(conn)
    }
}
```

**并发TCP服务器**

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "net"
    "sync"
    "time"
)

type TCPServer struct {
    listener net.Listener
    wg       sync.WaitGroup
    ctx      context.Context
    cancel   context.CancelFunc
}

func NewTCPServer(addr string) (*TCPServer, error) {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return nil, err
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    
    return &TCPServer{
        listener: listener,
        ctx:      ctx,
        cancel:   cancel,
    }, nil
}

func (s *TCPServer) Start() {
    fmt.Printf("TCP server listening on %s\n", s.listener.Addr())
    
    for {
        select {
        case <-s.ctx.Done():
            return
        default:
            conn, err := s.listener.Accept()
            if err != nil {
                select {
                case <-s.ctx.Done():
                    return
                default:
                    fmt.Printf("Accept error: %v\n", err)
                    continue
                }
            }
            
            s.wg.Add(1)
            go s.handleConnection(conn)
        }
    }
}

func (s *TCPServer) handleConnection(conn net.Conn) {
    defer s.wg.Done()
    defer conn.Close()
    
    addr := conn.RemoteAddr()
    fmt.Printf("New connection from %s\n", addr)
    
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        select {
        case <-s.ctx.Done():
            return
        default:
            message := scanner.Text()
            fmt.Printf("[%s] Received: %s\n", addr, message)
            
            response := fmt.Sprintf("Echo: %s\n", message)
            conn.Write([]byte(response))
        }
    }
    
    if err := scanner.Err(); err != nil {
        fmt.Printf("[%s] Error: %v\n", addr, err)
    }
    
    fmt.Printf("[%s] Connection closed\n", addr)
}

func (s *TCPServer) Stop() {
    s.cancel()
    s.listener.Close()
    s.wg.Wait()
    fmt.Println("TCP server stopped")
}

func main() {
    server, err := NewTCPServer(":8080")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    go server.Start()
    
    time.Sleep(30 * time.Second)
    server.Stop()
}
```

### 2.2 TCP客户端

**简单的TCP客户端**

```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer conn.Close()
    
    fmt.Println("Connected to server")
    
    go func() {
        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            fmt.Printf("Server: %s\n", scanner.Text())
        }
    }()
    
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        message := scanner.Text()
        conn.Write([]byte(message + "\n"))
    }
}
```

**带超时的TCP客户端**

```go
package main

import (
    "context"
    "fmt"
    "net"
    "time"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    dialer := &net.Dialer{}
    conn, err := dialer.DialContext(ctx, "tcp", "localhost:8080")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer conn.Close()
    
    fmt.Println("Connected to server")
    
    deadline, _ := ctx.Deadline()
    conn.SetDeadline(deadline)
    
    _, err = conn.Write([]byte("Hello\n"))
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Server response: %s\n", string(buffer[:n]))
}
```

### 2.3 UDP服务器

UDP是无连接的协议，适用于不需要可靠传输的场景。

**简单的UDP服务器**

```go
package main

import (
    "fmt"
    "net"
)

func main() {
    addr, err := net.ResolveUDPAddr("udp", ":8080")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer conn.Close()
    
    fmt.Println("UDP server listening on :8080")
    
    buffer := make([]byte, 1024)
    for {
        n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        
        message := string(buffer[:n])
        fmt.Printf("Received from %s: %s\n", clientAddr, message)
        
        response := fmt.Sprintf("Echo: %s", message)
        conn.WriteToUDP([]byte(response), clientAddr)
    }
}
```

**并发UDP服务器**

```go
package main

import (
    "context"
    "fmt"
    "net"
    "sync"
    "time"
)

type UDPServer struct {
    conn *net.UDPConn
    wg   sync.WaitGroup
    ctx  context.Context
    cancel context.CancelFunc
}

func NewUDPServer(addr string) (*UDPServer, error) {
    udpAddr, err := net.ResolveUDPAddr("udp", addr)
    if err != nil {
        return nil, err
    }
    
    conn, err := net.ListenUDP("udp", udpAddr)
    if err != nil {
        return nil, err
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    
    return &UDPServer{
        conn:   conn,
        ctx:    ctx,
        cancel: cancel,
    }, nil
}

func (s *UDPServer) Start() {
    fmt.Printf("UDP server listening on %s\n", s.conn.LocalAddr())
    
    buffer := make([]byte, 1024)
    for {
        select {
        case <-s.ctx.Done():
            return
        default:
            s.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
            
            n, clientAddr, err := s.conn.ReadFromUDP(buffer)
            if err != nil {
                if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                    continue
                }
                fmt.Printf("Error: %v\n", err)
                continue
            }
            
            s.wg.Add(1)
            go s.handleMessage(buffer[:n], clientAddr)
        }
    }
}

func (s *UDPServer) handleMessage(data []byte, addr *net.UDPAddr) {
    defer s.wg.Done()
    
    message := string(data)
    fmt.Printf("[%s] Received: %s\n", addr, message)
    
    response := fmt.Sprintf("Echo: %s", message)
    s.conn.WriteToUDP([]byte(response), addr)
}

func (s *UDPServer) Stop() {
    s.cancel()
    s.wg.Wait()
    s.conn.Close()
    fmt.Println("UDP server stopped")
}

func main() {
    server, err := NewUDPServer(":8080")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    go server.Start()
    
    time.Sleep(30 * time.Second)
    server.Stop()
}
```

### 2.4 UDP客户端

**简单的UDP客户端**

```go
package main

import (
    "fmt"
    "net"
)

func main() {
    serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    conn, err := net.DialUDP("udp", nil, serverAddr)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer conn.Close()
    
    fmt.Println("Connected to server")
    
    message := "Hello, UDP Server!"
    _, err = conn.Write([]byte(message))
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Server response: %s\n", string(buffer[:n]))
}
```

### 2.5 实例：聊天服务器

让我们创建一个简单的聊天服务器。

```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "sync"
)

type ChatServer struct {
    clients map[net.Conn]string
    mu      sync.RWMutex
    join    chan net.Conn
    leave   chan net.Conn
    message chan string
}

func NewChatServer() *ChatServer {
    return &ChatServer{
        clients: make(map[net.Conn]string),
        join:    make(chan net.Conn),
        leave:   make(chan net.Conn),
        message: make(chan string),
    }
}

func (s *ChatServer) Start(addr string) error {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    defer listener.Close()
    
    fmt.Printf("Chat server listening on %s\n", addr)
    
    go s.broadcast()
    
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Accept error: %v\n", err)
            continue
        }
        
        go s.handleClient(conn)
    }
}

func (s *ChatServer) handleClient(conn net.Conn) {
    defer conn.Close()
    
    scanner := bufio.NewScanner(conn)
    
    if !scanner.Scan() {
        return
    }
    
    username := scanner.Text()
    s.join <- conn
    
    s.mu.Lock()
    s.clients[conn] = username
    s.mu.Unlock()
    
    s.message <- fmt.Sprintf("%s has joined the chat", username)
    
    for scanner.Scan() {
        text := scanner.Text()
        s.message <- fmt.Sprintf("%s: %s", username, text)
    }
    
    s.leave <- conn
    s.message <- fmt.Sprintf("%s has left the chat", username)
}

func (s *ChatServer) broadcast() {
    for {
        select {
        case conn := <-s.join:
            s.mu.RLock()
            username := s.clients[conn]
            s.mu.RUnlock()
            fmt.Printf("%s joined\n", username)
            
        case conn := <-s.leave:
            s.mu.Lock()
            username := s.clients[conn]
            delete(s.clients, conn)
            s.mu.Unlock()
            fmt.Printf("%s left\n", username)
            
        case msg := <-s.message:
            s.mu.RLock()
            for conn := range s.clients {
                conn.Write([]byte(msg + "\n"))
            }
            s.mu.RUnlock()
        }
    }
}

func main() {
    server := NewChatServer()
    server.Start(":8080")
}
```

### 2.6 小结

本章介绍了Go语言的TCP/UDP编程，包括TCP服务器、TCP客户端、UDP服务器、UDP客户端等。我们学习了Go语言如何创建TCP/UDP服务器、处理网络连接、实现并发网络服务等。掌握好Go语言的TCP/UDP编程对于编写网络应用非常重要。

下一章我们将学习Go语言的WebSocket编程，包括WebSocket服务器、WebSocket客户端、实时通信等，以及它们的使用方法和注意事项。

## 第三章：WebSocket编程

### 3.1 WebSocket基础

WebSocket是一种在单个TCP连接上进行全双工通信的协议。

**使用gorilla/websocket**

```go
package main

import (
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize: 1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    defer conn.Close()

    log.Println("Client connected")

    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Read error: %v", err)
            break
        }

        log.Printf("Received: %s", message)

        err = conn.WriteMessage(messageType, message)
        if err != nil {
            log.Printf("Write error: %v", err)
            break
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleWebSocket)
    log.Println("WebSocket server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 3.2 WebSocket服务器

**Echo服务器**

```go
package main

import (
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type Client struct {
    conn *websocket.Conn
    send chan []byte
}

func (c *Client) readPump() {
    defer func() {
        c.conn.Close()
    }()

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            log.Printf("Read error: %v", err)
            break
        }

        c.send <- message
    }
}

func (c *Client) writePump() {
    defer func() {
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            err := c.conn.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Printf("Write error: %v", err)
                return
            }
        }
    }
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }

    client := &Client{
        conn: conn,
        send: make(chan []byte, 256),
    }

    go client.writePump()
    go client.readPump()
}

func main() {
    http.HandleFunc("/ws", handleWebSocket)
    log.Println("Echo server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**聊天服务器**

```go
package main

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type ChatServer struct {
    clients map[*Client]bool
    mu      sync.RWMutex
    join    chan *Client
    leave   chan *Client
    message chan []byte
}

type Client struct {
    server *ChatServer
    conn   *websocket.Conn
    send   chan []byte
}

func NewChatServer() *ChatServer {
    return &ChatServer{
        clients: make(map[*Client]bool),
        join:    make(chan *Client),
        leave:   make(chan *Client),
        message: make(chan []byte),
    }
}

func (s *ChatServer) Run() {
    for {
        select {
        case client := <-s.join:
            s.mu.Lock()
            s.clients[client] = true
            s.mu.Unlock()
            log.Printf("Client joined, total: %d", len(s.clients))

        case client := <-s.leave:
            s.mu.Lock()
            delete(s.clients, client)
            s.mu.Unlock()
            log.Printf("Client left, total: %d", len(s.clients))

        case msg := <-s.message:
            s.mu.RLock()
            for client := range s.clients {
                select {
                case client.send <- msg:
                default:
                    delete(s.clients, client)
                    close(client.send)
                }
            }
            s.mu.RUnlock()
        }
    }
}

func (c *Client) readPump() {
    defer func() {
        c.server.leave <- c
        c.conn.Close()
    }()

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            log.Printf("Read error: %v", err)
            break
        }

        c.server.message <- message
    }
}

func (c *Client) writePump() {
    defer func() {
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            err := c.conn.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Printf("Write error: %v", err)
                return
            }
        }
    }
}

func handleWebSocket(server *ChatServer) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Printf("WebSocket upgrade error: %v", err)
            return
        }

        client := &Client{
            server: server,
            conn:   conn,
            send:   make(chan []byte, 256),
        }

        server.join <- client

        go client.writePump()
        go client.readPump()
    }
}

func main() {
    server := NewChatServer()
    go server.Run()

    http.HandleFunc("/ws", handleWebSocket(server))
    log.Println("Chat server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 3.3 WebSocket客户端

**基本客户端**

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/gorilla/websocket"
)

func main() {
    url := "ws://localhost:8080/ws"
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        log.Fatal("Dial error:", err)
    }
    defer conn.Close()

    done := make(chan struct{})

    go func() {
        defer close(done)
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("Read error:", err)
                return
            }
            fmt.Printf("Received: %s\n", message)
        }
    }()

    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

    for {
        select {
        case <-done:
            return
        case <-interrupt:
            err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
            if err != nil {
                log.Println("Write close error:", err)
                return
            }
            select {
            case <-done:
            case <-time.After(time.Second):
            }
            return
        }
    }
}
```

**交互式客户端**

```go
package main

import (
    "bufio"
    "fmt"
    "log"
    "os"

    "github.com/gorilla/websocket"
)

func main() {
    url := "ws://localhost:8080/ws"
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        log.Fatal("Dial error:", err)
    }
    defer conn.Close()

    done := make(chan struct{})

    go func() {
        defer close(done)
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("Read error:", err)
                return
            }
            fmt.Printf("\nReceived: %s\n> ", message)
        }
    }()

    scanner := bufio.NewScanner(os.Stdin)
    fmt.Print("> ")
    for scanner.Scan() {
        text := scanner.Text()
        if text == "quit" {
            break
        }

        err := conn.WriteMessage(websocket.TextMessage, []byte(text))
        if err != nil {
            log.Println("Write error:", err)
            break
        }

        fmt.Print("> ")
    }

    <-done
}
```

### 3.4 实例：实时通知系统

让我们创建一个实时通知系统。

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type Notification struct {
    Type      string    `json:"type"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}

type NotificationServer struct {
    clients map[*Client]bool
    mu      sync.RWMutex
    join    chan *Client
    leave   chan *Client
    notify  chan Notification
}

type Client struct {
    server *NotificationServer
    conn   *websocket.Conn
    send   chan []byte
}

func NewNotificationServer() *NotificationServer {
    return &NotificationServer{
        clients: make(map[*Client]bool),
        join:    make(chan *Client),
        leave:   make(chan *Client),
        notify:  make(chan Notification),
    }
}

func (s *NotificationServer) Run() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case client := <-s.join:
            s.mu.Lock()
            s.clients[client] = true
            s.mu.Unlock()
            log.Printf("Client joined, total: %d", len(s.clients))

        case client := <-s.leave:
            s.mu.Lock()
            delete(s.clients, client)
            s.mu.Unlock()
            log.Printf("Client left, total: %d", len(s.clients))

        case notification := <-s.notify:
            data, err := json.Marshal(notification)
            if err != nil {
                log.Printf("Marshal error: %v", err)
                continue
            }

            s.mu.RLock()
            for client := range s.clients {
                select {
                case client.send <- data:
                default:
                    delete(s.clients, client)
                    close(client.send)
                }
            }
            s.mu.RUnlock()

        case <-ticker.C:
            notification := Notification{
                Type:      "heartbeat",
                Message:   "Server is alive",
                Timestamp: time.Now(),
            }
            s.notify <- notification
        }
    }
}

func (c *Client) readPump() {
    defer func() {
        c.server.leave <- c
        c.conn.Close()
    }()

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            log.Printf("Read error: %v", err)
            break
        }

        log.Printf("Received: %s", message)
    }
}

func (c *Client) writePump() {
    defer func() {
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            err := c.conn.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Printf("Write error: %v", err)
                return
            }
        }
    }
}

func handleWebSocket(server *NotificationServer) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Printf("WebSocket upgrade error: %v", err)
            return
        }

        client := &Client{
            server: server,
            conn:   conn,
            send:   make(chan []byte, 256),
        }

        server.join <- client

        go client.writePump()
        go client.readPump()
    }
}

func handleNotify(server *NotificationServer) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var notification Notification
        if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        notification.Timestamp = time.Now()
        server.notify <- notification

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Notification sent"))
    }
}

func main() {
    server := NewNotificationServer()
    go server.Run()

    mux := http.NewServeMux()
    mux.HandleFunc("/ws", handleWebSocket(server))
    mux.HandleFunc("/notify", handleNotify(server))

    log.Println("Notification server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

### 3.5 小结

本章介绍了Go语言的WebSocket编程，包括WebSocket基础、WebSocket服务器、WebSocket客户端、实时通信等。我们学习了Go语言如何创建WebSocket服务、处理WebSocket连接、实现实时通信等。掌握好Go语言的WebSocket编程对于编写实时应用非常重要。

## 总结

本文档全面介绍了Go语言的网络开发，包括：

1. **HTTP服务**：基础HTTP服务、HTTP路由、HTTP客户端、中间件、RESTful API
2. **TCP/UDP编程**：TCP服务器、TCP客户端、UDP服务器、UDP客户端、聊天服务器
3. **WebSocket编程**：WebSocket基础、WebSocket服务器、WebSocket客户端、实时通知系统

通过学习本文档，你应该对Go语言的网络开发有了全面的了解，能够更好地使用Go语言进行网络编程。希望本文档能够帮助你在Go语言的学习之路上更进一步！