package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	baseURL := "http://localhost:8080"

	// 1. Login
	loginReq := map[string]string{"email": "dinesh@test.com", "password": "password123"}
	loginBody, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		fmt.Println("Login failed:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		fmt.Println("Login non-200:", string(b))
		return
	}

	var loginResp struct{ Token string }
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Token

	// 2. Fetch slots
	resp, err = http.Get(baseURL + "/api/slots")
	if err != nil {
		fmt.Println("Fetch slots failed:", err)
		return
	}
	defer resp.Body.Close()

	var slots []struct {
		ID       string `json:"id"`
		IsBooked bool   `json:"is_booked"`
	}
	json.NewDecoder(resp.Body).Decode(&slots)

	var targetSlot string
	for _, s := range slots {
		if !s.IsBooked {
			targetSlot = s.ID
			break
		}
	}

	if targetSlot == "" {
		fmt.Println("No available slots found.")
		return
	}

	// 3. Book slot
	bookReq := map[string]string{"slot_id": targetSlot}
	bookBody, _ := json.Marshal(bookReq)

	req, _ := http.NewRequest("POST", baseURL+"/api/bookings", bytes.NewBuffer(bookBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Booking failed:", err)
		return
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	fmt.Printf("Booking status: %d, response: %s\n", resp.StatusCode, string(b))
}
