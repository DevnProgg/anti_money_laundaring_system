package services

import (
	"AML/internal/config"
	"AML/internal/models"
	"time"
)

// RuleViolation represents a rule that has been violated.
type RuleViolation struct {
	RuleID         string  `json:"rule_id"`
	ActualValue    float64 `json:"actual_value"`
	ThresholdValue float64 `json:"threshold_value"`
}

// EvaluateRules checks a transaction against a set of rules.
func EvaluateRules(tx models.Transaction, rules []config.Rule, history []models.Transaction) ([]RuleViolation, error) {
	if rules == nil {
		return nil, nil // No rules to evaluate
	}

	var violations []RuleViolation

	for _, rule := range rules {
		if !rule.Enabled {
			continue // Skip disabled rules
		}

		timeWindow, err := rule.GetTimeWindow()
		if err != nil {
			return nil, err
		}

		switch rule.RuleID {
		case "single_transaction_exceeds_10000":
			if tx.Amount > rule.ThresholdValue {
				violations = append(violations, RuleViolation{
					RuleID:         rule.RuleID,
					ActualValue:    tx.Amount,
					ThresholdValue: rule.ThresholdValue,
				})
			}
		case "daily_cumulative_exceeds_50000":
			transactionsInWindow := GetTransactionsInWindow(history, timeWindow)
			var totalAmount float64
			for _, t := range transactionsInWindow {
				totalAmount += t.Amount
			}
			totalAmount += tx.Amount // Include current transaction

			if totalAmount > rule.ThresholdValue {
				violations = append(violations, RuleViolation{
					RuleID:         rule.RuleID,
					ActualValue:    totalAmount,
					ThresholdValue: rule.ThresholdValue,
				})
			}
		case "more_than_5_transactions_in_1_hour":
			transactionsInWindow := GetTransactionsInWindow(history, timeWindow)
			transactionCount := len(transactionsInWindow) + 1 // Include current transaction

			if float64(transactionCount) > rule.ThresholdValue {
				violations = append(violations, RuleViolation{
					RuleID:         rule.RuleID,
					ActualValue:    float64(transactionCount),
					ThresholdValue: rule.ThresholdValue,
				})
			}
		default:
			// Optionally handle unknown rule IDs
		}
	}

	return violations, nil
}

// GetTransactionsInWindow filters transactions that fall within a given time window.
func GetTransactionsInWindow(history []models.Transaction, window time.Duration) []models.Transaction {
	if history == nil {
		return nil
	}

	var transactionsInWindow []models.Transaction
	now := time.Now()
	windowStart := now.Add(-window)

	for _, tx := range history {
		if tx.Timestamp.After(windowStart) && tx.Timestamp.Before(now) {
			transactionsInWindow = append(transactionsInWindow, tx)
		}
	}
	return transactionsInWindow
}
