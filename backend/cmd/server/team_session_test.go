package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Helper to get a valid token
func getAuthToken(t *testing.T) string {
	baseURL := "http://localhost:8080"
	loginReq := map[string]string{
		"email":    "dinesh@test.com",
		"password": "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	return loginResp.Token
}

// Helper to get an available slot
func getAvailableSlot(t *testing.T, token string) string {
	baseURL := "http://localhost:8080"
	req, _ := http.NewRequest("GET", baseURL+"/api/slots", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get slots: %v", err)
	}
	defer resp.Body.Close()

	var slots []struct {
		ID       string `json:"id"`
		IsBooked bool   `json:"is_booked"`
	}
	json.NewDecoder(resp.Body).Decode(&slots)

	for _, s := range slots {
		if !s.IsBooked {
			return s.ID
		}
	}
	t.Fatalf("No available slots found for testing. Please reset DB.")
	return ""
}

func TestTeamSessionFeature(t *testing.T) {
	baseURL := "http://localhost:8080"
	token := getAuthToken(t)
	slotID := getAvailableSlot(t, token)

	// 1. Happy path: Create team session
	t.Run("HappyPath_CreateTeamSession", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"slot_id":       slotID,
			"target_amount": 1000.00,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", baseURL+"/api/team-sessions", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var respData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respData)

		if respData["session_id"] == nil || respData["session_id"] == "" {
			t.Errorf("Response missing session_id")
		}
		if respData["status"] != "PENDING" {
			t.Errorf("Expected status PENDING, got %v", respData["status"])
		}

		// Connect to DB directly for the test
		pool, err := pgxpool.New(context.Background(), "postgres://tactive_user:password123@localhost:5432/tactive_db")
		if err != nil {
			t.Fatalf("Test DB connection failed: %v", err)
		}
		defer pool.Close()

		var targetAmount float64
		var expiresAt time.Time
		err = pool.QueryRow(context.Background(), "SELECT target_amount, expires_at FROM team_sessions WHERE id = $1", respData["session_id"]).Scan(&targetAmount, &expiresAt)
		if err != nil {
			t.Errorf("Failed to find session in DB: %v", err)
		}
		if targetAmount != 1000.00 {
			t.Errorf("Expected target_amount 1000, got %f", targetAmount)
		}
		
		expectedExpire := time.Now().Add(15 * time.Minute)
		diff := expiresAt.Sub(expectedExpire).Abs()
		if diff > 10*time.Second {
			t.Errorf("Expiration time %v is not approximately 15 minutes from now (%v)", expiresAt, expectedExpire)
		}
	})

	// 7. Creating another active team session for the same slot -> 409
	t.Run("DoubleBooking_TeamSession", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"slot_id":       slotID,
			"target_amount": 500.00,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", baseURL+"/api/team-sessions", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp.StatusCode)
		}
	})

	// 8. Creating a normal booking for a slot held by an active team session -> 409
	t.Run("DoubleBooking_NormalBooking_On_TeamSession", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"slot_id": slotID,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", baseURL+"/api/bookings", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp.StatusCode)
		}
	})

	// 9. Creating a team session for an already CONFIRMED booking -> 409
	t.Run("TeamSession_On_ConfirmedBooking", func(t *testing.T) {
		confirmedSlotID := getAvailableSlot(t, token)
		
		// Create normal booking
		bookReq := map[string]interface{}{"slot_id": confirmedSlotID}
		bookBytes, _ := json.Marshal(bookReq)
		req1, _ := http.NewRequest("POST", baseURL+"/api/bookings", bytes.NewBuffer(bookBytes))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		http.DefaultClient.Do(req1)

		// Try team session
		tsReq := map[string]interface{}{
			"slot_id":       confirmedSlotID,
			"target_amount": 1000.00,
		}
		tsBytes, _ := json.Marshal(tsReq)
		req2, _ := http.NewRequest("POST", baseURL+"/api/team-sessions", bytes.NewBuffer(tsBytes))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		
		resp, err := http.DefaultClient.Do(req2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp.StatusCode)
		}
	})

	// 10. Invalid/missing slot_id is rejected appropriately
	t.Run("InputValidation_MissingSlotID", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"target_amount": 1000.00,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", baseURL+"/api/team-sessions", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request for missing slot_id, got %d", resp.StatusCode)
		}
	})

	// 11. target_amount <= 0 is rejected appropriately
	t.Run("InputValidation_NegativeTargetAmount", func(t *testing.T) {
		anotherSlotID := getAvailableSlot(t, token)
		reqBody := map[string]interface{}{
			"slot_id":       anotherSlotID,
			"target_amount": -50.00,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", baseURL+"/api/team-sessions", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request for negative target_amount, got %d", resp.StatusCode)
		}
	})
}
