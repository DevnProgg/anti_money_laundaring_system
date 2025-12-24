package services

import (
	"AML/internal/models"
	"testing"
	"time"
)

func TestDetectAmountAnomaly(t *testing.T) {
	// Test Case 1: Normal transaction
	t.Run("normal_transaction", func(t *testing.T) {
		history := make([]models.Transaction, 10)
		for i := 0; i < 10; i++ {
			history[i] = models.Transaction{Amount: 100.00, Timestamp: time.Now()}
		}
		currentTx := models.Transaction{Amount: 110.00, Timestamp: time.Now()}
		isAnomaly, zScore, err := DetectAmountAnomaly(currentTx, history)
		if err != nil {
			t.Fatalf("DetectAmountAnomaly failed: %v", err)
		}
		if isAnomaly {
			t.Errorf("Expected normal transaction, but got anomaly (z-score: %f)", zScore)
		}
	})

	// Test Case 2: Anomalous transaction
	t.Run("anomalous_transaction", func(t *testing.T) {
		history := make([]models.Transaction, 10)
		for i := 0; i < 10; i++ {
			history[i] = models.Transaction{Amount: 100.00, Timestamp: time.Now()}
		}
		currentTx := models.Transaction{Amount: 500.00, Timestamp: time.Now()}
		isAnomaly, zScore, err := DetectAmountAnomaly(currentTx, history)
		if err != nil {
			t.Fatalf("DetectAmountAnomaly failed: %v", err)
		}
		if !isAnomaly {
			t.Errorf("Expected anomalous transaction, but got normal (z-score: %f)", zScore)
		}
	})

	// Test Case 3: Insufficient history
	t.Run("insufficient_history", func(t *testing.T) {
		history := make([]models.Transaction, 5)
		currentTx := models.Transaction{Amount: 100.00, Timestamp: time.Now()}
		_, _, err := DetectAmountAnomaly(currentTx, history)
		if err == nil {
			t.Errorf("Expected error for insufficient history, but got nil")
		}
	})

	// Test Case 4: Zero variance
	t.Run("zero_variance", func(t *testing.T) {
		history := make([]models.Transaction, 10)
		for i := 0; i < 10; i++ {
			history[i] = models.Transaction{Amount: 100.00, Timestamp: time.Now()}
		}

		// Sub-case 4a: Current transaction is same as history
		currentTxSame := models.Transaction{Amount: 100.00, Timestamp: time.Now()}
		isAnomalySame, _, err := DetectAmountAnomaly(currentTxSame, history)
		if err != nil {
			t.Fatalf("DetectAmountAnomaly failed: %v", err)
		}
		if isAnomalySame {
			t.Errorf("Expected normal transaction for zero variance case (same amount), but got anomaly")
		}

		// Sub-case 4b: Current transaction is different from history
		currentTxDiff := models.Transaction{Amount: 200.00, Timestamp: time.Now()}
		isAnomalyDiff, _, err := DetectAmountAnomaly(currentTxDiff, history)
		if err != nil {
			t.Fatalf("DetectAmountAnomaly failed: %v", err)
		}
		if !isAnomalyDiff {
			t.Errorf("Expected anomaly for zero variance case (different amount), but got normal")
		}
	})
}
