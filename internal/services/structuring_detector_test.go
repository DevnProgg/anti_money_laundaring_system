package services

import (
	"testing"
	"time"

	"AML/internal/models"
)

func TestDetectStructuring(t *testing.T) {
	accountID := "acc-123"
	timeWindow := 24 * time.Hour
	thresholdLow := 8000.00
	thresholdHigh := 9999.00
	minCount := 3

	// Test Case 1: Structuring pattern detected
	t.Run("structuring_pattern_detected", func(t *testing.T) {
		transactions := []models.Transaction{
			{AccountID: accountID, Amount: 8500.00, Timestamp: time.Now().Add(-1 * time.Hour)},
			{AccountID: accountID, Amount: 9200.00, Timestamp: time.Now().Add(-2 * time.Hour)},
			{AccountID: accountID, Amount: 9800.00, Timestamp: time.Now().Add(-3 * time.Hour)},
			{AccountID: accountID, Amount: 7000.00, Timestamp: time.Now().Add(-4 * time.Hour)},
			{AccountID: accountID, Amount: 15000.00, Timestamp: time.Now().Add(-5 * time.Hour)},
			{AccountID: "acc-456", Amount: 9000.00, Timestamp: time.Now().Add(-6 * time.Hour)}, // Different account
		}

		detected, matchingTxs := DetectStructuring(accountID, transactions, timeWindow, thresholdLow, thresholdHigh, minCount)

		if !detected {
			t.Errorf("Expected structuring pattern to be detected, but it was not")
		}
		if len(matchingTxs) != 3 {
			t.Errorf("Expected 3 matching transactions, but got %d", len(matchingTxs))
		}
		// Check if sorted correctly
		if matchingTxs[0].Amount != 8500.00 {
			t.Errorf("Expected transactions to be sorted by timestamp descending")
		}
	})

	// Test Case 2: Transactions spread over 30 hours (should not detect)
	t.Run("transactions_outside_timewindow", func(t *testing.T) {
		transactions := []models.Transaction{
			{AccountID: accountID, Amount: 8500.00, Timestamp: time.Now().Add(-1 * time.Hour)},
			{AccountID: accountID, Amount: 9200.00, Timestamp: time.Now().Add(-15 * time.Hour)},
			{AccountID: accountID, Amount: 9800.00, Timestamp: time.Now().Add(-30 * time.Hour)},
		}

		detected, _ := DetectStructuring(accountID, transactions, timeWindow, thresholdLow, thresholdHigh, minCount)

		if detected {
			t.Errorf("Expected structuring pattern to not be detected, but it was")
		}
	})

	// Test Case 3: Not enough transactions in range
	t.Run("not_enough_transactions_in_range", func(t *testing.T) {
		transactions := []models.Transaction{
			{AccountID: accountID, Amount: 8500.00, Timestamp: time.Now().Add(-1 * time.Hour)},
			{AccountID: accountID, Amount: 9200.00, Timestamp: time.Now().Add(-2 * time.Hour)},
			{AccountID: accountID, Amount: 7000.00, Timestamp: time.Now().Add(-3 * time.Hour)},
		}

		detected, _ := DetectStructuring(accountID, transactions, timeWindow, thresholdLow, thresholdHigh, minCount)

		if detected {
			t.Errorf("Expected structuring pattern to not be detected, but it was")
		}
	})
}
