package testing

import (
	"context"
	"fmt"
	"os"

	authgraph "github.com/authgraph/authgraph-go"
)

// WhatIfRunner simulates schema changes and runs tests against the simulation.
type WhatIfRunner struct {
	client  *authgraph.Client
	verbose bool
}

// NewWhatIfRunner creates a runner for what-if schema simulation.
func NewWhatIfRunner(client *authgraph.Client, verbose bool) *WhatIfRunner {
	return &WhatIfRunner{client: client, verbose: verbose}
}

// WhatIfResult holds the result of a what-if simulation.
type WhatIfResult struct {
	SchemaValid bool         `json:"schema_valid"`
	Errors      []string     `json:"errors,omitempty"`
	Report      *Report      `json:"report,omitempty"`
	Regressions []Regression `json:"regressions,omitempty"`
}

// Regression represents a test that would change behavior after a schema change.
type Regression struct {
	TestName string `json:"test_name"`
	Before   string `json:"before"` // "allowed" or "denied"
	After    string `json:"after"`  // "allowed" or "denied"
}

// Simulate validates a new schema and runs the test suite against it to find regressions.
// It compares current results (before schema change) with results after the change.
func (w *WhatIfRunner) Simulate(ctx context.Context, schemaFile string, suite *Suite) (*WhatIfResult, error) {
	result := &WhatIfResult{}

	// Read the proposed schema
	schemaData, err := os.ReadFile(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("reading schema file: %w", err)
	}
	schemaContent := string(schemaData)

	// Step 1: Validate the proposed schema
	validationResult, err := w.client.ValidateSchema(ctx, schemaContent)
	if err != nil {
		return nil, fmt.Errorf("validating schema: %w", err)
	}

	if !validationResult.Valid {
		result.SchemaValid = false
		result.Errors = validationResult.Errors
		return result, nil
	}
	result.SchemaValid = true

	// Step 2: Run tests against the CURRENT schema (baseline)
	baselineRunner := NewRunner(w.client, w.verbose)
	baselineReport, err := baselineRunner.Run(ctx, suite)
	if err != nil {
		return nil, fmt.Errorf("running baseline tests: %w", err)
	}

	// Step 3: Push the new schema
	if err := w.client.PushSchema(ctx, schemaContent); err != nil {
		return nil, fmt.Errorf("pushing proposed schema: %w", err)
	}

	// Step 4: Run tests against the NEW schema
	newRunner := NewRunner(w.client, w.verbose)
	newReport, err := newRunner.Run(ctx, suite)
	if err != nil {
		return nil, fmt.Errorf("running tests against new schema: %w", err)
	}

	result.Report = newReport

	// Step 5: Detect regressions (tests that changed outcome)
	for i, baseResult := range baselineReport.Results {
		if i >= len(newReport.Results) {
			break
		}
		newResult := newReport.Results[i]
		if baseResult.Actual != newResult.Actual && baseResult.Actual != "error" && newResult.Actual != "error" {
			result.Regressions = append(result.Regressions, Regression{
				TestName: baseResult.Name,
				Before:   baseResult.Actual,
				After:    newResult.Actual,
			})
		}
	}

	return result, nil
}
