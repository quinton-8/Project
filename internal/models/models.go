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
}
