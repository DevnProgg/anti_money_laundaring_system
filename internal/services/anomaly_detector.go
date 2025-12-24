package services

import (
	"AML/internal/models"
	"fmt"
	"math"
)

// DetectAmountAnomaly checks for anomalous transaction amounts.
func DetectAmountAnomaly(currentTx models.Transaction, history []models.Transaction) (isAnomaly bool, zScore float64, err error) {
	if len(history) < 10 {
		return false, 0, fmt.Errorf("insufficient transaction history for anomaly detection (requires at least 10 transactions)")
	}

	mean := calculateMean(history)
	stdDev := calculateStdDev(history, mean)

	if stdDev == 0 {
		// If all historical amounts are identical, any different amount is an anomaly.
		if currentTx.Amount != mean {
			return true, math.Inf(1), nil // Infinite z-score
		}
		return false, 0, nil
	}

	zScore = (currentTx.Amount - mean) / stdDev
	isAnomaly = math.Abs(zScore) > 3

	return isAnomaly, zScore, nil
}

// calculateMean calculates the mean of transaction amounts.
func calculateMean(transactions []models.Transaction) float64 {
	var sum float64
	for _, tx := range transactions {
		sum += tx.Amount
	}
	return sum / float64(len(transactions))
}

// calculateStdDev calculates the standard deviation of transaction amounts.
func calculateStdDev(transactions []models.Transaction, mean float64) float64 {
	if len(transactions) == 0 {
		return 0
	}

	var sumOfSquares float64
	for _, tx := range transactions {
		sumOfSquares += math.Pow(tx.Amount-mean, 2)
	}
	variance := sumOfSquares / float64(len(transactions))
	return math.Sqrt(variance)
}
