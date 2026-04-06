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

// func healthCheck(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Hello world. Time to connect with the most credible health care all at your convenience."))
// }

// // enrollDoctor handles the onboarding of new healthcare professionals
// func enrollDoctor(w http.ResponseWriter, r *http.Request) {
//     // In a full implementation, parse JSON from r.Body and save to DB
// 	w.WriteHeader(http.StatusCreated)
// 	w.Write([]byte(`{"message": "Doctor enrolled successfully. Awaiting data collection phase."}`))
// }

// // bookAppointment manages the scheduling and transport confirmation
// func bookAppointment(w http.ResponseWriter, r *http.Request) {
//     // Logic to check doctor availability, assess transport needs, and confirm time
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"message": "Schedule proposed. Please confirm transport and timing."}`))
// }