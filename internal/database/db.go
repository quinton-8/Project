package database

import (
	"errors"
	"time"

	"github.com/quinton-8/project/internal/models"
)

// DataStore represents our in-memory database
type DataStore struct {
	Doctors      map[string]models.Doctor
	Clients      map[string]models.Client
	Appointments map[string]models.Appointment
}

// NewDataStore initializes the store with some hardcoded seed data
func NewDataStore() *DataStore {
	store := &DataStore{
		Doctors:      make(map[string]models.Doctor),
		Clients:      make(map[string]models.Client),
		Appointments: make(map[string]models.Appointment),
	}

	// Seed Data: Contractual and Independent Doctors in the target region
	store.Doctors["doc-1"] = models.Doctor{
		ID:            "doc-1",
		Name:          "Dr. Sarah Ochieng",
		Specialty:     "Pediatrician",
		Hospital:      "Aga Khan Hospital Kisumu",
		IsContractual: true,
		City:          "Kisumu",
		IsEnrolled:    true,
		AvailableTime: []time.Time{
			time.Now().Add(2 * time.Hour),  // Available in 2 hours
			time.Now().Add(24 * time.Hour), // Available tomorrow
		},
	}

	store.Doctors["doc-2"] = models.Doctor{
		ID:            "doc-2",
		Name:          "Dr. David Kamau",
		Specialty:     "Pediatric Surgeon",
		Hospital:      "Independent Clinic - Milimani",
		IsContractual: false,
		City:          "Kisumu",
		IsEnrolled:    true,
		AvailableTime: []time.Time{
			time.Now().Add(4 * time.Hour),
		},
	}

	return store
}

// GetAvailableDoctors filters doctors by city and enrollment status
func (ds *DataStore) GetAvailableDoctors(city string) []models.Doctor {
	var available []models.Doctor
	for _, doc := range ds.Doctors {
		if doc.City == city && doc.IsEnrolled && len(doc.AvailableTime) > 0 {
			available = append(available, doc)
		}
	}
	return available
}

// CreateAppointment saves a new booking
func (ds *DataStore) CreateAppointment(app models.Appointment) error {
	if _, exists := ds.Appointments[app.ID]; exists {
		return errors.New("appointment ID already exists")
	}
	ds.Appointments[app.ID] = app
	return nil
}

