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
	Hospitals    map[string]models.Hospital
}

// NewDataStore initializes the store with some hardcoded seed data
func NewDataStore() *DataStore {
	store := &DataStore{
		Doctors:      make(map[string]models.Doctor),
		Clients:      make(map[string]models.Client),
		Appointments: make(map[string]models.Appointment),
		Hospitals:    make(map[string]models.Hospital),
	}

	// Seed Data: Hospitals in Kisumu
	store.Hospitals["hosp-1"] = models.Hospital{
		ID:   "hosp-1",
		Name: "Aga Khan Hospital Kisumu",
		Lat:  -0.1035,
		Lng:  34.7550,
	}
	store.Hospitals["hosp-2"] = models.Hospital{
		ID:   "hosp-2",
		Name: "Jaramogi Oginga Odinga Teaching and Referral Hospital",
		Lat:  -0.0900,
		Lng:  34.7700,
	}
	store.Hospitals["hosp-3"] = models.Hospital{
		ID:   "hosp-3",
		Name: "Milimani Independent Clinic",
		Lat:  -0.1100,
		Lng:  34.7500,
	}

// Seed Data: Contractual and Independent Doctors in the target region
	store.Doctors["doc-1"] = models.Doctor{
		ID:             "doc-1",
		Name:           "Dr. Sarah Ochieng",
		Specialty:      "Pediatrician",
		Hospital:       "Aga Khan Hospital Kisumu", // Major hospital
		IsContractual:  true,
		City:           "Kisumu",
		IsEnrolled:     true,
		AvailableTime:  []time.Time{time.Now().Add(2 * time.Hour)},
		Lat:            -0.1035,
		Lng:            34.7550,
		CurrentQueue:   12, // Crowded: 12 patients waiting
		AvgConsultTime: 15, // 15 mins per patient = 180 minutes wait time
	}

	store.Doctors["doc-2"] = models.Doctor{
		ID:             "doc-2",
		Name:           "Dr. David Kamau",
		Specialty:      "Pediatric Surgeon",
		Hospital:       "Independent Clinic - Milimani", // Independent
		IsContractual:  false,
		City:           "Kisumu",
		IsEnrolled:     true,
		AvailableTime:  []time.Time{time.Now().Add(1 * time.Hour)},
		Lat:            -0.1100,
		Lng:            34.7500,
		CurrentQueue:   1,  // Nearly empty: 1 patient waiting
		AvgConsultTime: 20, // 20 mins per patient = 20 minutes wait time
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

// GetAppointment retrieves an appointment by its ID
func (ds *DataStore) GetAppointment(id string) (models.Appointment, error) {
	app, exists := ds.Appointments[id]
	if !exists {
		return models.Appointment{}, errors.New("appointment not found")
	}
	return app, nil
}

// UpdateAppointment replaces an existing appointment with updated data
func (ds *DataStore) UpdateAppointment(app models.Appointment) error {
	if _, exists := ds.Appointments[app.ID]; !exists {
		return errors.New("appointment not found")
	}
	ds.Appointments[app.ID] = app
	return nil
}