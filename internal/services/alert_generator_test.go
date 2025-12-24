package services

import (
	"AML/internal/models"
	"testing"
)

func TestGenerateAlert(t *testing.T) {
	tx := models.Transaction{
		TransactionID: "tx-123",
		Amount:        50000.00,
	}

	// Test Case 1: Medium priority alert
	t.Run("medium_priority_alert", func(t *testing.T) {
		alert, err := GenerateAlert(tx, AlertTypeThresholdViolation, nil)
		if err != nil {
			t.Fatalf("GenerateAlert failed: %v", err)
		}

		if alert.Priority != models.Medium {
			t.Errorf("Expected priority Medium, got %s", alert.Priority)
		}
		expectedScore := 50*0.8 + ((50000.0/1000000.0)*20)*0.2
		if alert.Score != expectedScore {
			t.Errorf("Expected score %f, got %f", expectedScore, alert.Score)
		}
	})

	// Test Case 2: High priority alert
	t.Run("high_priority_alert", func(t *testing.T) {
		alert, err := GenerateAlert(tx, AlertTypeAnomalyDetected, nil)
		if err != nil {
			t.Fatalf("GenerateAlert failed: %v", err)
		}

		if alert.Priority != models.High {
			t.Errorf("Expected priority High, got %s", alert.Priority)
		}
		expectedScore := 75*0.8 + ((50000.0/1000000.0)*20)*0.2
		if alert.Score != expectedScore {
			t.Errorf("Expected score %f, got %f", expectedScore, alert.Score)
		}
	})

	// Test Case 3: Critical priority alert
	t.Run("critical_priority_alert", func(t *testing.T) {
		alert, err := GenerateAlert(tx, AlertTypeStructuringPattern, nil)
		if err != nil {
			t.Fatalf("GenerateAlert failed: %v", err)
		}

		if alert.Priority != models.Critical {
			t.Errorf("Expected priority Critical, got %s", alert.Priority)
		}
		expectedScore := 100*0.8 + ((50000.0/1000000.0)*20)*0.2
		if alert.Score != expectedScore {
			t.Errorf("Expected score %f, got %f", expectedScore, alert.Score)
		}
	})

	// Test Case 4: Invalid alert type
	t.Run("invalid_alert_type", func(t *testing.T) {
		_, err := GenerateAlert(tx, "INVALID_TYPE", nil)
		if err == nil {
			t.Errorf("Expected error for invalid alert type, but got nil")
		}
	})
}
