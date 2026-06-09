package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	authgraph "github.com/authgraph/authgraph-go"
	"github.com/authgraph/cli/internal/config"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check <subject_type:subject_id> <permission> <resource_type:resource_id>",
	Short: "Test a permission check",
	Long: `Check if a subject has a permission on a resource.

Examples:
  authgraph check user:alice read document:readme
  authgraph check user:bob write project:main --trace
  authgraph check service:agent-1 execute task:deploy --context '{"mfa":true}'`,
	Args: cobra.ExactArgs(3),
	RunE: runCheck,
}

var checkTrace bool

func init() {
	checkCmd.Flags().BoolVar(&checkTrace, "trace", false, "Show the traversal path")
}

func runCheck(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	subject, err := parseEntity(args[0])
	if err != nil {
		return fmt.Errorf("invalid subject: %w", err)
	}
	permission := args[1]
	resource, err := parseEntity(args[2])
	if err != nil {
		return fmt.Errorf("invalid resource: %w", err)
	}

	client, err := authgraph.NewClient(authgraph.Config{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.Check(ctx, authgraph.CheckRequest{
		Subject:    authgraph.Subject{Type: subject.Type, ID: subject.ID},
		Permission: permission,
		Resource:   authgraph.Resource{Type: resource.Type, ID: resource.ID},
		Trace:      checkTrace,
	})
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	if result.Allowed {
		fmt.Printf("✓ ALLOWED — %s:%s can %s %s:%s\n",
			subject.Type, subject.ID, permission, resource.Type, resource.ID)
	} else {
		fmt.Printf("✗ DENIED — %s:%s cannot %s %s:%s\n",
			subject.Type, subject.ID, permission, resource.Type, resource.ID)
	}

	if checkTrace && len(result.Path) > 0 {
		fmt.Println("\n  Path:")
		for _, step := range result.Path {
			fmt.Printf("    → %s\n", step)
		}
	}

	fmt.Printf("  Duration: %s\n", result.Duration)

	if !result.Allowed {
		os.Exit(1)
	}
	return nil
}

type entity struct {
	Type string
	ID   string
}

func parseEntity(s string) (*entity, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("expected format type:id, got %q", s)
	}
	return &entity{Type: parts[0], ID: parts[1]}, nil
}
