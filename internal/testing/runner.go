package testing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	authgraph "github.com/authgraph/authgraph-go"
	"gopkg.in/yaml.v3"
)

// Runner executes a test suite against the Authgraph API.
type Runner struct {
	client  *authgraph.Client
	verbose bool
}

// NewRunner creates a new test runner.
func NewRunner(client *authgraph.Client, verbose bool) *Runner {
	return &Runner{client: client, verbose: verbose}
}

// LoadSuite reads and parses a test suite from a YAML file.
func LoadSuite(path string) (*Suite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading test file: %w", err)
	}

	var suite Suite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("parsing test file: %w", err)
	}

	if len(suite.Tests) == 0 {
		return nil, fmt.Errorf("no tests found in %s", path)
	}

	// Resolve relative schema file path
	if suite.Setup != nil && suite.Setup.Schema != nil && suite.Setup.Schema.File != "" {
		dir := filepath.Dir(path)
		suite.Setup.Schema.File = filepath.Join(dir, suite.Setup.Schema.File)
	}

	return &suite, nil
}

// Run executes the full test suite: setup → tests → teardown.
func (r *Runner) Run(ctx context.Context, suite *Suite) (*Report, error) {
	start := time.Now()

	// Setup phase
	if suite.Setup != nil {
		if err := r.runSetup(ctx, suite.Setup); err != nil {
			return nil, fmt.Errorf("setup failed: %w", err)
		}
	}

	// Test phase
	report := &Report{
		Total:   len(suite.Tests),
		Results: make([]TestResult, 0, len(suite.Tests)),
	}

	for _, tc := range suite.Tests {
		result := r.runTestCase(ctx, &tc)
		report.Results = append(report.Results, result)
		if result.Passed {
			report.Passed++
		} else {
			report.Failed++
		}
	}

	// Teardown phase
	if suite.Teardown != nil {
		r.runTeardown(ctx, suite)
	}

	report.DurationMs = float64(time.Since(start).Milliseconds())
	return report, nil
}

func (r *Runner) runSetup(ctx context.Context, setup *Setup) error {
	// Push schema if specified
	if setup.Schema != nil {
		var schemaContent string
		if setup.Schema.File != "" {
			data, err := os.ReadFile(setup.Schema.File)
			if err != nil {
				return fmt.Errorf("reading schema file %s: %w", setup.Schema.File, err)
			}
			schemaContent = string(data)
		} else if setup.Schema.Inline != "" {
			schemaContent = setup.Schema.Inline
		}

		if schemaContent != "" {
			if err := r.client.PushSchema(ctx, schemaContent); err != nil {
				return fmt.Errorf("pushing schema: %w", err)
			}
		}
	}

	// Create setup tuples
	for _, t := range setup.Tuples {
		subject, err := authgraph.ParseSubject(t.Subject)
		if err != nil {
			return fmt.Errorf("invalid subject %q: %w", t.Subject, err)
		}
		resource, err := authgraph.ParseResource(t.Resource)
		if err != nil {
			return fmt.Errorf("invalid resource %q: %w", t.Resource, err)
		}

		req := authgraph.GrantRequest{
			Subject:  subject,
			Relation: t.Relation,
			Resource: resource,
		}
		if t.ExpiresIn != "" {
			req.Condition = &authgraph.GrantCondition{
				ExpiresAt: t.ExpiresIn,
			}
		}
		if _, err := r.client.Grant(ctx, req); err != nil {
			return fmt.Errorf("creating tuple (%s %s %s): %w", t.Subject, t.Relation, t.Resource, err)
		}
	}

	return nil
}

func (r *Runner) runTestCase(ctx context.Context, tc *TestCase) TestResult {
	start := time.Now()
	result := TestResult{
		Name:     tc.Name,
		Expected: tc.GetExpected(),
	}

	subjectStr := tc.GetSubject()
	action := tc.GetAction()
	resourceStr := tc.GetResource()

	if subjectStr == "" || action == "" || resourceStr == "" {
		result.Error = "missing subject, action, or resource"
		result.Actual = "error"
		return result
	}

	subject, err := authgraph.ParseSubject(subjectStr)
	if err != nil {
		result.Error = fmt.Sprintf("invalid subject %q: %v", subjectStr, err)
		result.Actual = "error"
		return result
	}

	resource, err := authgraph.ParseResource(resourceStr)
	if err != nil {
		result.Error = fmt.Sprintf("invalid resource %q: %v", resourceStr, err)
		result.Actual = "error"
		return result
	}

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	checkResult, err := r.client.Check(checkCtx, authgraph.CheckRequest{
		Subject:    subject,
		Permission: action,
		Resource:   resource,
	})

	result.LatencyMs = float64(time.Since(start).Milliseconds())

	if err != nil {
		result.Error = err.Error()
		result.Actual = "error"
		return result
	}

	if checkResult.Allowed {
		result.Actual = "allowed"
	} else {
		result.Actual = "denied"
	}

	if len(checkResult.Path) > 0 {
		result.Path = strings.Join(checkResult.Path, " → ")
	}

	result.Passed = (checkResult.Allowed == tc.IsAllowed())
	return result
}

func (r *Runner) runTeardown(ctx context.Context, suite *Suite) {
	if suite.Teardown == nil {
		return
	}

	// If cleanup_setup is true, remove all setup tuples
	if suite.Teardown.CleanupSetup && suite.Setup != nil {
		for _, t := range suite.Setup.Tuples {
			subject, _ := authgraph.ParseSubject(t.Subject)
			resource, _ := authgraph.ParseResource(t.Resource)
			_ = r.client.Revoke(ctx, authgraph.RevokeRequest{
				Subject:  subject,
				Relation: t.Relation,
				Resource: resource,
			})
		}
	}

	// Remove explicit teardown tuples
	for _, t := range suite.Teardown.Tuples {
		subject, _ := authgraph.ParseSubject(t.Subject)
		resource, _ := authgraph.ParseResource(t.Resource)
		_ = r.client.Revoke(ctx, authgraph.RevokeRequest{
			Subject:  subject,
			Relation: t.Relation,
			Resource: resource,
		})
	}
}
