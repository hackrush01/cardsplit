package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hackrush01/cardsplit/internal/api"
	"github.com/hackrush01/cardsplit/internal/config"
	"github.com/hackrush01/cardsplit/internal/handlers"
	"github.com/hackrush01/cardsplit/internal/middleware"
	"github.com/hackrush01/cardsplit/internal/storage"
)

func main() {
	log.Println("Starting CardSplit Server...")

	db := storage.InitDB("./cardsplit.db")
	defer db.Close()

	users, err := config.GetConfiguredUsers(config.MappingFilePath())
	if err != nil {
		log.Fatalf("Read card mapping: %v", err)
	}

	if err := storage.EnsureUsersExist(users, db); err != nil {
		log.Fatalf("Initialize users in DB: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.LoginHandler(db))

	mux.Handle("/statement", middleware.Auth(db, http.HandlerFunc(handlers.StatementViewHandler(db))))
	mux.Handle("/dashboard", middleware.Auth(db, middleware.AdminOnly(http.HandlerFunc(handlers.AdminDashboardHandler))))
	mux.Handle("/upload", middleware.Auth(db, middleware.AdminOnly(http.HandlerFunc(api.UploadHandler(db)))))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "CardSplit is running smoothly!")
	})

	port := ":8080"
	log.Printf("Server listening on %s\n", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
