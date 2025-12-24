package services

import (
	"sort"
	"time"

	"AML/internal/models"
)

// DetectStructuring identifies a pattern of transactions indicative of smurfing.
func DetectStructuring(accountID string, transactions []models.Transaction, timeWindow time.Duration, thresholdLow float64, thresholdHigh float64, minCount int) (detected bool, matchingTxs []models.Transaction) {
	var candidates []models.Transaction
	now := time.Now()
	windowStart := now.Add(-timeWindow)

	for _, tx := range transactions {
		if tx.AccountID == accountID &&
			tx.Timestamp.After(windowStart) &&
			tx.Timestamp.Before(now) &&
			tx.Amount >= thresholdLow &&
			tx.Amount <= thresholdHigh {
			candidates = append(candidates, tx)
		}
	}

	if len(candidates) >= minCount {
		// Sort by timestamp descending
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Timestamp.After(candidates[j].Timestamp)
		})
		return true, candidates
	}

	return false, nil
}
