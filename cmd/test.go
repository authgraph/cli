package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	authgraph "github.com/authgraph/authgraph-go"
	"github.com/authgraph/cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run permission test assertions",
	Long: `Run a suite of permission test assertions from a YAML file.

The test file defines expected permission outcomes:

  tests:
    - name: "alice can read the readme"
      subject: "user:alice"
      permission: "read"
      resource: "document:readme"
      expected: allowed

    - name: "bob cannot delete the project"
      subject: "user:bob"
      permission: "delete"
      resource: "project:main"
      expected: denied

Examples:
  authgraph test --file permission-tests.yaml
  authgraph test -f tests/access.yaml`,
	RunE: runTest,
}

var testFile string

func init() {
	testCmd.Flags().StringVarP(&testFile, "file", "f", "", "Test file path (required)")
	testCmd.MarkFlagRequired("file")
}

// TestSuite represents a permission test file.
type TestSuite struct {
	Tests []TestCase `yaml:"tests"`
}

// TestCase represents a single permission assertion.
type TestCase struct {
	Name       string `yaml:"name"`
	Subject    string `yaml:"subject"`
	Permission string `yaml:"permission"`
	Resource   string `yaml:"resource"`
	Expected   string `yaml:"expected"` // "allowed" or "denied"
}

func runTest(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("reading test file: %w", err)
	}

	var suite TestSuite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return fmt.Errorf("parsing test file: %w", err)
	}

	if len(suite.Tests) == 0 {
		return fmt.Errorf("no tests found in %s", testFile)
	}

	client, err := authgraph.NewClient(authgraph.Config{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Timeout: 10 * time.Second,
		CacheEnabled: boolPtr(false), // Don't cache during tests
	})
	if err != nil {
		return err
	}

	fmt.Printf("Running %d permission tests...\n\n", len(suite.Tests))

	passed := 0
	failed := 0
	var failures []string

	for _, tc := range suite.Tests {
		subject, err := parseEntity(tc.Subject)
		if err != nil {
			failures = append(failures, fmt.Sprintf("  ✗ %s — invalid subject: %s", tc.Name, tc.Subject))
			failed++
			continue
		}
		resource, err := parseEntity(tc.Resource)
		if err != nil {
			failures = append(failures, fmt.Sprintf("  ✗ %s — invalid resource: %s", tc.Name, tc.Resource))
			failed++
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		result, err := client.Check(ctx, authgraph.CheckRequest{
			Subject:    authgraph.Subject{Type: subject.Type, ID: subject.ID},
			Permission: tc.Permission,
			Resource:   authgraph.Resource{Type: resource.Type, ID: resource.ID},
		})
		cancel()

		if err != nil {
			failures = append(failures, fmt.Sprintf("  ✗ %s — error: %v", tc.Name, err))
			failed++
			continue
		}

		expectAllowed := tc.Expected == "allowed" || tc.Expected == "allow" || tc.Expected == "true"

		if result.Allowed == expectAllowed {
			fmt.Printf("  ✓ %s\n", tc.Name)
			passed++
		} else {
			actual := "denied"
			if result.Allowed {
				actual = "allowed"
			}
			msg := fmt.Sprintf("  ✗ %s — expected %s, got %s", tc.Name, tc.Expected, actual)
			fmt.Println(msg)
			failures = append(failures, msg)
			failed++
		}
	}

	fmt.Printf("\n%d passed, %d failed, %d total\n", passed, failed, len(suite.Tests))

	if failed > 0 {
		fmt.Println("\nFailures:")
		for _, f := range failures {
			fmt.Println(f)
		}
		os.Exit(1)
	}

	fmt.Println("\n✓ All tests passed")
	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
