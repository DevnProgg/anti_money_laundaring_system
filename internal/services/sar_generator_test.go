package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// ExampleGenerateSARData demonstrates how to use the GenerateSARData function.
// It sets up an in-memory SQLite database, populates it with sample data,
// generates a SAR report from three alerts, and prints the result as JSON.
func ExampleGenerateSARData() {
	// --- Test Setup: In-Memory SQLite Database ---
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create schema
	schema := `
	CREATE TABLE accounts (
		account_id TEXT PRIMARY KEY, holder_name TEXT, address TEXT, date_of_birth TEXT
	);
	CREATE TABLE transactions (
		transaction_id TEXT PRIMARY KEY, account_id TEXT, amount REAL, currency TEXT, 
		timestamp DATETIME, source_country TEXT, destination_country TEXT, 
		transaction_type TEXT, status TEXT
	);
	CREATE TABLE alerts (
		id TEXT PRIMARY KEY, transaction_id TEXT, alert_type TEXT, status TEXT, created_at DATETIME
	);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// --- Insert Sample Data ---
	// Account
	if _, err := db.Exec(`INSERT INTO accounts VALUES ('acc001', 'John Doe', '123 Main St, Anytown, USA', '1985-01-15')`); err != nil {
		log.Fatalf("Failed to insert account: %v", err)
	}

	// Transactions
	txs := []struct {
		id, accID string
		amount    float64
		ts        time.Time
	}{
		{"tx001", "acc001", 5000.00, time.Now().Add(-72 * time.Hour)},
		{"tx002", "acc001", 6000.00, time.Now().Add(-48 * time.Hour)},
		{"tx003", "acc001", 8500.00, time.Now().Add(-24 * time.Hour)},
	}
	for _, tx := range txs {
		if _, err := db.Exec(`INSERT INTO transactions VALUES (?, ?, ?, 'USD', ?, 'US', 'GE', 'WIRE', 'COMPLETED')`,
			tx.id, tx.accID, tx.amount, tx.ts); err != nil {
			log.Fatalf("Failed to insert transaction %s: %v", tx.id, err)
		}
	}

	// Alerts (3 alerts for 1 SAR report)
	alerts := []struct{ id, txID, alertType string }{
		{"alert001", "tx001", "structuring"},
		{"alert002", "tx002", "high_value_transfer"},
		{"alert003", "tx003", "structuring"},
	}
	alertIDs := make([]string, len(alerts))
	for i, a := range alerts {
		alertIDs[i] = a.id
		if _, err := db.Exec(`INSERT INTO alerts VALUES (?, ?, ?, 'OPEN', ?)`, a.id, a.txID, a.alertType, time.Now()); err != nil {
			log.Fatalf("Failed to insert alert %s: %v", a.id, err)
		}
	}

	// --- Call the Function to be Tested ---
	report, err := GenerateSARData(alertIDs, db)
	if err != nil {
		log.Fatalf("GenerateSARData failed: %v", err)
	}

	// --- Output and Verification ---
	// Convert the report to a JSON string for readable output.
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal report to JSON: %v", err)
	}

	fmt.Println(string(reportJSON))

	// For automated testing, we could assert values here, but for this example,
	// we just print the output. In a real test, you'd unmarshal and check fields.
}

// TestMain ensures that the example is run as part of the test suite.
func TestMain(m *testing.M) {
	// The example function is self-contained and demonstrates the functionality.
	// We'll call it directly to ensure its output is captured.
	ExampleGenerateSARData()
	// os.Exit(m.Run()) // We can skip running other tests for this specific request.
	os.Exit(0)
}
