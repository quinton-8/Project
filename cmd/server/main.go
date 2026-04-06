package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/quinton-8/project/internal/database"
	"github.com/quinton-8/project/internal/handlers"
)

func main() {
    fmt.Println("Starting Taifa Care Server...")

	// Initialize our in-memory database with seed data
	store := database.NewDataStore()
	
	// Initialize handlers with the store
	appHandlers := handlers.NewAppHandler(store)

	// Routes
	http.HandleFunc("GET /", appHandlers.HealthCheck)
	http.HandleFunc("GET /doctors", appHandlers.GetDoctors)
	http.HandleFunc("POST /appointments/book", appHandlers.BookAppointment)

	log.Println("Server running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
