package main

import (
	"testing"
)

// TestCancellationQuery is a simple unit test to verify the SQL syntax for the cancel feature.
// It is intentionally designed to fail to demonstrate the CI/CD Red Run.
func TestCancellationQuery(t *testing.T) {
	// The expected correct SQL query
	expectedQuery := `UPDATE bookings SET status = 'CANCELLED' WHERE id = $1 AND user_id = $2`
	
	// The actual buggy query currently in our main.go code
	actualQuery := `UPDATE bookings SET status = 'CANCELLED' WHRE id = $1 AND user_id = $2`

	if actualQuery != expectedQuery {
		t.Errorf("\n\n❌ [CI/CD RED RUN TRIGGERED]\nSQL Syntax Error Detected in Cancellation Route!\nExpected: %s\nActual:   %s\n\nTake a screenshot of this GitHub Actions failure!", expectedQuery, actualQuery)
	}
}
