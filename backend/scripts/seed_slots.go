package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables (navigate up since script is in /scripts)
	godotenv.Load("../.env")
	godotenv.Load("../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback for local testing
		dbURL = "postgres://tactive_user:password123@localhost:5432/tactive_db"
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	fmt.Println("🧹 Clearing old slots and bookings...")
	_, err = dbPool.Exec(context.Background(), "DELETE FROM slots")
	if err != nil {
		log.Fatalf("Failed to clear slots: %v", err)
	}

	fmt.Println("🔍 Fetching all turfs...")
	rows, err := dbPool.Query(context.Background(), "SELECT id FROM turfs")
	if err != nil {
		log.Fatalf("Failed to fetch turfs: %v", err)
	}
	
	var turfIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		turfIDs = append(turfIDs, id)
	}
	rows.Close()

	fmt.Printf("📅 Generating 7 days of slots for %d turfs (05:00 to 00:00)...\n", len(turfIDs))
	
	now := time.Now()
	// Start from today at 00:00:00
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	inserted := 0

	for _, turfID := range turfIDs {
		for day := 0; day < 7; day++ {
			// Generate slots from 05:00 to 23:00 (which is 19 hours)
			for hour := 5; hour <= 23; hour++ {
				startTime := startOfDay.AddDate(0, 0, day).Add(time.Duration(hour) * time.Hour)
				
				// Add 1 hour to get the correct end time
				endTime := startTime.Add(1 * time.Hour) 

				_, err := dbPool.Exec(context.Background(), 
					"INSERT INTO slots (turf_id, start_time, end_time) VALUES ($1, $2, $3)",
					turfID, startTime, endTime)
					
				if err != nil {
					log.Fatalf("\n❌ [RED RUN TRIGGERED] Database Error:\n%v\nTake a screenshot for your documentation!", err)
				}
				inserted++
			}
		}
	}

	fmt.Printf("✅ Successfully inserted %d slots!\n", inserted)
}
