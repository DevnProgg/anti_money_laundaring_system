package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// Rule defines the structure for an AML threshold rule.
type Rule struct {
	RuleID         string  `json:"rule_id"`
	Name           string  `json:"name"`
	ThresholdValue float64 `json:"threshold_value"`
	TimeWindow     string  `json:"time_window"`
	Enabled        bool    `json:"enabled"`
}

// GetTimeWindow returns the parsed time duration for the rule.
func (r *Rule) GetTimeWindow() (time.Duration, error) {
	return time.ParseDuration(r.TimeWindow)
}

// LoadRules loads and parses AML threshold rules from a JSON file.
func LoadRules(filepath string) ([]Rule, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rule file: %w", err)
	}

	var rules []Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse rule file: %w", err)
	}

	if err := validateRules(rules); err != nil {
		return nil, fmt.Errorf("rule validation failed: %w", err)
	}

	return rules, nil
}

// validateRules validates the loaded AML threshold rules.
func validateRules(rules []Rule) error {
	ruleIDs := make(map[string]bool)
	for _, rule := range rules {
		if rule.ThresholdValue <= 0 {
			return fmt.Errorf("threshold_value must be > 0 for rule '%s'", rule.RuleID)
		}
		if _, err := time.ParseDuration(rule.TimeWindow); err != nil {
			return fmt.Errorf("invalid time_window for rule '%s'", rule.RuleID)
		}
		if _, exists := ruleIDs[rule.RuleID]; exists {
			return fmt.Errorf("duplicate rule_id: '%s'", rule.RuleID)
		}
		ruleIDs[rule.RuleID] = true
	}
	return nil
}
