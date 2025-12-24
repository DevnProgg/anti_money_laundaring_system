package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"AML/internal/models"
)

// TransactionHandler handles the creation of new transactions.
func TransactionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var t models.Transaction
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := validateTransaction(&t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if t.TransactionID == "" {
			t.TransactionID = uuid.New().String()
		}

		if err := insertTransaction(db, &t); err != nil {
			http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"transaction_id": t.TransactionID})
	}
}

// validateTransaction validates the transaction data.
func validateTransaction(t *models.Transaction) error {
	if t.AccountID == "" || t.Currency == "" || t.SourceCountry == "" || t.DestinationCountry == "" || t.TransactionType == "" || t.Status == "" {
		return fmt.Errorf("all required fields must be present")
	}

	if t.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if !isValidCurrency(t.Currency) {
		return fmt.Errorf("currency must be a 3-letter ISO code")
	}

	return nil
}

// isValidCurrency checks if a currency is a 3-letter ISO code.
func isValidCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	return regexp.MustCompile(`^[A-Z]{3}$`).MatchString(strings.ToUpper(currency))
}

// insertTransaction inserts a new transaction into the database.
func insertTransaction(db *sql.DB, t *models.Transaction) error {
	query := `
		INSERT INTO transactions (transaction_id, account_id, amount, currency, "timestamp", source_country, destination_country, transaction_type, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.TransactionID, t.AccountID, t.Amount, t.Currency, time.Now(), t.SourceCountry, t.DestinationCountry, t.TransactionType, t.Status)
	return err
}
