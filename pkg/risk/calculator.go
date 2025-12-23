// Package risk provides risk calculation functionality for assets based on vulnerabilities.
package risk

import (
	"math"
)

// VulInfo represents the vulnerability information needed for risk calculation.
// This is a simplified struct to avoid circular dependencies with the model package.
type VulInfo struct {
	Severity  string  // critical, high, medium, low, info
	CvssScore float64 // 0.0 - 10.0
}

// SeverityWeight defines the weight for each severity level.
// Used in risk score calculation.
var SeverityWeight = map[string]float64{
	"critical": 10.0,
	"high":     7.0,
	"medium":   4.0,
	"low":      1.0,
	"info":     0.1,
}

// RiskLevel constants
const (
	RiskLevelCritical = "critical"
	RiskLevelHigh     = "high"
	RiskLevelMedium   = "medium"
	RiskLevelLow      = "low"
	RiskLevelInfo     = "info"
)

// Risk level thresholds
const (
	ThresholdCritical = 80.0
	ThresholdHigh     = 60.0
	ThresholdMedium   = 40.0
	ThresholdLow      = 20.0
)

// RiskCalculator calculates risk scores for assets based on their vulnerabilities.
type RiskCalculator struct{}

// NewRiskCalculator creates a new RiskCalculator instance.
func NewRiskCalculator() *RiskCalculator {
	return &RiskCalculator{}
}

// CalculateRiskScore calculates the risk score for an asset based on its vulnerabilities.
// The score is calculated using the following formula:
//   - baseScore = maxCvssScore * 10 (highest CVSS score * 10)
//   - vulCountBonus = sum of severity weights for all vulnerabilities
//   - score = min(100, baseScore + vulCountBonus)
//
// Returns a score between 0 and 100.
func (c *RiskCalculator) CalculateRiskScore(vuls []VulInfo) float64 {
	if len(vuls) == 0 {
		return 0.0
	}

	// Find the maximum CVSS score
	maxCvssScore := 0.0
	for _, vul := range vuls {
		if vul.CvssScore > maxCvssScore {
			maxCvssScore = vul.CvssScore
		}
	}

	// Calculate base score from max CVSS (0-10 scale to 0-100)
	baseScore := maxCvssScore * 10.0

	// Calculate vulnerability count bonus based on severity weights
	vulCountBonus := 0.0
	for _, vul := range vuls {
		weight, ok := SeverityWeight[vul.Severity]
		if !ok {
			// Default weight for unknown severity
			weight = 0.1
		}
		vulCountBonus += weight
	}

	// Calculate final score, capped at 100
	score := baseScore + vulCountBonus
	return math.Min(100.0, score)
}

// GetRiskLevel returns the risk level based on the risk score.
// Risk level mapping:
//   - critical: score >= 80
//   - high:     score >= 60
//   - medium:   score >= 40
//   - low:      score >= 20
//   - info:     score < 20
func (c *RiskCalculator) GetRiskLevel(score float64) string {
	switch {
	case score >= ThresholdCritical:
		return RiskLevelCritical
	case score >= ThresholdHigh:
		return RiskLevelHigh
	case score >= ThresholdMedium:
		return RiskLevelMedium
	case score >= ThresholdLow:
		return RiskLevelLow
	default:
		return RiskLevelInfo
	}
}

// CalculateRiskScoreAndLevel is a convenience method that calculates both
// the risk score and risk level in one call.
func (c *RiskCalculator) CalculateRiskScoreAndLevel(vuls []VulInfo) (float64, string) {
	score := c.CalculateRiskScore(vuls)
	level := c.GetRiskLevel(score)
	return score, level
}
