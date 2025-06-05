//go:build !frontend
// +build !frontend

package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Online   bool   `json:"online"`
}

var (
	userStore = make(map[string]*User)
	userMu    sync.Mutex
)

var (
	addr       = flag.String("addr", ":8080", "http service address")
	redisAddr  = flag.String("redis", "localhost:6379", "redis service address")
	redisTopic = "chat"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Envelope is the message wrapper
type Envelope struct {
	Type    string `json:"type"`
	Sender  string `json:"sender,omitempty"`
	Content string `json:"content,omitempty"`
	Seq     int64  `json:"seq,omitempty"`
}

type Client struct {
	conn *websocket.Conn
	send chan Envelope
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Envelope
	mu         sync.Mutex
	seq        int64
	redis      *redis.Client
	ctx        context.Context
}

func newHub(rdb *redis.Client) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Envelope, 256),
		seq:        0,
		redis:      rdb,
		ctx:        context.Background(),
	}
}

func (h *Hub) run() {
	pubsub := h.redis.Subscribe(h.ctx, redisTopic)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case msg := <-ch:
			var env Envelope
			if err := json.Unmarshal([]byte(msg.Payload), &env); err == nil {
				h.mu.Lock()
				for client := range h.clients {
					select {
					case client.send <- env:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
				h.mu.Unlock()
			}
		}
	}
}

// --- REST API Handlers ---

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	userMu.Lock()
	defer userMu.Unlock()
	if _, exists := userStore[u.Username]; exists {
		http.Error(w, "user exists", http.StatusConflict)
		return
	}
	userStore[u.Username] = &User{Username: u.Username, Password: u.Password, Online: false}
	w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	userMu.Lock()
	defer userMu.Unlock()
	stored, exists := userStore[u.Username]
	if !exists || stored.Password != u.Password {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	stored.Online = true
	w.WriteHeader(http.StatusOK)
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	userMu.Lock()
	defer userMu.Unlock()
	users := make([]User, 0, len(userStore))
	for _, u := range userStore {
		users = append(users, *u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func presenceHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	status := r.URL.Query().Get("online")
	userMu.Lock()
	defer userMu.Unlock()
	if u, ok := userStore[username]; ok {
		u.Online = (status == "true")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "user not found", http.StatusNotFound)
}

type SignalKeys struct {
	IdentityKey  string   `json:"identityKey"`
	SignedPreKey string   `json:"signedPreKey"`
	PreKeys      []string `json:"preKeys"`
}

var (
	signalKeyStore = make(map[string]SignalKeys)
	signalKeyMu    sync.Mutex
)

func setupRoutes() {
	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/users", listUsersHandler)
	http.HandleFunc("/api/presence", presenceHandler)
	http.HandleFunc("/api/upload-keys", uploadKeysHandler)
	http.HandleFunc("/api/get-keys", getKeysHandler)
}

func uploadKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "missing username", http.StatusBadRequest)
		return
	}
	var keys SignalKeys
	if err := json.NewDecoder(r.Body).Decode(&keys); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	signalKeyMu.Lock()
	signalKeyStore[username] = keys
	signalKeyMu.Unlock()
	w.WriteHeader(http.StatusCreated)
}

func getKeysHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "missing username", http.StatusBadRequest)
		return
	}
	signalKeyMu.Lock()
	keys, ok := signalKeyStore[username]
	signalKeyMu.Unlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func serveWS(h *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	client := &Client{conn: conn, send: make(chan Envelope, 256)}
	h.register <- client

	// writer
	go func() {
		defer conn.Close()
		for env := range client.send {
			if err := conn.WriteJSON(env); err != nil {
				return
			}
		}
	}()

	// reader
	for {
		var env Envelope
		if err := conn.ReadJSON(&env); err != nil {
			break
		}
		// assign sequence
		h.mu.Lock()
		h.seq++
		env.Seq = h.seq
		h.mu.Unlock()

		data, _ := json.Marshal(env)
		_ = h.redis.Publish(h.ctx, redisTopic, data).Err()
	}
	h.unregister <- client
}

func main() {
	flag.Parse()
	rdb := redis.NewClient(&redis.Options{Addr: *redisAddr})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatal("redis connect:", err)
	}

	hub := newHub(rdb)
	go hub.run()

	setupRoutes()
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("../client"))))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})

	srv := &http.Server{
		Addr:         *addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Println("Chat server listening on", *addr)
	log.Fatal(srv.ListenAndServe())
}
