package models

import (
	"time"
)

// DELIVERABLE 3: Go struct for transactions
type Transaction struct {
	TransactionID      string    `db:"transaction_id" json:"transaction_id"`
	AccountID          string    `db:"account_id" json:"account_id"`
	Amount             float64   `db:"amount" json:"amount"`
	Currency           string    `db:"currency" json:"currency"`
	Timestamp          time.Time `db:"timestamp" json:"timestamp"`
	SourceCountry      string    `db:"source_country" json:"source_country"`
	DestinationCountry string    `db:"destination_country" json:"destination_country"`
	TransactionType    string    `db:"transaction_type" json:"transaction_type"`
	Status             string    `db:"status" json:"status"`
}
