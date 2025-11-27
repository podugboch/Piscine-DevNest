package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"piscine-devnest/controllers"
)

func RegisterRoutes(r *mux.Router) {
	// Test route
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Piscine-DevNest API is running 🚀"))
	}).Methods("GET")

	// Auth routes
	r.HandleFunc("/auth/register", controllers.RegisterHandler).Methods("POST")
	r.HandleFunc("/auth/login", controllers.LoginHandler).Methods("POST")

	// Profile routes
	r.HandleFunc("/profiles", controllers.GetProfilesHandler).Methods("GET")
	r.HandleFunc("/profiles", controllers.CreateProfileHandler).Methods("POST")

	// Resource routes
	r.HandleFunc("/resources", controllers.GetResourcesHandler).Methods("GET")
	r.HandleFunc("/resources", controllers.CreateResourceHandler).Methods("POST")
}
