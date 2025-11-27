package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/joho/godotenv"
    "github.com/gorilla/mux"
)

func main() {
    // Load .env file
    err := godotenv.Load()
    if err != nil {
        log.Println("⚠️  No .env file found, using system environment variables")
    }

    // Create router
    r := mux.NewRouter()

    // Test route
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Piscine-DevNest API is running 🚀"))
    }).Methods("GET")

    // Read PORT from env or fallback to 8080
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    fmt.Println("🚀 Server running on http://localhost:" + port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}

