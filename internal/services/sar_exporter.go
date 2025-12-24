package services

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"AML/internal/models"
)

// MaskAccountNumber masks the account number, showing only the last 4 digits.
func MaskAccountNumber(accountNum string) string {
	if len(accountNum) <= 4 {
		return accountNum
	}
	return "XXXX-XXXX-XXXX-" + accountNum[len(accountNum)-4:]
}

// ExportSARToJSON exports a SARReport to a JSON file with specified formatting and masking.
func ExportSARToJSON(report *models.SARReport, filepath string) error {
	// Custom type to handle JSON marshaling for SARReport, applying masking and ISO 8601

	type TransactionJSON struct {
		TransactionID      string  `json:"transaction_id"`
		AccountID          string  `json:"account_id"` // Masked
		Amount             float64 `json:"amount"`
		Currency           string  `json:"currency"`
		Timestamp          string  `json:"timestamp"` // Will be ISO 8601
		SourceCountry      string  `json:"source_country"`
		DestinationCountry string  `json:"destination_country"`
		TransactionType    string  `json:"transaction_type"`
		Status             string  `json:"status"`
	}

	type PatternJSON struct {
		PatternDescription string            `json:"pattern_description"`
		Transactions       []TransactionJSON `json:"transactions"`
		TotalAmount        float64           `json:"total_amount"`
		TransactionCount   int               `json:"transaction_count"`
	}

	type SARReportJSON struct {
		SubjectName           string                 `json:"subject_name"`
		SubjectAddress        string                 `json:"subject_address"`
		SubjectDateOfBirth    string                 `json:"subject_date_of_birth"`
		StartDate             string                 `json:"start_date"` // Will be ISO 8601
		EndDate               string                 `json:"end_date"`   // Will be ISO 8601
		TotalSuspiciousAmount float64                `json:"total_suspicious_amount"`
		TotalTransactionCount int                    `json:"total_transaction_count"`
		Patterns              map[string]PatternJSON `json:"patterns"`
	}

	sarReportJSON := SARReportJSON{
		SubjectName:           report.SubjectName,
		SubjectAddress:        report.SubjectAddress,
		SubjectDateOfBirth:    report.SubjectDateOfBirth,
		StartDate:             report.StartDate.Format(time.RFC3339),
		EndDate:               report.EndDate.Format(time.RFC3339),
		TotalSuspiciousAmount: report.TotalSuspiciousAmount,
		TotalTransactionCount: report.TotalTransactionCount,
		Patterns:              make(map[string]PatternJSON),
	}

	for key, pattern := range report.Patterns {
		patternJSON := PatternJSON{
			PatternDescription: pattern.PatternDescription,
			TotalAmount:        pattern.TotalAmount,
			TransactionCount:   pattern.TransactionCount,
			Transactions:       make([]TransactionJSON, len(pattern.Transactions)),
		}
		for i, transaction := range pattern.Transactions {
			patternJSON.Transactions[i] = TransactionJSON{
				TransactionID:      transaction.TransactionID,
				AccountID:          MaskAccountNumber(transaction.AccountID),
				Amount:             transaction.Amount,
				Currency:           transaction.Currency,
				Timestamp:          transaction.Timestamp.Format(time.RFC3339),
				SourceCountry:      transaction.SourceCountry,
				DestinationCountry: transaction.DestinationCountry,
				TransactionType:    transaction.TransactionType,
				Status:             transaction.Status,
			}
		}
		sarReportJSON.Patterns[key] = patternJSON
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(sarReportJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SAR report to JSON: %w", err)
	}

	// Write to file with 0600 permissions
	err = os.WriteFile(filepath, jsonData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write SAR report to file: %w", err)
	}

	return nil
}
