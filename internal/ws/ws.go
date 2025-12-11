package ws

import (
	"net/http"
)

// NewWebSocketHandler returns an HTTP handler for the Hub
func NewWebSocketHandler(h *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.HandleConnections(w, r)
	}
}
