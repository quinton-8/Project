package worker

import (
	"fmt"
	"time"

	"github.com/quinton-8/project/internal/database"
)

// StartReminderJob runs infinitely in the background
func StartReminderJob(store *database.DataStore) {
	// Create a ticker that fires exactly once per minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	fmt.Println("[WORKER] Background SMS reminder job started...")

	// This infinite loop listens for the ticker
	for {
		<-ticker.C // Pauses here until 1 minute passes
		checkAndSendReminders(store)
	}
}

func checkAndSendReminders(store *database.DataStore) {
	now := time.Now()

	for id, app := range store.Appointments {
		// Only look at confirmed appointments that haven't received an SMS yet
		if app.Status == "confirmed" && !app.ReminderSent {
			// If the current time has passed the target ReminderTime
			if now.After(app.ReminderTime) || now.Equal(app.ReminderTime) {

				// TODO: Integrate Africa's Talking or Twilio API here
				fmt.Printf("\n[SMS TRIGGERED] -> Sending to Client %s: Time to leave for Doctor %s!\n", app.ClientID, app.DoctorID)

				// Mark as sent so we don't spam the user on the next tick
				app.ReminderSent = true
				store.UpdateAppointment(app)
			}
		}
	}
}
