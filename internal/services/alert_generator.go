package services

import (
	"fmt"
	"math"
	"time"

	"AML/internal/models"
	"github.com/google/uuid"
)

const (
	// AlertTypeThresholdViolation is for threshold violation alerts.
	AlertTypeThresholdViolation = "THRESHOLD_VIOLATION"
	// AlertTypeAnomalyDetected is for anomaly detection alerts.
	AlertTypeAnomalyDetected = "ANOMALY_DETECTED"
	// AlertTypeStructuringPattern is for structuring pattern alerts.
	AlertTypeStructuringPattern = "STRUCTURING_PATTERN"
	// AlertTypeGeographicRisk is for geographic risk alerts.
	AlertTypeGeographicRisk = "GEOGRAPHIC_RISK"
)

// GenerateAlert creates a new alert for a suspicious transaction.
func GenerateAlert(tx models.Transaction, alertType string, ruleDetails map[string]interface{}) (*models.Alert, error) {
	priority, err := getPriorityForAlertType(alertType)
	if err != nil {
		return nil, err
	}

	score, err := CalculateRiskScore(tx.Amount, priority)
	if err != nil {
		return nil, err
	}

	alert := &models.Alert{
		ID:            uuid.New().String(),
		TransactionID: tx.TransactionID,
		AlertType:     alertType,
		Priority:      priority,
		Score:         score,
		CreatedAt:     time.Now(),
		Status:        "OPEN",
		AssignedTo:    "", // Initially unassigned
		RuleDetails:   ruleDetails,
	}

	return alert, nil
}

// CalculateRiskScore calculates the risk score for an alert.
func CalculateRiskScore(amount float64, priority models.PriorityLevel) (float64, error) {
	var typeWeight float64
	switch priority {
	case models.Medium:
		typeWeight = 50
	case models.High:
		typeWeight = 75
	case models.Critical:
		typeWeight = 100
	default:
		return 0, fmt.Errorf("invalid priority level")
	}

	// Normalize amount to a 0-20 range, capped at $1M
	amountScore := (math.Min(amount, 1000000) / 1000000) * 20

	score := typeWeight*0.8 + amountScore*0.2
	return math.Min(score, 100), nil // Cap score at 100
}

func getPriorityForAlertType(alertType string) (models.PriorityLevel, error) {
	switch alertType {
	case AlertTypeThresholdViolation, AlertTypeGeographicRisk:
		return models.Medium, nil
	case AlertTypeAnomalyDetected:
		return models.High, nil
	case AlertTypeStructuringPattern:
		return models.Critical, nil
	default:
		return 0, fmt.Errorf("invalid alert type: %s", alertType)
	}
}
