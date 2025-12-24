package services

import (
	"fmt"
	"time"

	"AML/internal/models"
)

var (
	// ErrInvalidTransition is returned for invalid status transitions.
	ErrInvalidTransition = fmt.Errorf("invalid status transition")
)

var allowedTransitions = map[string][]string{
	models.StatusOpen:          {models.StatusInvestigating},
	models.StatusInvestigating: {models.StatusEscalated, models.StatusFalsePositive},
	models.StatusEscalated:     {models.StatusClosed},
	models.StatusClosed:        {},
	models.StatusFalsePositive: {},
}

// TransitionAlertStatus updates the status of an alert after validation.
func TransitionAlertStatus(alert *models.Alert, newStatus string, investigatorID string) error {
	currentStatus := alert.Status
	validTransitions, ok := allowedTransitions[currentStatus]
	if !ok {
		return fmt.Errorf("unknown status: %s", currentStatus)
	}

	isValid := false
	for _, status := range validTransitions {
		if status == newStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, currentStatus, newStatus)
	}

	alert.Status = newStatus
	alert.TransitionAt = time.Now()
	if investigatorID != "" {
		alert.AssignedTo = investigatorID
	}

	return nil
}
