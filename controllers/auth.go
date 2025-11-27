package controllers

import (
	"encoding/json"
	"net/http"
)

// Dummy response structs
type AuthResponse struct {
	Message string `json:"message"`
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	resp := AuthResponse{Message: "Register endpoint hit ✅"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	resp := AuthResponse{Message: "Login endpoint hit ✅"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
