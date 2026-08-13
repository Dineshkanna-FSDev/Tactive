package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("my_super_secret_tactive_key")

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
	IsBooked  bool   `json:"is_booked"`
}

type BookingRequest struct {
	SlotID string `json:"slot_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserBooking struct {
	BookingID string `json:"booking_id"`
	TurfName  string `json:"turf_name"`
	Location  string `json:"location"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing Token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Invalid Token", http.StatusUnauthorized)
			return
		}

		// Store user_id in the request context so our routes can access it
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v\n", err)
	}
	fmt.Println("✅ Successfully connected to PostgreSQL database!")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// --- PUBLIC ROUTES ---
	r.Post("/api/auth/login", func(w http.ResponseWriter, req *http.Request) {
		var creds LoginRequest
		if err := json.NewDecoder(req.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		var id, hash, name string
		err := dbPool.QueryRow(context.Background(), "SELECT id, name, password_hash FROM users WHERE email=$1", creds.Email).Scan(&id, &name, &hash)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(creds.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			UserID: id,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "name": name})
	})

	r.Get("/api/turfs", func(w http.ResponseWriter, req *http.Request) {
		rows, err := dbPool.Query(context.Background(), "SELECT id, name, location, price_per_hour FROM turfs")
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var turfs []Turf
		for rows.Next() {
			var t Turf
			if err := rows.Scan(&t.ID, &t.Name, &t.Location, &t.PricePerHour); err != nil {
				http.Error(w, "Error reading data", http.StatusInternalServerError)
				return
			}
			turfs = append(turfs, t)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(turfs)
	})

	r.Get("/api/slots", func(w http.ResponseWriter, req *http.Request) {
		query := `
			SELECT s.id, s.turf_id, s.start_time, s.end_time, 
			       CASE WHEN b.id IS NOT NULL THEN true ELSE false END as is_booked
			FROM slots s
			LEFT JOIN bookings b ON s.id = b.slot_id AND b.status = 'CONFIRMED'
		`
		rows, err := dbPool.Query(context.Background(), query)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var slots []Slot
		for rows.Next() {
			var s Slot
			var startTime, endTime time.Time
			if err := rows.Scan(&s.ID, &s.TurfID, &startTime, &endTime, &s.IsBooked); err != nil {
				http.Error(w, "Error reading data", http.StatusInternalServerError)
				return
			}
			s.StartTime = startTime.Format(time.RFC3339)
			s.EndTime = endTime.Format(time.RFC3339)
			slots = append(slots, s)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(slots)
	})

	// --- PROTECTED ROUTES (Requires JWT) ---
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Post("/api/bookings", func(w http.ResponseWriter, req *http.Request) {
			userID := req.Context().Value("user_id").(string) // Extracted securely from the JWT token!
			
			var b BookingRequest
			if err := json.NewDecoder(req.Body).Decode(&b); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}

			_, err := dbPool.Exec(context.Background(), 
				"INSERT INTO bookings (user_id, slot_id, status) VALUES ($1, $2, 'CONFIRMED')", 
				userID, b.SlotID)
			
			if err != nil {
				if strings.Contains(err.Error(), "23505") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(`{"error": "SLOT_ALREADY_BOOKED", "message": "Sorry, someone just booked this slot!"}`))
					return
				}
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"status": "SUCCESS", "message": "Booking Confirmed!"}`))
		})

		r.Get("/api/user/bookings", func(w http.ResponseWriter, req *http.Request) {
			userID := req.Context().Value("user_id").(string)
			
			query := `
				SELECT b.id, t.name, t.location, s.start_time, s.end_time, b.status
				FROM bookings b
				JOIN slots s ON b.slot_id = s.id
				JOIN turfs t ON s.turf_id = t.id
				WHERE b.user_id = $1
				ORDER BY s.start_time DESC
			`
			rows, err := dbPool.Query(context.Background(), query, userID)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var bookings []UserBooking
			for rows.Next() {
				var b UserBooking
				var startTime, endTime time.Time
				if err := rows.Scan(&b.BookingID, &b.TurfName, &b.Location, &startTime, &endTime, &b.Status); err != nil {
					http.Error(w, "Error reading data", http.StatusInternalServerError)
					return
				}
				b.StartTime = startTime.Format(time.RFC3339)
				b.EndTime = endTime.Format(time.RFC3339)
				bookings = append(bookings, b)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(bookings)
		})

		r.Post("/api/bookings/{id}/cancel", func(w http.ResponseWriter, req *http.Request) {
			bookingID := chi.URLParam(req, "id")
			userID := req.Context().Value("user_id").(string)

			// Update booking status to CANCELLED
			// BUG FIXED: Corrected 'WHRE' to 'WHERE'
			query := `UPDATE bookings SET status = 'CANCELLED' WHERE id = $1 AND user_id = $2`
			_, err := dbPool.Exec(context.Background(), query, bookingID, userID)
			if err != nil {
				http.Error(w, "Failed to process cancellation", http.StatusInternalServerError)
				return
			}
			
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "SUCCESS"}`))
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Starting Go server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}