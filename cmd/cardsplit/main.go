package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hackrush01/cardsplit/internal/api"
	"github.com/hackrush01/cardsplit/internal/config"
	"github.com/hackrush01/cardsplit/internal/handlers"
	"github.com/hackrush01/cardsplit/internal/middleware"
	"github.com/hackrush01/cardsplit/internal/storage"
)

func main() {
	log.Println("Starting CardSplit Server...")

	db := storage.InitDB()
	defer db.Close()

	config.LoadCardRules(config.RuleFilePath())

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
	mux.Handle("/dashboard", middleware.Auth(db, middleware.AdminOnly(http.HandlerFunc(handlers.AdminDashboardHandler(db)))))
	mux.Handle("/upload", middleware.Auth(db, middleware.AdminOnly(http.HandlerFunc(api.UploadHandler(db)))))
	mux.Handle("/adjustment", middleware.Auth(db, middleware.AdminOnly(http.HandlerFunc(api.AdjustmentHandler(db)))))
	mux.Handle("/load-statement", middleware.Auth(db, middleware.AdminOnly(http.HandlerFunc(api.LoadStatementHandler(db)))))
	mux.Handle("/statement-dates", middleware.Auth(db, middleware.AdminOnly(http.HandlerFunc(api.StatementDatesOptionsHandler(db)))))
	mux.Handle("/reset-password", middleware.Auth(db, middleware.AdminOnly(http.HandlerFunc(api.ResetPasswordHandler(db)))))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "CardSplit is running smoothly!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on 0.0.0.0:%s\n", port)

	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
