package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"
	"os"

	"github.com/quinton-8/project/internal/database"
	"github.com/quinton-8/project/internal/models"

	"context"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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
		message += "Please select and confirm your departure time to finalize the booking."
		// Note: Here you would integrate an SMS service like Twilio or Africa's Talking later
	} else {
		message += " Please confirm your appointment in the app. An SMS reminder will be scheduled."
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

// ConfirmAppointment finalizes the booking after transport/time agreement
func (h *AppHandler) ConfirmAppointment(w http.ResponseWriter, r *http.Request) {
	// Define the new expected payload from the frontend
	var req struct {
		AppointmentID      string  `json:"appointment_id"`
		DepartureTime      string  `json:"departure_time"`       // e.g., "2026-04-15T14:00:00Z"
		ReminderMode       string  `json:"reminder_mode"`        // "auto" or "custom"
		CustomReminderMins int     `json:"custom_reminder_mins"` // If custom, how many mins before?
		UserLat            float64 `json:"user_lat"`             // Needed for auto distance calc
		UserLng            float64 `json:"user_lng"`             // Needed for auto distance calc
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	app, err := h.Store.GetAppointment(req.AppointmentID)
	if err != nil {
		http.Error(w, "Appointment not found", http.StatusNotFound)
		return
	}

	if app.Status == "cancelled" {
		http.Error(w, "Cannot confirm a cancelled appointment", http.StatusBadRequest)
		return
	}

	// Parse the departure time chosen in the app
	depTime, err := time.Parse(time.RFC3339, req.DepartureTime)
	if err == nil {
		app.DepartureTime = depTime
	} else {
		// Fallback to now if frontend sends bad data
		app.DepartureTime = time.Now().Add(1 * time.Hour) 
	}

	// REMINDER CALCULATION LOGIC
	var reminderTime time.Time

	if req.ReminderMode == "auto" {
		// Auto calculate based on distance to the doctor
		doc, docExists := h.Store.Doctors[app.DoctorID]
		if docExists {
			dist := haversineDistance(req.UserLat, req.UserLng, doc.Lat, doc.Lng)
			travelTimeMins := int(math.Round(dist * 2)) // Assuming 0.5 km/min in city traffic
			
			// Set reminder 15 minutes BEFORE they need to start traveling
			totalBufferMins := travelTimeMins + 15
			reminderTime = app.DepartureTime.Add(-time.Duration(totalBufferMins) * time.Minute)
		} else {
			// Fallback if doctor isn't found
			reminderTime = app.DepartureTime.Add(-30 * time.Minute)
		}
	} else if req.ReminderMode == "custom" {
		// User explicitly selected something like "Remind me 45 mins before"
		reminderTime = app.DepartureTime.Add(-time.Duration(req.CustomReminderMins) * time.Minute)
	} else {
		// Default fallback
		reminderTime = app.DepartureTime.Add(-30 * time.Minute)
	}

	// Update the appointment record
	app.ReminderTime = reminderTime
	app.Status = "confirmed"
	h.Store.UpdateAppointment(app)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Appointment confirmed. SMS reminder scheduled successfully.",
		"appointment":  app,
	})
}

// CancelAppointment handles cancellations and frees up the queue
func (h *AppHandler) CancelAppointment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppointmentID string `json:"appointment_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	app, err := h.Store.GetAppointment(req.AppointmentID)
	if err != nil {
		http.Error(w, "Appointment not found", http.StatusNotFound)
		return
	}

	app.Status = "cancelled"
	h.Store.UpdateAppointment(app)

	// Note: In a full database, you would also trigger a function here to notify 
	// the next user in the "waiting" queue that a spot has opened up.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Appointment cancelled. The spot has been made available to the next incoming user.",
		"status":  app.Status,
	})
}

// haversineDistance calculates the distance between two coordinates in kilometers
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in kilometers
	
	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(dLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// GetNearestHospitals returns hospitals sorted by distance from the user's location
func (h *AppHandler) GetNearestHospitals(w http.ResponseWriter, r *http.Request) {
	// 1. Get coordinates from the URL query
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

	userLat, errLat := strconv.ParseFloat(latStr, 64)
	userLng, errLng := strconv.ParseFloat(lngStr, 64)

	if errLat != nil || errLng != nil {
		http.Error(w, "Invalid latitude or longitude provided", http.StatusBadRequest)
		return
	}

	// 2. Calculate distance for all hospitals
	type HospitalWithDistance struct {
		models.Hospital
		DistanceKM float64 `json:"distance_km"`
	}

	var results []HospitalWithDistance
	for _, hosp := range h.Store.Hospitals {
		dist := haversineDistance(userLat, userLng, hosp.Lat, hosp.Lng)
		results = append(results, HospitalWithDistance{
			Hospital:   hosp,
			DistanceKM: math.Round(dist*100) / 100, // Round to 2 decimal places
		})
	}

	// 3. Sort the results from closest to furthest
	sort.Slice(results, func(i, j int) bool {
		return results[i].DistanceKM < results[j].DistanceKM
	})

	// 4. Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// SmartMatchDoctors returns categorized lists to balance patient flow
func (h *AppHandler) SmartMatchDoctors(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

	userLat, errLat := strconv.ParseFloat(latStr, 64)
	userLng, errLng := strconv.ParseFloat(lngStr, 64)

	if errLat != nil || errLng != nil {
		http.Error(w, "Invalid latitude or longitude provided", http.StatusBadRequest)
		return
	}

	type SmartMatchResult struct {
		models.Doctor
		DistanceKM      float64 `json:"distance_km"`
		EstTravelTime   int     `json:"est_travel_time_mins"`
		EstWaitTime     int     `json:"est_wait_time_mins"`
		TotalTimeToSeen int     `json:"total_time_to_seen_mins"`
	}

	// Calculate base metrics for all valid doctors
	var allResults []SmartMatchResult
	for _, doc := range h.Store.Doctors {
		if doc.City == "Kisumu" && doc.IsEnrolled {
			dist := haversineDistance(userLat, userLng, doc.Lat, doc.Lng)
			travelTime := int(math.Round(dist * 2)) // Assuming 0.5 km/min city driving
			waitTime := doc.CurrentQueue * doc.AvgConsultTime

			allResults = append(allResults, SmartMatchResult{
				Doctor:          doc,
				DistanceKM:      math.Round(dist*100) / 100,
				EstTravelTime:   travelTime,
				EstWaitTime:     waitTime,
				TotalTimeToSeen: travelTime + waitTime,
			})
		}
	}

	// Create categorized response structure
	type CategorizedResponse struct {
		QuickestAvailable    []SmartMatchResult `json:"quickest_available"`
		TopRatedIndependents []SmartMatchResult `json:"top_rated_independents"`
		TopRatedOverall      []SmartMatchResult `json:"top_rated_overall"`
	}

	response := CategorizedResponse{
		QuickestAvailable:    make([]SmartMatchResult, len(allResults)),
		TopRatedOverall:      make([]SmartMatchResult, len(allResults)),
		TopRatedIndependents: []SmartMatchResult{}, // Will append dynamically
	}

	// 1. Populate Quickest Available (Sort by TotalTimeToSeen Ascending)
	copy(response.QuickestAvailable, allResults)
	sort.Slice(response.QuickestAvailable, func(i, j int) bool {
		return response.QuickestAvailable[i].TotalTimeToSeen < response.QuickestAvailable[j].TotalTimeToSeen
	})

	// 2. Populate Top Rated Overall (Sort by Rating Descending)
	copy(response.TopRatedOverall, allResults)
	sort.Slice(response.TopRatedOverall, func(i, j int) bool {
		return response.TopRatedOverall[i].Rating > response.TopRatedOverall[j].Rating
	})

	// 3. Populate Top Rated Independents (Filter by !IsContractual, Sort by Rating Descending)
	for _, res := range allResults {
		if !res.IsContractual {
			response.TopRatedIndependents = append(response.TopRatedIndependents, res)
		}
	}
	sort.Slice(response.TopRatedIndependents, func(i, j int) bool {
		return response.TopRatedIndependents[i].Rating > response.TopRatedIndependents[j].Rating
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AppHandler) AIRecommendDoctors(w http.ResponseWriter, r *http.Request) {
	// 1. Get user query and location
	userPreference := r.URL.Query().Get("preference") // e.g., "I don't care about distance, I just want the highest rated doctor"
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

	if userPreference == "" || latStr == "" || lngStr == "" {
		http.Error(w, "Missing preference, lat, or lng", http.StatusBadRequest)
		return
	}

	userLat, _ := strconv.ParseFloat(latStr, 64)
	userLng, _ := strconv.ParseFloat(lngStr, 64)

	// 2. Gather system data & calculate metrics
	var availableData string
	for _, doc := range h.Store.Doctors {
		if doc.City == "Kisumu" && doc.IsEnrolled {
			dist := haversineDistance(userLat, userLng, doc.Lat, doc.Lng)
			travelTime := int(math.Round(dist * 2)) // 0.5 km/min
			waitTime := doc.CurrentQueue * doc.AvgConsultTime
			totalTime := travelTime + waitTime
            
            // Format doctor details into a readable string for the AI
			availableData += fmt.Sprintf(
				"- %s (ID: %s, %s at %s). Dist: %.1f km, Wait: %d mins, Total Time to Seen: %d mins, Rating: %.1f, Contractual: %t\n",
				doc.Name, doc.ID, doc.Specialty, doc.Hospital, dist, waitTime, totalTime, doc.Rating, doc.IsContractual,
			)
		}
	}

	// 3. Initialize Gemini
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		http.Error(w, "AI Initialization failed", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash") // Fast and cheap for text summarization
	
	// 4. Construct the Prompt with System Data
	systemInstruction := `You are Taifa Care's AI Assistant. Your job is to read the live database of available doctors provided below, compare it against the user's specific preferences, and write a short, friendly summary recommending the best 1 or 2 options. Do not make up doctors. Use only the data provided.

LIVE DOCTOR DATA:
` + availableData

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	// 5. Send to AI
	resp, err := model.GenerateContent(ctx, genai.Text("User Preference: " + userPreference))
	if err != nil {
		http.Error(w, "Failed to generate AI response", http.StatusInternalServerError)
		return
	}

	// 6. Return the AI Summary
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"ai_summary": string(resp.Candidates[0].Content.Parts[0].(genai.Text)),
	})
}
