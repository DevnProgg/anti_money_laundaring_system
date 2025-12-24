package services_test

import (
	"AML/internal/config"
	"AML/internal/models"
	"AML/internal/services"
	"database/sql"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Helper function to setup an in-memory SQLite database
func setupTestDB(t *testing.T) (*gorm.DB, *sql.DB, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Transaction{},
		&models.Alert{},
		&models.Account{},
		&models.SARReport{},
	)
	require.NoError(t, err)

	// Insert a dummy account for SAR generation
	dummyAccount := models.Account{
		AccountID:   "ACC123",
		HolderName:  "John Doe",
		Address:     "123 Main St",
		DateOfBirth: "1980-01-01",
	}
	err = db.Create(&dummyAccount).Error
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	return db, sqlDB, func() {
		sqlDB.Close()
	}
}

func TestAMLEndToEndFlow(t *testing.T) {
	// 1. Setup: Create in-memory database and clean up afterwards
	gormDB, sqlDB, teardown := setupTestDB(t)
	defer teardown()

	// 2. Load rules from rules.json
	// Create a temporary rules.json file for the test
	rulesContent := `
	[
		{
			"rule_id": "structuring_pattern_detection",
			"name": "Structuring Pattern Detection (e.g., Smurfing)",
			"threshold_value": 27000.0,
			"time_window": "12h",
			"enabled": true,
			"min_count": 3
		}
	]`
	tmpfile, err := ioutil.TempFile("", "rules-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up

	_, err = tmpfile.Write([]byte(rulesContent))
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	rules, err := config.LoadRules(tmpfile.Name())
	require.NoError(t, err)
	require.NotEmpty(t, rules, "rules should not be empty")

	var structuringRule config.Rule
	foundStructuringRule := false
	for _, r := range rules {
		if r.RuleID == "structuring_pattern_detection" {
			structuringRule = r
			foundStructuringRule = true
			break
		}
	}
	require.True(t, foundStructuringRule, "Structuring pattern rule not found")

	structuringTimeWindow, err := structuringRule.GetTimeWindow()
	require.NoError(t, err)

	type ruleConfig struct {
		RuleID         string  `json:"rule_id"`
		Name           string  `json:"name"`
		ThresholdValue float64 `json:"threshold_value"`
		TimeWindow     string  `json:"time_window"`
		Enabled        bool    `json:"enabled"`
		MinCount       int     `json:"min_count"`
	}
	var loadedRuleConfigs []ruleConfig
	err = json.Unmarshal([]byte(rulesContent), &loadedRuleConfigs)
	require.NoError(t, err)

	var minCount int
	for _, rc := range loadedRuleConfigs {
		if rc.RuleID == "structuring_pattern_detection" {
			minCount = rc.MinCount
			break
		}
	}
	require.NotZero(t, minCount, "min_count not found or invalid for structuring_pattern_detection rule")

	// 3. Ingest 10 transactions for account "ACC123"
	// 3 transactions trigger structuring pattern ($9000, $9500, $9800 within 12 hours)
	now := time.Now()
	transactions := []models.Transaction{
		// Suspicious transactions for ACC123
		{TransactionID: uuid.New().String(), AccountID: "ACC123", Amount: 9000.00, Currency: "USD", Timestamp: now.Add(-1 * time.Hour), SourceCountry: "USA", DestinationCountry: "USA", TransactionType: "deposit", Status: "completed"},
		{TransactionID: uuid.New().String(), AccountID: "ACC123", Amount: 9500.00, Currency: "USD", Timestamp: now.Add(-2 * time.Hour), SourceCountry: "USA", DestinationCountry: "USA", TransactionType: "deposit", Status: "completed"},
		{TransactionID: uuid.New().String(), AccountID: "ACC123", Amount: 9800.00, Currency: "USD", Timestamp: now.Add(-3 * time.Hour), SourceCountry: "USA", DestinationCountry: "USA", TransactionType: "deposit", Status: "completed"},
		// Benign transactions
		{TransactionID: uuid.New().String(), AccountID: "ACC124", Amount: 15000.00, Currency: "USD", Timestamp: now.Add(-1 * time.Hour), SourceCountry: "CAN", DestinationCountry: "USA", TransactionType: "withdrawal", Status: "completed"},
		{TransactionID: uuid.New().String(), AccountID: "ACC125", Amount: 500.00, Currency: "USD", Timestamp: now.Add(-4 * time.Hour), SourceCountry: "MEX", DestinationCountry: "USA", TransactionType: "payment", Status: "completed"},
		{TransactionID: uuid.New().String(), AccountID: "ACC123", Amount: 100.00, Currency: "USD", Timestamp: now.Add(-13 * time.Hour), SourceCountry: "USA", DestinationCountry: "USA", TransactionType: "deposit", Status: "completed"}, // Outside window
		{TransactionID: uuid.New().String(), AccountID: "ACC126", Amount: 2000.00, Currency: "EUR", Timestamp: now.Add(-5 * time.Hour), SourceCountry: "DEU", DestinationCountry: "FRA", TransactionType: "transfer", Status: "completed"},
		{TransactionID: uuid.New().String(), AccountID: "ACC127", Amount: 750.00, Currency: "GBP", Timestamp: now.Add(-6 * time.Hour), SourceCountry: "GBR", DestinationCountry: "USA", TransactionType: "purchase", Status: "completed"},
		{TransactionID: uuid.New().String(), AccountID: "ACC128", Amount: 12000.00, Currency: "USD", Timestamp: now.Add(-7 * time.Hour), SourceCountry: "USA", DestinationCountry: "USA", TransactionType: "deposit", Status: "completed"}, // Single large, but not structuring
		{TransactionID: uuid.New().String(), AccountID: "ACC129", Amount: 300.00, Currency: "JPY", Timestamp: now.Add(-8 * time.Hour), SourceCountry: "JPN", DestinationCountry: "USA", TransactionType: "withdrawal", Status: "completed"},
	}

	for _, tx := range transactions {
		err := gormDB.Create(&tx).Error
		require.NoError(t, err, fmt.Sprintf("Failed to create transaction: %v", tx))
	}

	// 4. Run detection engine
	// Get all transactions for ACC123 to pass to the detector
	var acc123Transactions []models.Transaction
	err = gormDB.Where("account_id = ?", "ACC123").Find(&acc123Transactions).Error
	require.NoError(t, err)

	detected, matchingTxs := services.DetectStructuring(
		"ACC123",
		acc123Transactions,
		structuringTimeWindow,
		1.0, // Assuming a low threshold for individual transactions for structuring detection
		structuringRule.ThresholdValue, // Using the overall threshold value as the upper bound for individual txns to be part of the pattern
		minCount,
	)

	require.True(t, detected, "Structuring pattern should have been detected")
	require.Len(t, matchingTxs, minCount, "Incorrect number of matching transactions")

	// Generate Alert for the detected pattern
	ruleDetails := map[string]interface{}{
		"matching_transactions": matchingTxs,
		"total_amount":          9000.00 + 9500.00 + 9800.00,
		"time_window":           structuringRule.TimeWindow,
		"threshold_value":       structuringRule.ThresholdValue,
	}

	// For alert generation, we typically associate with a single representative transaction
	// or create a conceptual alert not tied to a single transaction ID.
	// For this test, let's tie it to the *first* suspicious transaction for simplicity,
	// though in a real system, you might generate an alert that aggregates the pattern.
	representativeTx := matchingTxs[0]
	alert, err := services.GenerateAlert(representativeTx, services.AlertTypeStructuringPattern, ruleDetails)
	require.NoError(t, err)
	require.NotNil(t, alert, "Alert should not be nil")

	err = gormDB.Create(&alert).Error
	require.NoError(t, err, "Failed to create alert in DB")

	// 5. Verify alert generated with correct priority
	var fetchedAlert models.Alert
	err = gormDB.First(&fetchedAlert, "id = ?", alert.ID).Error
	require.NoError(t, err, "Failed to fetch alert from DB")

	assert.Equal(t, services.AlertTypeStructuringPattern, fetchedAlert.AlertType, "Alert type should be STRUCTURING_PATTERN")
	assert.Equal(t, models.Critical, fetchedAlert.Priority, "Alert priority should be Critical")
	assert.Equal(t, models.StatusOpen, fetchedAlert.Status, "Alert status should be OPEN")
	assert.Equal(t, "ACC123", representativeTx.AccountID, "Alert should be for ACC123")

	// 6. Transition alert to INVESTIGATING
	fetchedAlert.Status = models.StatusInvestigating
	fetchedAlert.TransitionAt = time.Now()
	err = gormDB.Save(&fetchedAlert).Error
	require.NoError(t, err, "Failed to update alert status")

	var updatedAlert models.Alert
	err = gormDB.First(&updatedAlert, "id = ?", alert.ID).Error
	require.NoError(t, err, "Failed to fetch updated alert from DB")
	assert.Equal(t, models.StatusInvestigating, updatedAlert.Status, "Alert status should be INVESTIGATING")

	// 7. Generate SAR report
	// GenerateSARData expects alertIDs, even if it's just one for the pattern
	sarReport, err := services.GenerateSARData([]string{updatedAlert.ID}, sqlDB) // Use sqlDB here
	require.NoError(t, err)
	require.NotNil(t, sarReport, "SAR report should not be nil")

	// Convert matchingTxs to JSON for comparison, assuming SARReport stores them as JSON
	// The SAR report's Patterns map stores []models.Transaction directly
	expectedTotalAmount := 9000.00 + 9500.00 + 9800.00

	// 8. Verify SAR report
	assert.Equal(t, "John Doe", sarReport.SubjectName, "SAR SubjectName mismatch")
	assert.Equal(t, "123 Main St", sarReport.SubjectAddress, "SAR SubjectAddress mismatch")
	assert.Equal(t, "1980-01-01", sarReport.SubjectDateOfBirth, "SAR SubjectDateOfBirth mismatch")
	assert.InDelta(t, expectedTotalAmount, sarReport.TotalSuspiciousAmount, 0.001, "SAR TotalSuspiciousAmount mismatch")
	assert.Equal(t, minCount, sarReport.TotalTransactionCount, "SAR TotalTransactionCount mismatch")

	// Verify the structuring pattern details within the SAR
	pattern, ok := sarReport.Patterns[services.AlertTypeStructuringPattern]
	require.True(t, ok, "SAR report should contain structuring pattern")
	assert.Equal(t, services.AlertTypeStructuringPattern, pattern.PatternDescription, "Pattern description mismatch")
	assert.Len(t, pattern.Transactions, minCount, "SAR pattern should contain %d transactions", minCount)
	assert.InDelta(t, expectedTotalAmount, pattern.TotalAmount, 0.001, "SAR pattern total amount mismatch")
	assert.Equal(t, minCount, pattern.TransactionCount, "SAR pattern transaction count mismatch")

	// Check if the specific transactions are included in the SAR pattern
	sarTxIDs := make(map[string]bool)
	for _, tx := range pattern.Transactions {
		sarTxIDs[tx.TransactionID] = true
	}
	for _, expectedTx := range matchingTxs {
		assert.Contains(t, sarTxIDs, expectedTx.TransactionID, "Suspicious transaction %s not found in SAR", expectedTx.TransactionID)
	}
}
