package models

import "time"

// Doctor represents pediatricians and contractual surgeons.
type Doctor struct {
	ID            string
	Name          string
	Specialty     string // e.g., "Pediatrician"
	Hospital      string // Independent or associated hospital
	IsContractual bool
	City          string // e.g., "Kisumu"
	IsEnrolled    bool
	AvailableTime []time.Time
	Lat            float64   `json:"lat"`
	Lng            float64   `json:"lng"`
	CurrentQueue   int       `json:"current_queue"`      // Number of patients currently waiting
	AvgConsultTime int       `json:"avg_consult_time"`   // Average minutes spent per patient
	Rating         float64     `json:"rating"` // New field for top-rated categories
}

// Client represents the parent/patient using the app.
type Client struct {
	ID           string
	Name         string
	LocationLat  float64
	LocationLong float64
	Registered   bool
}

// Appointment handles the booking, transport, and scheduling logic.
type Appointment struct {
	ID             string
	DoctorID       string
	ClientID       string
	ScheduledTime  time.Time
	NeedsTransport bool
	PickupPoint    string
	TransportCost  float64
	Status         string // "pending", "confirmed", "cancelled"
	DepartureTime  time.Time `json:"departure_time,omitempty"`
	ReminderTime   time.Time `json:"reminder_time,omitempty"`
	ReminderSent   bool      `json:"reminder_sent"`
}

// Hospital represents a physical location for checkups
type Hospital struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

