package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	authgraph "github.com/authgraph/authgraph-go"
	"github.com/authgraph/cli/internal/config"
	permstest "github.com/authgraph/cli/internal/testing"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run permission test assertions",
	Long: `Run a suite of permission test assertions from a YAML file.

The test file defines setup, assertions, and teardown:

  setup:
    schema:
      file: ./permissions.yaml
    tuples:
      - { subject: "user:alice", relation: "editor", resource: "document:readme" }

  tests:
    - name: "editors can write"
      check:
        subject: "user:alice"
        action: "write"
        resource: "document:readme"
      expect: allowed

    - name: "viewers cannot delete"
      check:
        subject: "user:bob"
        action: "delete"
        resource: "document:readme"
      expect: denied

  teardown:
    cleanup_setup: true

Use --what-if to simulate a schema change and detect regressions:

  authgraph test --file tests.yaml --what-if new-schema.yaml

Examples:
  authgraph test --file permission-tests.yaml
  authgraph test -f tests/access.yaml --output json
  authgraph test -f tests.yaml --output junit > results.xml
  authgraph test -f tests.yaml --what-if proposed-schema.yaml`,
	RunE: runTest,
}

var (
	testFile   string
	whatIfFile string
	outputFmt  string
	verbose    bool
)

func init() {
	testCmd.Flags().StringVarP(&testFile, "file", "f", "", "Test file path (required)")
	testCmd.Flags().StringVar(&whatIfFile, "what-if", "", "Simulate schema change and detect regressions")
	testCmd.Flags().StringVarP(&outputFmt, "output", "o", "text", "Output format: text, json, junit")
	testCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	testCmd.MarkFlagRequired("file")
}

func runTest(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	suite, err := permstest.LoadSuite(testFile)
	if err != nil {
		return err
	}

	client, err := authgraph.NewClient(authgraph.Config{
		BaseURL:      cfg.BaseURL,
		APIKey:       cfg.APIKey,
		Timeout:      10 * time.Second,
		CacheEnabled: boolPtr(false),
	})
	if err != nil {
		return err
	}

	ctx := context.Background()

	// What-if mode: simulate schema change and detect regressions
	if whatIfFile != "" {
		return runWhatIf(ctx, client, suite)
	}

	// Standard mode: run tests
	return runStandard(ctx, client, suite)
}

func runStandard(ctx context.Context, client *authgraph.Client, suite *permstest.Suite) error {
	runner := permstest.NewRunner(client, verbose)

	if outputFmt == "text" {
		fmt.Printf("Running %d permission tests...\n\n", len(suite.Tests))
	}

	report, err := runner.Run(ctx, suite)
	if err != nil {
		return err
	}

	switch outputFmt {
	case "json":
		if err := permstest.FormatJSON(os.Stdout, report); err != nil {
			return err
		}
	case "junit":
		if err := permstest.FormatJUnit(os.Stdout, report, testFile); err != nil {
			return err
		}
	default:
		permstest.FormatText(os.Stdout, report)
	}

	if report.Failed > 0 {
		os.Exit(1)
	}
	return nil
}

func runWhatIf(ctx context.Context, client *authgraph.Client, suite *permstest.Suite) error {
	whatIfRunner := permstest.NewWhatIfRunner(client, verbose)

	if outputFmt == "text" {
		fmt.Printf("Simulating schema change: %s\n", whatIfFile)
		fmt.Printf("Running %d tests for regression detection...\n\n", len(suite.Tests))
	}

	result, err := whatIfRunner.Simulate(ctx, whatIfFile, suite)
	if err != nil {
		return err
	}

	if !result.SchemaValid {
		fmt.Fprintf(os.Stderr, "✗ Schema validation failed:\n")
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		os.Exit(1)
	}

	switch outputFmt {
	case "json":
		if err := permstest.FormatJSON(os.Stdout, result.Report); err != nil {
			return err
		}
	case "junit":
		if err := permstest.FormatJUnit(os.Stdout, result.Report, testFile); err != nil {
			return err
		}
	default:
		permstest.FormatText(os.Stdout, result.Report)
		permstest.FormatRegressions(os.Stdout, result.Regressions)
	}

	if result.Report.Failed > 0 || len(result.Regressions) > 0 {
		os.Exit(1)
	}
	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
