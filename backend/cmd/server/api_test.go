package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// TestCancellationRedRun runs an integration test against the local API server.
// It logs in, finds a booking, and attempts to cancel it.
func TestCancellationRedRun(t *testing.T) {
	baseURL := "http://localhost:8080"
	
	// 1. Authenticate to get JWT token
	loginReq := map[string]string{
		"email":    "dinesh@test.com",
		"password": "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	
	t.Log("Attempting to login...")
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("FATAL: Failed to connect to server: %v (Did you forget to start the Go server in another terminal?)", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("FATAL: Expected login status 200, got %d", resp.StatusCode)
	}
	
	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	
	if loginResp.Token == "" {
		t.Fatalf("FATAL: Login succeeded but failed to parse JWT token")
	}
	
	// 2. Fetch User Bookings to find a target booking to cancel
	t.Log("Fetching user bookings...")
	req, _ := http.NewRequest("GET", baseURL+"/api/user/bookings", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("FATAL: Failed to fetch bookings: %v", err)
	}
	defer resp.Body.Close()
	
	var bookings []struct {
		BookingID string `json:"booking_id"`
		Status    string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&bookings)
	
	if len(bookings) == 0 {
		t.Fatalf("FATAL: No bookings found for this user. Please go to the React UI and create at least 1 booking before running this test!")
	}
	
	targetBookingID := bookings[0].BookingID
	t.Logf("Found active booking to cancel: %s", targetBookingID)
	
	// 3. Request Cancellation (THIS IS WHERE THE SQL BUG WILL TRIGGER)
	t.Log("Requesting cancellation via API...")
	cancelReq, _ := http.NewRequest("POST", baseURL+"/api/bookings/"+targetBookingID+"/cancel", nil)
	cancelReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	
	resp, err = http.DefaultClient.Do(cancelReq)
	if err != nil {
		t.Fatalf("FATAL: Failed to send cancel request: %v", err)
	}
	defer resp.Body.Close()
	
	// The test strictly expects a 200 OK status code.
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("\n\n❌ [RED RUN TRIGGERED] INTENTIONAL BUG DETECTED!\nExpected HTTP Status 200 OK, but got %d Internal Server Error.\nBackend Error Details: %s\n\nTake a screenshot of this output for your AI Change Loop documentation!", resp.StatusCode, string(bodyBytes))
	}
	
	t.Log("\n\n✅ [GREEN RUN SUCCESS] The Cancellation API processed the request perfectly!")
}
