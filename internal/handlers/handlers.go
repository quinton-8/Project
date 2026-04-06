package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/quinton-8/project/internal/database"
	"github.com/quinton-8/project/internal/models"
)

type AppHandler struct {
	Store *database.DataStore
}

func NewAppHandler(store *database.DataStore) *AppHandler {
	return &AppHandler{Store: store}
}

// HealthCheck is a simple endpoint to verify the API is running
func (h *AppHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "active", "message": "Taifa Care API is running. Time to connect with credible healthcare."}`))
}

// GetDoctors returns a list of available doctors for a specific region
func (h *AppHandler) GetDoctors(w http.ResponseWriter, r *http.Request) {
	// For now, hardcode the city based on the phase 1 rollout plan
	city := r.URL.Query().Get("city")
	if city == "" {
		city = "Kisumu"
	}

	doctors := h.Store.GetAvailableDoctors(city)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doctors)
}

// BookAppointment handles the core scheduling and transport logic
func (h *AppHandler) BookAppointment(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the incoming JSON request
	var req models.Appointment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 2. Generate an ID and set initial status
	req.ID = fmt.Sprintf("apt-%d", time.Now().Unix())
	req.Status = "pending_confirmation"

	// 3. Logic for transport
	message := "Schedule proposed."
	if req.NeedsTransport {
		message += " Please confirm your transport departure time via the SMS link sent."
		// Note: Here you would integrate an SMS service like Twilio or Africa's Talking later
	} else {
		message += " An SMS reminder will be sent shortly."
	}

	// 4. Save to our in-memory store
	if err := h.Store.CreateAppointment(req); err != nil {
		http.Error(w, "Failed to book appointment", http.StatusInternalServerError)
		return
	}

	// 5. Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     message,
		"appointment": req,
	})
}
