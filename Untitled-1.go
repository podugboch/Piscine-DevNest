package piscinedevnest
// Piscine-DevNest — Backend (Go)
// Project scaffold: a starter REST + WebSocket backend for Piscine-DevNest
// Files shown below. Save each into its own file in a single repo.

// -----------------------------
// File: go.mod
// -----------------------------
module piscine-devnest

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
	github.com/golang-jwt/jwt/v5 v5.0.0
	gorm.io/driver/postgres v1.5.0
	gorm.io/gorm v1.26.0
	github.com/gorilla/websocket v1.5.0
	golang.org/x/crypto v0.11.0
)

// -----------------------------
// File: README.md
// -----------------------------
# Piscine-DevNest - Backend (Go)

This repo is a starter backend for the Piscine-DevNest social & collaboration app.

Features included in this scaffold
- REST API (Gin)
- JWT-based authentication (email/password)
- Profile CRUD and searchable directory
- Basic resource sharing endpoints
- WebSocket hub for real-time chat
- GORM + Postgres for persistence
- Docker Compose example

## Quick start
1. Create `.env` with DATABASE_URL and JWT_SECRET
2. `go run ./cmd/server`
3. Use the HTTP endpoints documented in handlers


// -----------------------------
// File: .env.example
// -----------------------------
# DATABASE_URL=postgres://user:password@localhost:5432/piscinedevnest?sslmode=disable
# JWT_SECRET=replace_with_a_strong_secret


// -----------------------------
// File: cmd/server/main.go
// -----------------------------
package main

import (
	"log"
	"os"

	"piscine-devnest/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}


// -----------------------------
// File: internal/app/app.go
// -----------------------------
package app

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"piscine-devnest/internal/handlers"
	"piscine-devnest/internal/model"
	"piscine-devnest/internal/ws"
)

func Run() error {
	// Load env
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/piscine?sslmode=disable"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}

	// DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	// Migrate
	if err := db.AutoMigrate(&model.User{}, &model.Resource{}, &model.Connection{}); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	// create Gin
	r := gin.Default()

	// initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// wire handlers
	h := handlers.NewHandler(db, jwtSecret, hub)
	h.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return r.Run(":" + port)
}


// -----------------------------
// File: internal/model/models.go
// -----------------------------
package model

import "time"

// User represents a Piscine candidate
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Name      string    `json:"name"`
	Bio       string    `json:"bio"`
	Skills    string    `json:"skills"` // comma-separated for simplicity
	Batch     string    `json:"batch"`
	Location  string    `json:"location"`
	AvatarURL string    `json:"avatar_url"`
}

// Resource is a shared note / link / snippet
type Resource struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	OwnerID   uint      `json:"owner_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"` // text or snippet
	Link      string    `json:"link"`
	Likes     int       `json:"likes"`
}

// Connection represents friendship/connection between users
type Connection struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	FromID    uint `json:"from_id"`
	ToID      uint `json:"to_id"`
	Accepted  bool `json:"accepted"`
}


// -----------------------------
// File: internal/handlers/handlers.go
// -----------------------------
package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"piscine-devnest/internal/model"
	"piscine-devnest/internal/ws"
)

type Handler struct {
	db        *gorm.DB
	jwtSecret string
	hub       *ws.Hub
}

func NewHandler(db *gorm.DB, jwtSecret string, hub *ws.Hub) *Handler {
	return &Handler{db: db, jwtSecret: jwtSecret, hub: hub}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)

		api.GET("/profiles", h.ListProfiles)
		api.GET("/profiles/:id", h.GetProfile)

		api.POST("/resources", h.AuthMiddleware(), h.CreateResource)
		api.GET("/resources", h.ListResources)

		api.POST("/ws", h.WSHandler)
	}
}

// Register — simple email/password registration (password stored hashed)
func (h *Handler) Register(c *gin.Context) {
	var in struct {
		Email string `json:"email" binding:"required,email"`
		Pass  string `json:"password" binding:"required,min=6"`
		Name  string `json:"name"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// hash password
	hashed := HashPassword(in.Pass)
	u := model.User{Email: strings.ToLower(in.Email), Password: hashed, Name: in.Name}
	if err := h.db.Create(&u).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already used or invalid"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email})
}

// Login — returns JWT
func (h *Handler) Login(c *gin.Context) {
	var in struct {
		Email string `json:"email" binding:"required,email"`
		Pass  string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var u model.User
	if err := h.db.Where("email = ?", strings.ToLower(in.Email)).First(&u).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !CheckPasswordHash(in.Pass, u.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	// create token
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID,
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	})
	signed, _ := t.SignedString([]byte(h.jwtSecret))
	c.JSON(http.StatusOK, gin.H{"token": signed})
}

// AuthMiddleware extracts user ID from JWT and puts into context
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tok := strings.TrimPrefix(auth, "Bearer ")
		parsed, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.jwtSecret), nil
		})
		if err != nil || !parsed.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims := parsed.Claims.(jwt.MapClaims)
		uid := uint(claims["sub"].(float64))
		c.Set("user_id", uid)
		c.Next()
	}
}

// ListProfiles — supports simple search by skill, batch, location
func (h *Handler) ListProfiles(c *gin.Context) {
	q := c.Query("q")
	skill := c.Query("skill")
	batch := c.Query("batch")
	loc := c.Query("location")

	var users []model.User
	db := h.db
	if q != "" {
		db = db.Where("name ILIKE ? OR bio ILIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if skill != "" {
		db = db.Where("skills ILIKE ?", "%"+skill+"%")
	}
	if batch != "" {
		db = db.Where("batch = ?", batch)
	}
	if loc != "" {
		db = db.Where("location ILIKE ?", "%"+loc+"%")
	}
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	var u model.User
	if err := h.db.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// CreateResource
func (h *Handler) CreateResource(c *gin.Context) {
	var in struct {
		Title string `json:"title" binding:"required"`
		Body  string `json:"body"`
		Link  string `json:"link"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := c.Get("user_id")
	r := model.Resource{Title: in.Title, Body: in.Body, Link: in.Link, OwnerID: uid.(uint)}
	if err := h.db.Create(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create"})
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *Handler) ListResources(c *gin.Context) {
	var rs []model.Resource
	if err := h.db.Order("created_at desc").Limit(50).Find(&rs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, rs)
}

// WSHandler accepts WebSocket upgrades and hooks clients into hub
func (h *Handler) WSHandler(c *gin.Context) {
	hub := h.hub
	wsHandler := ws.NewWSHandler(hub)
	wsHandler.ServeHTTP(c.Writer, c.Request)
}


// -----------------------------
// File: internal/handlers/auth_utils.go
// -----------------------------
package handlers

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(p string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b)
}

func CheckPasswordHash(p, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil
}


// -----------------------------
// File: internal/ws/hub.go
// -----------------------------
package ws

import (
	"log"
)

// Hub maintains clients and broadcasts
type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	clients    map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		clients:    map[*Client]bool{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			log.Println("client registered")
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}


// -----------------------------
// File: internal/ws/ws.go
// -----------------------------
package ws

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func NewWSHandler(hub *Hub) *http.HandlerFunc {
	h := func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		go client.writePump()
		client.readPump()
	}
	return &h
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		c.hub.broadcast <- msg
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}


// -----------------------------
// File: docker-compose.yml
// -----------------------------
version: "3.8"
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: piscine
    ports:
      - "5432:5432"
    volumes:
      - db-data:/var/lib/postgresql/data

  api:
    build: .
    command: ["/piscine-devnest"]
    environment:
      DATABASE_URL: postgres://postgres:postgres@db:5432/piscine?sslmode=disable
      JWT_SECRET: supersecret
    ports:
      - "8080:8080"
    depends_on:
      - db

volumes:
  db-data:


// -----------------------------
// Notes & Next Steps
// -----------------------------
// - This scaffold is intentionally minimal and built to be a starting point.
// - Add validations, pagination, file uploads, rate limiting, and secure CORS.
// - For realtime chat persistence, extend the hub to store messages into Postgres.
// - For production, never use CheckOrigin: always validate origins and use HTTPS, strong JWT keys, and secure cookie flags if using cookies.
// - You can replace GORM with sqlc or pgx if you prefer typed queries.

// If you want, I can:
// - split this into separate files in a GitHub-ready repo and provide Dockerfile + Makefile
// - add sample frontend (React Native) to connect to the API and WS
// - add tests for handlers and unit tests for auth

// End of scaffold
