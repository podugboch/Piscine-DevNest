package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader (must be here!)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	Clients    map[*websocket.Conn]bool
	BroadcastC chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*websocket.Conn]bool),
		BroadcastC: make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			log.Println("client registered")

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
				log.Println("client unregistered")
			}

		case message := <-h.BroadcastC:
			for client := range h.Clients {
				client.WriteMessage(websocket.TextMessage, message)
			}
		}
	}
}

func (h *Hub) HandleConnections(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// Register client
	h.Register <- wsConn

	// Read messages from client
	for {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			h.Unregister <- wsConn
			break
		}

		// Broadcast message to all connected clients
		h.Broadcast(string(msg))
	}
}

// Public method for handlers
func (h *Hub) Broadcast(msg string) {
	h.BroadcastC <- []byte(msg)
}
