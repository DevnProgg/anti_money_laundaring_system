package services

import (
	"database/sql"
	"encoding/json" // Added
	"fmt"           // Added
	"strings"
	"time"

	"AML/internal/models"
)

// ensure time is imported and used to avoid linter errors when only type declarations use it.
var _ = time.Now

// GenerateSARData aggregates data from multiple alerts into a single Suspicious Activity Report (SAR).
func GenerateSARData(alertIDs []string, db *sql.DB) (*models.SARReport, error) {
	if len(alertIDs) == 0 {
		// Return nil, nil if no alert IDs are provided, as there's nothing to report.
		return nil, nil
	}

	queryArgs := make([]interface{}, len(alertIDs))
	for i, id := range alertIDs {
		queryArgs[i] = id
	}

	// 1. Fetch Alerts based on alertIDs
	alertQuery := `
		SELECT id, alert_type, transaction_id, rule_details
		FROM alerts
		WHERE id IN (?` + strings.Repeat(",?", len(alertIDs)-1) + `)
	`
	alertRows, err := db.Query(alertQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer alertRows.Close()

	var alerts []models.Alert
	var allRelatedTxIDs []string
	alertIDToDetails := make(map[string]models.Alert)

	for alertRows.Next() {
		var alert models.Alert
		err := alertRows.Scan(&alert.ID, &alert.AlertType, &alert.TransactionID, &alert.RuleDetails)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %w", err)
		}

		// `alert.RuleDetails` is now directly populated by its Scan method.
		// We can directly use alert.RuleDetails now.
		ruleDetailsMap := alert.RuleDetails

		if rawMatchingTxs, ok := ruleDetailsMap["matching_transactions"]; ok {
			if list, isList := rawMatchingTxs.([]interface{}); isList {
				for _, item := range list {
					txBytes, err := json.Marshal(item)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal transaction item from rule_details for alert %s: %w", alert.ID, err)
					}
					var tx models.Transaction
					err = json.Unmarshal(txBytes, &tx)
					if err != nil {
						return nil, fmt.Errorf("failed to unmarshal transaction item to models.Transaction for alert %s: %w", alert.ID, err)
					}
					allRelatedTxIDs = append(allRelatedTxIDs, tx.TransactionID)
				}
			} else {
				// Fallback for single transaction if not a list
				txBytes, err := json.Marshal(rawMatchingTxs)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal single transaction from rule_details for alert %s: %w", alert.ID, err)
				}
				var tx models.Transaction
				err = json.Unmarshal(txBytes, &tx)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal single transaction to models.Transaction for alert %s: %w", alert.ID, err)
				}
				allRelatedTxIDs = append(allRelatedTxIDs, tx.TransactionID)
			}
		} else {
			// If no matching_transactions in rule_details, fall back to alert's TransactionID
			allRelatedTxIDs = append(allRelatedTxIDs, alert.TransactionID)
		}

		alertIDToDetails[alert.ID] = alert
		alerts = append(alerts, alert)
	}

	if err = alertRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert rows: %w", err)
	}

	if len(allRelatedTxIDs) == 0 {
		return nil, nil // No related transactions found
	}

	// 2. Fetch all related transactions
	uniqueRelatedTxIDs := make(map[string]bool)
	for _, id := range allRelatedTxIDs {
		uniqueRelatedTxIDs[id] = true
	}
	var finalTxIDs []string
	for id := range uniqueRelatedTxIDs {
		finalTxIDs = append(finalTxIDs, id)
	}

	txQueryArgs := make([]interface{}, len(finalTxIDs))
	for i, id := range finalTxIDs {
		txQueryArgs[i] = id
	}

	txQuery := `
		SELECT transaction_id, account_id, amount, currency, timestamp,
			source_country, destination_country, transaction_type, status
		FROM transactions
		WHERE transaction_id IN (?` + strings.Repeat(",?", len(finalTxIDs)-1) + `)
		ORDER BY timestamp ASC
	`
	txRows, err := db.Query(txQuery, txQueryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer txRows.Close()

	transactionsMap := make(map[string]models.Transaction)
	var accountIDs []string // To collect account IDs for fetching account details
	uniqueAccountIDs := make(map[string]bool)

	for txRows.Next() {
		var tx models.Transaction
		err := txRows.Scan(
			&tx.TransactionID, &tx.AccountID, &tx.Amount, &tx.Currency, &tx.Timestamp,
			&tx.SourceCountry, &tx.DestinationCountry, &tx.TransactionType, &tx.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		transactionsMap[tx.TransactionID] = tx
		if _, exists := uniqueAccountIDs[tx.AccountID]; !exists {
			uniqueAccountIDs[tx.AccountID] = true
			accountIDs = append(accountIDs, tx.AccountID)
		}
	}

	if err = txRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	if len(accountIDs) == 0 {
		return nil, fmt.Errorf("no account IDs found for related transactions")
	}

	// 3. Fetch account details
	accountQueryArgs := make([]interface{}, len(accountIDs))
	for i, id := range accountIDs {
		accountQueryArgs[i] = id
	}

	accountQuery := `
		SELECT account_id, holder_name, address, date_of_birth
		FROM accounts
		WHERE account_id IN (?` + strings.Repeat(",?", len(accountIDs)-1) + `)
	`
	accRows, err := db.Query(accountQuery, accountQueryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer accRows.Close()

	accountsMap := make(map[string]models.Account)
	for accRows.Next() {
		var account models.Account
		err := accRows.Scan(&account.AccountID, &account.HolderName, &account.Address, &account.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account row: %w", err)
		}
		accountsMap[account.AccountID] = account
	}

	if err = accRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating account rows: %w", err)
	}

	// 4. Aggregate data into SARReport
	report := &models.SARReport{
		Patterns: make(map[string]models.SuspiciousActivityPattern),
	}
	isFirstRow := true // To set subject details and initial start date

	var minStartDate time.Time = time.Now().Add(100 * 365 * 24 * time.Hour) // Far future
	var maxEndDate time.Time

	for _, alert := range alerts {
		// Get account info for the primary account associated with the alert (or first transaction)
		var associatedAccount models.Account
		if len(allRelatedTxIDs) > 0 {
			if tx, ok := transactionsMap[alert.TransactionID]; ok { // Use alert's primary transaction ID
				associatedAccount = accountsMap[tx.AccountID]
			} else { // Fallback to any related transaction's account if alert.TransactionID isn't a direct match in map
				for _, txID := range allRelatedTxIDs {
					if tx, ok := transactionsMap[txID]; ok {
						associatedAccount = accountsMap[tx.AccountID]
						break
					}
				}
			}
		}

		if isFirstRow && associatedAccount.AccountID != "" {
			report.SubjectName = associatedAccount.HolderName
			report.SubjectAddress = associatedAccount.Address
			report.SubjectDateOfBirth = associatedAccount.DateOfBirth
			isFirstRow = false
		}

		pattern := report.Patterns[alert.AlertType]
		pattern.PatternDescription = alert.AlertType

		// Extract transactions for this specific alert from RuleDetails
		var alertRuleDetails map[string]interface{}
		if len(alert.RuleDetails) > 0 { // Check if RuleDetails has content
			// Since RuleDetails is now JSONMap, we need to unmarshal it from []byte if it was stored as text
			// For in-memory GORM setup, it might be directly available as map[string]interface{}
			// Assuming alert.RuleDetails is already unmarshalled to map[string]interface{} by models.JSONMap Scan method
			alertRuleDetails = alert.RuleDetails
		}

		var currentAlertMatchingTxs []models.Transaction
		if rawMatchingTxs, ok := alertRuleDetails["matching_transactions"]; ok {
			if list, isList := rawMatchingTxs.([]interface{}); isList {
				for _, item := range list {
					txBytes, err := json.Marshal(item)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal transaction item from alert rule_details for report: %w", err)
					}
					var tx models.Transaction
					err = json.Unmarshal(txBytes, &tx)
					if err != nil {
						return nil, fmt.Errorf("failed to unmarshal transaction item to models.Transaction from alert rule_details for report: %w", err)
					}
					currentAlertMatchingTxs = append(currentAlertMatchingTxs, tx)
				}
			} else {
				// Fallback for single transaction if not a list
				txBytes, err := json.Marshal(rawMatchingTxs)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal single transaction from alert rule_details for report: %w", err)
				}
				var tx models.Transaction
				err = json.Unmarshal(txBytes, &tx)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal single transaction to models.Transaction from alert rule_details for report: %w", err)
				}
				currentAlertMatchingTxs = append(currentAlertMatchingTxs, tx)
			}
		} else {
			// Fallback: If no 'matching_transactions' in RuleDetails, just use the alert's primary transaction
			if tx, ok := transactionsMap[alert.TransactionID]; ok {
				currentAlertMatchingTxs = append(currentAlertMatchingTxs, tx)
			}
		}

		for _, tx := range currentAlertMatchingTxs {
			pattern.Transactions = append(pattern.Transactions, tx)
			pattern.TotalAmount += tx.Amount
			pattern.TransactionCount++

			if tx.Timestamp.Before(minStartDate) {
				minStartDate = tx.Timestamp
			}
			if tx.Timestamp.After(maxEndDate) {
				maxEndDate = tx.Timestamp
			}
			report.TotalSuspiciousAmount += tx.Amount
			report.TotalTransactionCount++
		}
		report.Patterns[alert.AlertType] = pattern
	}

	if len(alerts) > 0 {
		report.StartDate = minStartDate
		report.EndDate = maxEndDate
	}

	return report, nil
}
