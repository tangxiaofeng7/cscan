// Package template provides utilities for parsing Nuclei template YAML files.
package template

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Classification represents the vulnerability classification information from Nuclei templates.
// It contains CVSS metrics, CVE/CWE identifiers for vulnerability categorization.
type Classification struct {
	CvssMetrics string  `yaml:"cvss-metrics" json:"cvssMetrics,omitempty"` // CVSS vector string
	CvssScore   float64 `yaml:"cvss-score" json:"cvssScore,omitempty"`     // CVSS score (0-10)
	CveId       string  `yaml:"cve-id" json:"cveId,omitempty"`             // CVE identifier (e.g., CVE-2021-34473)
	CweId       string  `yaml:"cwe-id" json:"cweId,omitempty"`             // CWE identifier (e.g., CWE-79)
}

// TemplateInfo represents the info section of a Nuclei template.
// It contains metadata about the vulnerability including name, severity, references, and remediation.
type TemplateInfo struct {
	Name           string            `yaml:"name" json:"name,omitempty"`
	Author         string            `yaml:"author" json:"author,omitempty"`
	Severity       string            `yaml:"severity" json:"severity,omitempty"`
	Description    string            `yaml:"description" json:"description,omitempty"`
	Reference      []string          `yaml:"reference" json:"reference,omitempty"`
	Remediation    string            `yaml:"remediation" json:"remediation,omitempty"`
	Classification *Classification   `yaml:"classification" json:"classification,omitempty"`
	Tags           string            `yaml:"tags" json:"tags,omitempty"`
	Metadata       map[string]string `yaml:"metadata" json:"metadata,omitempty"`
}

// templateWrapper is used to extract only the info section from a Nuclei template.
type templateWrapper struct {
	Id   string        `yaml:"id"`
	Info *TemplateInfo `yaml:"info"`
}

// ParseTemplateInfo parses a Nuclei template YAML content and extracts the info section.
// It handles missing fields gracefully by returning zero values for missing data.
// Returns an error only if the YAML is malformed.
func ParseTemplateInfo(content string) (*TemplateInfo, error) {
	if content == "" {
		return &TemplateInfo{}, nil
	}

	var wrapper templateWrapper
	if err := yaml.Unmarshal([]byte(content), &wrapper); err != nil {
		return nil, err
	}

	// Return empty TemplateInfo if info section is missing
	if wrapper.Info == nil {
		return &TemplateInfo{}, nil
	}

	return wrapper.Info, nil
}

// GetCveIds extracts CVE IDs from the template info.
// It handles both single CVE ID in classification and multiple CVE IDs separated by commas.
// Returns an empty slice if no CVE IDs are found.
func (t *TemplateInfo) GetCveIds() []string {
	if t == nil || t.Classification == nil || t.Classification.CveId == "" {
		return nil
	}

	// Handle multiple CVE IDs separated by commas
	cveId := strings.TrimSpace(t.Classification.CveId)
	if cveId == "" {
		return nil
	}

	// Split by comma and clean up each ID
	parts := strings.Split(cveId, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}

	return result
}

// GetCweIds extracts CWE IDs from the template info.
// It handles both single CWE ID and multiple CWE IDs separated by commas.
// Returns an empty slice if no CWE IDs are found.
func (t *TemplateInfo) GetCweIds() []string {
	if t == nil || t.Classification == nil || t.Classification.CweId == "" {
		return nil
	}

	// Handle multiple CWE IDs separated by commas
	cweId := strings.TrimSpace(t.Classification.CweId)
	if cweId == "" {
		return nil
	}

	// Split by comma and clean up each ID
	parts := strings.Split(cweId, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}

	return result
}

// GetCvssScore returns the CVSS score from the template info.
// Returns 0 if no CVSS score is available.
func (t *TemplateInfo) GetCvssScore() float64 {
	if t == nil || t.Classification == nil {
		return 0
	}
	return t.Classification.CvssScore
}

// GetCvssMetrics returns the CVSS metrics string from the template info.
// Returns empty string if no CVSS metrics are available.
func (t *TemplateInfo) GetCvssMetrics() string {
	if t == nil || t.Classification == nil {
		return ""
	}
	return t.Classification.CvssMetrics
}

// GetReferences returns the reference URLs from the template info.
// Returns nil if no references are available.
func (t *TemplateInfo) GetReferences() []string {
	if t == nil {
		return nil
	}
	return t.Reference
}

// GetRemediation returns the remediation advice from the template info.
// Returns empty string if no remediation is available.
func (t *TemplateInfo) GetRemediation() string {
	if t == nil {
		return ""
	}
	return t.Remediation
}
