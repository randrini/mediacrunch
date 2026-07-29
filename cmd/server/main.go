package main

import (
	"fmt"
	"log"

	"github.com/mediacrunch/mediacrunch/internal/api"
	"github.com/mediacrunch/mediacrunch/internal/config"
	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/version"
)

func main() {
	cfg := config.Load()

	// Initialize database
	database, err := db.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Create router
	router := api.NewRouter(database)

	log.Printf("MediaCrunch %s", version.String())

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server listening on %s", addr)
	log.Printf("Data directory: %s", cfg.DataDir)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
