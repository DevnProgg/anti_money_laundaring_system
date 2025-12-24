package services

import (
	"AML/internal/models"
	"testing"
	"time"
)

var _ = time.Now

func TestTransitionAlertStatus(t *testing.T) {
	// Test Case 1: Valid: OPEN → INVESTIGATING
	t.Run("valid_open_to_investigating", func(t *testing.T) {
		alert := &models.Alert{Status: models.StatusOpen}
		err := TransitionAlertStatus(alert, models.StatusInvestigating, "investigator-1")
		if err != nil {
			t.Fatalf("TransitionAlertStatus failed: %v", err)
		}
		if alert.Status != models.StatusInvestigating {
			t.Errorf("Expected status %s, got %s", models.StatusInvestigating, alert.Status)
		}
		if alert.AssignedTo != "investigator-1" {
			t.Errorf("Expected AssignedTo to be 'investigator-1', got %s", alert.AssignedTo)
		}
		if alert.TransitionAt.IsZero() {
			t.Errorf("Expected TransitionAt to be set")
		}
	})

	// Test Case 2: Valid: INVESTIGATING → FALSE_POSITIVE
	t.Run("valid_investigating_to_false_positive", func(t *testing.T) {
		alert := &models.Alert{Status: models.StatusInvestigating}
		err := TransitionAlertStatus(alert, models.StatusFalsePositive, "")
		if err != nil {
			t.Fatalf("TransitionAlertStatus failed: %v", err)
		}
		if alert.Status != models.StatusFalsePositive {
			t.Errorf("Expected status %s, got %s", models.StatusFalsePositive, alert.Status)
		}
	})

	// Test Case 3: Invalid: CLOSED → INVESTIGATING
	t.Run("invalid_closed_to_investigating", func(t *testing.T) {
		alert := &models.Alert{Status: models.StatusClosed}
		err := TransitionAlertStatus(alert, models.StatusInvestigating, "")
		if err == nil {
			t.Errorf("Expected error for invalid transition, but got nil")
		}
	})

	// Test Case 4: Invalid: OPEN → ESCALATED
	t.Run("invalid_open_to_escalated", func(t *testing.T) {
		alert := &models.Alert{Status: models.StatusOpen}
		err := TransitionAlertStatus(alert, models.StatusEscalated, "")
		if err == nil {
			t.Errorf("Expected error for invalid transition, but got nil")
		}
	})

	// Test Case 5: Valid: ESCALATED → CLOSED
	t.Run("valid_escalated_to_closed", func(t *testing.T) {
		alert := &models.Alert{Status: models.StatusEscalated}
		err := TransitionAlertStatus(alert, models.StatusClosed, "")
		if err != nil {
			t.Fatalf("TransitionAlertStatus failed: %v", err)
		}
		if alert.Status != models.StatusClosed {
			t.Errorf("Expected status %s, got %s", models.StatusClosed, alert.Status)
		}
	})
}
