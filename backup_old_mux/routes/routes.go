package routes

import (
	"piscine-devnest/internal/handlers"
	"piscine-devnest/internal/ws"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all HTTP routes for Gin
func RegisterRoutes(r *gin.Engine) {
	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Initialize handlers with the hub
	h := handlers.NewHandler(nil, hub) // Pass your DB instance instead of nil

	// Register all routes
	h.RegisterRoutes(r)
}
