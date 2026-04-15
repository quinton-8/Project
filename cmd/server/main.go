package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/quinton-8/project/internal/database"
	"github.com/quinton-8/project/internal/handlers"
	"github.com/quinton-8/project/internal/worker"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found. Relying on system environment variables.")
	}
    fmt.Println("Starting Taifa Care Server...")

	// Initialize our in-memory database with seed data
	store := database.NewDataStore()
	
	// Initialize handlers with the store
	appHandlers := handlers.NewAppHandler(store)

	mux := http.NewServeMux()
	// Routes
	mux.HandleFunc("GET /", appHandlers.HealthCheck)
	mux.HandleFunc("GET /doctors", appHandlers.GetDoctors)
	mux.HandleFunc("POST /appointments/book", appHandlers.BookAppointment)
	
	// New Routes for expanding logic
	mux.HandleFunc("POST /appointments/confirm", appHandlers.ConfirmAppointment)
	mux.HandleFunc("POST /appointments/cancel", appHandlers.CancelAppointment)
	mux.HandleFunc("GET /hospitals/nearest", appHandlers.GetNearestHospitals)
	mux.HandleFunc("GET /doctors/smart-match", appHandlers.SmartMatchDoctors)

	// NEW AI ROUTE
    mux.HandleFunc("GET /ai/recommend", appHandlers.AIRecommendDoctors)

	// START THE BACKGROUND WORKER
	go worker.StartReminderJob(store)

	log.Println("Server running on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
