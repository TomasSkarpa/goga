package main

import (
	"goga/internal/server"
	"log"
	"os"
)

func main() {
	// Configuration
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./goga.db")
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")

	// Initialize server
	srv, err := server.New(dbPath, uploadDir)
	if err != nil {
		log.Fatal("Failed to initialize server:", err)
	}
	defer srv.Close()

	// Start server
	if err := srv.Start(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}