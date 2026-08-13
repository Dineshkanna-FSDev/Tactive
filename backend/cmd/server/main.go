package main
    
    import (
        "context"
        "fmt"
        "log"
        "net/http"
        "os"
		"strings"
    
        "github.com/go-chi/chi/v5"
        "github.com/go-chi/chi/v5/middleware"
        "github.com/jackc/pgx/v5/pgxpool"
        "github.com/joho/godotenv"

        "encoding/json"
    )

    type Turf struct {
        ID           string  `json:"id"`
        Name         string  `json:"name"`
        Location     string  `json:"location"`
        PricePerHour float64 `json:"price_per_hour"`
    }

	type Slot struct {
        ID        string `json:"id"`
        TurfID    string `json:"turf_id"`
        StartTime string `json:"start_time"`
        EndTime   string `json:"end_time"`
    }

    type BookingRequest struct {
        UserID string `json:"user_id"`
        SlotID string `json:"slot_id"`
    }
    
    
    func main() {
        // 1. Load environment variables from the .env file
        err := godotenv.Load()
        if err != nil {
            log.Println("No .env file found, relying on system environment variables")
        }
    
        // 2. Connect to PostgreSQL
        dbURL := os.Getenv("DATABASE_URL")
        if dbURL == "" {
            log.Fatal("DATABASE_URL environment variable is required")
        }
    
        // Create a connection pool to the database
        dbPool, err := pgxpool.New(context.Background(), dbURL)
        if err != nil {
            log.Fatalf("Unable to connect to database: %v\n", err)
        }
        defer dbPool.Close() // Ensures we close the connection when the server stops
    
        // Verify the connection actually works by pinging the database
        if err := dbPool.Ping(context.Background()); err != nil {
            log.Fatalf("Failed to ping database: %v\n", err)
        }
        fmt.Println("✅ Successfully connected to PostgreSQL database!")
    
        // 3. Set up the Router
        r := chi.NewRouter()
        r.Use(middleware.Logger)
        r.Use(middleware.Recoverer)
    
        r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("Server is healthy and database is connected!"))
        })
    


		r.Get("/api/turfs", func(w http.ResponseWriter, req *http.Request) {
            // 1. Query the database
            rows, err := dbPool.Query(context.Background(), "SELECT id, name, location, price_per_hour FROM turfs")
            if err != nil {
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
            }
            defer rows.Close()
    
            // 2. Loop through the results and put them in a slice (array)
            var turfs []Turf
            for rows.Next() {
                var t Turf
                if err := rows.Scan(&t.ID, &t.Name, &t.Location, &t.PricePerHour); err != nil {
                    http.Error(w, "Error reading data", http.StatusInternalServerError)
                    return
                }
                turfs = append(turfs, t)
            }

            // 3. Send the array back to the frontend as JSON!
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(turfs)
        })

		r.Post("/api/bookings", func(w http.ResponseWriter, req *http.Request) {
            var b BookingRequest
            if err := json.NewDecoder(req.Body).Decode(&b); err != nil {
                http.Error(w, "Invalid JSON body", http.StatusBadRequest)
                return
            }
    
            // Try to insert the booking into the database
            _, err := dbPool.Exec(context.Background(), 
                "INSERT INTO bookings (user_id, slot_id, status) VALUES ($1, $2, 'CONFIRMED')", 
                b.UserID, b.SlotID)
            
            if err != nil {
                // This is where the magic happens! 
                // If Postgres blocks it because of our Unique Index, it throws error code '23505'
                if strings.Contains(err.Error(), "23505") {
                    w.Header().Set("Content-Type", "application/json")
                    w.WriteHeader(http.StatusConflict) // 409 Conflict
                    w.Write([]byte(`{"error": "SLOT_ALREADY_BOOKED", "message": "Sorry, someone just booked this slot!"}`))
                    return
                }
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
            }
    
            w.WriteHeader(http.StatusCreated) // 201 Created
            w.Write([]byte(`{"status": "SUCCESS", "message": "Booking Confirmed!"}`))
        })

		 r.Get("/api/slots", func(w http.ResponseWriter, req *http.Request) {
            rows, err := dbPool.Query(context.Background(), "SELECT id, turf_id, start_time, end_time FROM slots")
            if err != nil {
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
            }
            defer rows.Close()
    
            var slots []Slot
            for rows.Next() {
                var s Slot
                if err := rows.Scan(&s.ID, &s.TurfID, &s.StartTime, &s.EndTime); err != nil {
                    http.Error(w, "Error reading data", http.StatusInternalServerError)
                    return
                }
                slots = append(slots, s)
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(slots)
        })

		
		// 4. Start the Server
        port := os.Getenv("PORT")
        if port == "" {
            port = "8080"
        }

        fmt.Printf("🚀 Starting Go server on port %s...\n", port)
        if err := http.ListenAndServe(":"+port, r); err != nil {
            log.Fatalf("Server failed to start: %v", err)
        }
    }