package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type SARPatterns map[string]SuspiciousActivityPattern

// Value implements the driver.Valuer interface.
func (sp SARPatterns) Value() (driver.Value, error) {
	if sp == nil {
		return nil, nil
	}
	return json.Marshal(sp)
}

// Scan implements the sql.Scanner interface.
func (sp *SARPatterns) Scan(value interface{}) error {
	if value == nil {
		*sp = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal SARPatterns value: %v", value)
	}
	// Ensure the map is initialized before unmarshalling
	if *sp == nil {
		*sp = make(SARPatterns)
	}
	return json.Unmarshal(bytes, sp)
}

// SuspiciousActivityPattern groups transactions by a specific pattern of activity.
type SuspiciousActivityPattern struct {
	PatternDescription string        `json:"pattern_description"`
	Transactions       []Transaction `json:"transactions"`
	TotalAmount        float64       `json:"total_amount"`
	TransactionCount   int           `json:"transaction_count"`
}

// SARReport represents a complete Suspicious Activity Report, aggregating multiple alerts.
type SARReport struct {
	SubjectName           string                               `json:"subject_name"`
	SubjectAddress        string                               `json:"subject_address"`
	SubjectDateOfBirth    string                               `json:"subject_date_of_birth"`
	StartDate             time.Time                            `json:"start_date"`
	EndDate               time.Time                            `json:"end_date"`
	TotalSuspiciousAmount float64                              `json:"total_suspicious_amount"`
	TotalTransactionCount int                                  `json:"total_transaction_count"`
	Patterns              SARPatterns `json:"patterns" gorm:"type:text"`
}
