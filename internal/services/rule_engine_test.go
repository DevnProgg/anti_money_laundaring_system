package services

import (
	"AML/internal/config"
	"AML/internal/models"
	"testing"
	"time"
)

func TestEvaluateRules(t *testing.T) {
	rules := []config.Rule{
		{
			RuleID:         "single_transaction_exceeds_10000",
			Name:           "Single Transaction Exceeds $10,000",
			ThresholdValue: 10000.00,
			TimeWindow:     "0h",
			Enabled:        true,
		},
		{
			RuleID:         "daily_cumulative_exceeds_50000",
			Name:           "Daily Cumulative Transactions Exceeds $50,000",
			ThresholdValue: 50000.00,
			TimeWindow:     "24h",
			Enabled:        true,
		},
		{
			RuleID:         "more_than_5_transactions_in_1_hour",
			Name:           "More Than 5 Transactions in 1 Hour",
			ThresholdValue: 5.0,
			TimeWindow:     "1h",
			Enabled:        true,
		},
	}

	// Test Case 1: Single transaction exceeds threshold
	t.Run("single_transaction_exceeds", func(t *testing.T) {
		tx := models.Transaction{Amount: 15000.00, Timestamp: time.Now()}
		violations, err := EvaluateRules(tx, rules, nil)
		if err != nil {
			t.Fatalf("EvaluateRules failed: %v", err)
		}
		if len(violations) != 1 || violations[0].RuleID != "single_transaction_exceeds_10000" {
			t.Errorf("Expected 1 violation for single_transaction_exceeds_10000, got %d", len(violations))
		}
	})

	// Test Case 2: Daily cumulative transactions exceed threshold
	t.Run("daily_cumulative_exceeds", func(t *testing.T) {
		history := []models.Transaction{
			{Amount: 20000.00, Timestamp: time.Now().Add(-2 * time.Hour)},
			{Amount: 20000.00, Timestamp: time.Now().Add(-1 * time.Hour)},
		}
		tx := models.Transaction{Amount: 15000.00, Timestamp: time.Now()}
		violations, err := EvaluateRules(tx, rules, history)
		if err != nil {
			t.Fatalf("EvaluateRules failed: %v", err)
		}

		var found bool
		for _, v := range violations {
			if v.RuleID == "daily_cumulative_exceeds_50000" {
				found = true
				if v.ActualValue != 55000.00 {
					t.Errorf("Expected actual value of 55000.00, got %f", v.ActualValue)
				}
			}
		}
		if !found {
			t.Errorf("Expected violation for daily_cumulative_exceeds_50000")
		}
	})

	// Test Case 3: More than 5 transactions in 1 hour
	t.Run("too_many_transactions_in_hour", func(t *testing.T) {
		history := []models.Transaction{
			{Amount: 100.00, Timestamp: time.Now().Add(-10 * time.Minute)},
			{Amount: 200.00, Timestamp: time.Now().Add(-20 * time.Minute)},
			{Amount: 300.00, Timestamp: time.Now().Add(-30 * time.Minute)},
			{Amount: 400.00, Timestamp: time.Now().Add(-40 * time.Minute)},
			{Amount: 500.00, Timestamp: time.Now().Add(-50 * time.Minute)},
		}
		tx := models.Transaction{Amount: 600.00, Timestamp: time.Now()}
		violations, err := EvaluateRules(tx, rules, history)
		if err != nil {
			t.Fatalf("EvaluateRules failed: %v", err)
		}

		var found bool
		for _, v := range violations {
			if v.RuleID == "more_than_5_transactions_in_1_hour" {
				found = true
				if v.ActualValue != 6 {
					t.Errorf("Expected actual value of 6, got %f", v.ActualValue)
				}
			}
		}
		if !found {
			t.Errorf("Expected violation for more_than_5_transactions_in_1_hour")
		}
	})

	// Test Case 4: No violations
	t.Run("no_violations", func(t *testing.T) {
		history := []models.Transaction{
			{Amount: 1000.00, Timestamp: time.Now().Add(-2 * time.Hour)},
		}
		tx := models.Transaction{Amount: 1000.00, Timestamp: time.Now()}
		violations, err := EvaluateRules(tx, rules, history)
		if err != nil {
			t.Fatalf("EvaluateRules failed: %v", err)
		}
		if len(violations) != 0 {
			t.Errorf("Expected 0 violations, got %d", len(violations))
		}
	})
}
