package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface.
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface.
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONMap value: %v", value)
	}
	// Ensure the map is initialized before unmarshalling
	if *j == nil {
		*j = make(JSONMap)
	}
	return json.Unmarshal(bytes, j)
}

// PriorityLevel defines the priority of an alert.
type PriorityLevel int

const (
	// Medium is a medium priority alert.
	Medium PriorityLevel = iota
	// High is a high priority alert.
	High
	// Critical is a critical priority alert.
	Critical
)

// String returns the string representation of a PriorityLevel.
func (p PriorityLevel) String() string {
	return [...]string{"MEDIUM", "HIGH", "CRITICAL"}[p]
}

const (
	// StatusOpen is for newly created alerts.
	StatusOpen = "OPEN"
	// StatusInvestigating is for alerts under investigation.
	StatusInvestigating = "INVESTIGATING"
	// StatusEscalated is for alerts that have been escalated.
	StatusEscalated = "ESCALATED"
	// StatusClosed is for alerts that have been closed.
	StatusClosed = "CLOSED"
	// StatusFalsePositive is for alerts that have been marked as false positives.
	StatusFalsePositive = "FALSE_POSITIVE"
)

// Alert represents a generated alert for a suspicious transaction.
type Alert struct {
	ID            string                 `json:"id"`
	TransactionID string                 `json:"transaction_id"`
	AlertType     string                 `json:"alert_type"`
	Priority      PriorityLevel          `json:"priority"`
	Score         float64                `json:"score"`
	CreatedAt     time.Time              `json:"created_at"`
	Status        string                 `json:"status"`
	AssignedTo    string                 `json:"assigned_to"`
	RuleDetails   JSONMap                `json:"rule_details" gorm:"type:text"`
	TransitionAt  time.Time              `json:"transition_at"`
}
