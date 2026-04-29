package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hackrush01/cardsplit/internal/api"
	"github.com/hackrush01/cardsplit/internal/handlers"
	"github.com/hackrush01/cardsplit/internal/middleware"
	"github.com/hackrush01/cardsplit/internal/storage"
)

func main() {
	log.Println("Starting CardSplit Server...")

	// 1. Initialize SQLite Database
	db := storage.InitDB("./cardsplit.db")
	defer db.Close()

	// 2. Setup standard HTTP router
	mux := http.NewServeMux()

	mux.HandleFunc("/login", handlers.LoginHandler(db))

	// Temporary health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "CardSplit is running smoothly!")
	})

	mux.Handle("/", middleware.AuthMiddleware(db, http.HandlerFunc(api.PageHandler)))
	mux.Handle("/upload", middleware.AuthMiddleware(db, http.HandlerFunc(api.UploadHandler(db))))

	// 3. Start the Server
	port := ":8080"
	log.Printf("Server listening on %s\n", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
