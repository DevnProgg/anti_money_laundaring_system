package services

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"AML/internal/models"
)

func TestMaskAccountNumber(t *testing.T) {
	tests := []struct {
		name       string
		accountNum string
		expected   string
	}{
		{"Short account number", "123", "123"},
		{"4-digit account number", "1234", "1234"},
		{"Long account number", "1234567890123456", "XXXX-XXXX-XXXX-3456"},
		{"Empty account number", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskAccountNumber(tt.accountNum)
			if got != tt.expected {
				t.Errorf("MaskAccountNumber(%q) = %q, want %q", tt.accountNum, got, tt.expected)
			}
		})
	}
}

func TestExportSARToJSON(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := ioutil.TempDir("", "sar_export_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up the temporary directory

	testFilePath := filepath.Join(tmpDir, "test_sar_report.json")

	// Create a mock SARReport
	mockTransaction := models.Transaction{
		TransactionID:      "TXN123",
		AccountID:          "ACCT9876543210",
		Amount:             1500.75,
		Currency:           "USD",
		Timestamp:          time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
		SourceCountry:      "USA",
		DestinationCountry: "CAN",
		TransactionType:    "Wire Transfer",
		Status:             "Completed",
	}

	mockPattern := models.SuspiciousActivityPattern{
		PatternDescription: "Large single transaction",
		Transactions:       []models.Transaction{mockTransaction},
		TotalAmount:        1500.75,
		TransactionCount:   1,
	}

	mockSARReport := &models.SARReport{
		SubjectName:           "John Doe",
		SubjectAddress:        "123 Main St",
		SubjectDateOfBirth:    "1980-05-20",
		StartDate:             time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:               time.Date(2023, 1, 31, 23, 59, 59, 999999999, time.UTC),
		TotalSuspiciousAmount: 1500.75,
		TotalTransactionCount: 1,
		Patterns: map[string]models.SuspiciousActivityPattern{
			"LargeTransaction": mockPattern,
		},
	}

	// 1. Test successful export
	err = ExportSARToJSON(mockSARReport, testFilePath)
	if err != nil {
		t.Fatalf("ExportSARToJSON failed: %v", err)
	}

	// 2. Verify file permissions (0600)
	fileInfo, err := os.Stat(testFilePath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if fileInfo.Mode().Perm() != 0600 {
		t.Errorf("file permissions are %o, want 0600", fileInfo.Mode().Perm())
	}

	// 3. Read and parse the JSON file
	jsonData, err := ioutil.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("failed to read exported JSON file: %v", err)
	}

	var exportedData map[string]interface{}
	err = json.Unmarshal(jsonData, &exportedData)
	if err != nil {
		t.Fatalf("failed to unmarshal exported JSON: %v", err)
	}

	// 4. Verify JSON content (indentation, masked account, ISO 8601)
	// Check for indentation (indirectly by ensuring it's not a single line and has specific formatting)
	// This is a basic check; a more robust check might involve comparing against a pre-formatted string.
	jsonString := string(jsonData)
	if !containsIndentation(jsonString) {
		t.Errorf("exported JSON does not appear to be indented")
	}

	// Verify masked account number
	patterns := exportedData["patterns"].(map[string]interface{})
	largeTransactionPattern := patterns["LargeTransaction"].(map[string]interface{})
	transactions := largeTransactionPattern["transactions"].([]interface{})
	firstTransaction := transactions[0].(map[string]interface{})
	maskedAccountID := firstTransaction["account_id"].(string)
	if maskedAccountID != "XXXX-XXXX-XXXX-3210" {
		t.Errorf("expected masked account ID 'XXXX-XXXX-XXXX-3210', got %q", maskedAccountID)
	}

	// Verify ISO 8601 timestamps
	startDate := exportedData["start_date"].(string)
	endDate := exportedData["end_date"].(string)
	transactionTimestamp := firstTransaction["timestamp"].(string)

	expectedStartDate := mockSARReport.StartDate.Format(time.RFC3339)
	expectedEndDate := mockSARReport.EndDate.Format(time.RFC3339)
	expectedTransactionTimestamp := mockTransaction.Timestamp.Format(time.RFC3339)

	if startDate != expectedStartDate {
		t.Errorf("expected start_date %q, got %q", expectedStartDate, startDate)
	}
	if endDate != expectedEndDate {
		t.Errorf("expected end_date %q, got %q", expectedEndDate, endDate)
	}
	if transactionTimestamp != expectedTransactionTimestamp {
		t.Errorf("expected transaction timestamp %q, got %q", expectedTransactionTimestamp, transactionTimestamp)
	}

	// 5. Test error handling (e.g., invalid directory)
	invalidFilePath := filepath.Join("/nonexistent_dir", "test_sar_report.json")
	err = ExportSARToJSON(mockSARReport, invalidFilePath)
	if err == nil {
		t.Errorf("ExportSARToJSON did not return an error for invalid file path")
	}
}

// Helper to check for indentation
func containsIndentation(s string) bool {
	return len(s) > 0 && s[0] == '{' && (json.Valid([]byte(s)) && len(s) > 100) // Simple heuristic, not perfect
}
