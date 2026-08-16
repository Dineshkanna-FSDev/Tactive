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

	loginReq := map[string]string{"email": "dinesh@test.com", "password": "password123"}
	loginBody, _ := json.Marshal(loginReq)
	resp, _ := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(loginBody))
	var loginResp struct{ Token string }
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Token

	resp, _ = http.Get(baseURL + "/api/slots")
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
	fmt.Println("Slot:", targetSlot)

	teamReq := map[string]interface{}{"slot_id": targetSlot, "target_amount": 1200}
	teamBody, _ := json.Marshal(teamReq)

	req, _ := http.NewRequest("POST", baseURL+"/api/team-sessions", bytes.NewBuffer(teamBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, _ = http.DefaultClient.Do(req)
	b, _ := io.ReadAll(resp.Body)
	fmt.Printf("Team Session status: %d, response: %s\n", resp.StatusCode, string(b))
}
