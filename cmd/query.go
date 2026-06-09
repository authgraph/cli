package cmd

import (
	"context"
	"fmt"
	"time"

	authgraph "github.com/authgraph/authgraph-go"
	"github.com/authgraph/cli/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <subject_type:subject_id> <permission> [resource_type]",
	Short: "List resources a subject can access",
	Long: `List all resources a subject has a specific permission on.

Examples:
  authgraph list user:alice read
  authgraph list user:alice read document
  authgraph list group:admins manage --limit 50`,
	Args: cobra.RangeArgs(2, 3),
	RunE: runList,
}

var listLimit int

func init() {
	listCmd.Flags().IntVar(&listLimit, "limit", 100, "Maximum results to return")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	subject, err := parseEntity(args[0])
	if err != nil {
		return fmt.Errorf("invalid subject: %w", err)
	}
	permission := args[1]
	resourceType := ""
	if len(args) > 2 {
		resourceType = args[2]
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

	resp, err := client.ListResources(ctx, authgraph.ListResourcesRequest{
		Subject:      authgraph.Subject{Type: subject.Type, ID: subject.ID},
		Permission:   permission,
		ResourceType: resourceType,
		Limit:        listLimit,
	})
	if err != nil {
		return fmt.Errorf("list failed: %w", err)
	}

	if len(resp.Resources) == 0 {
		fmt.Printf("No resources found for %s:%s with permission %q\n", subject.Type, subject.ID, permission)
		return nil
	}

	fmt.Printf("Resources accessible by %s:%s (%s):\n\n", subject.Type, subject.ID, permission)
	for _, r := range resp.Resources {
		fmt.Printf("  %s:%s  (via %s)\n", r.ResourceType, r.ResourceID, r.Relation)
	}
	fmt.Printf("\n  Total: %d\n", resp.Page.Total)
	return nil
}

var expandCmd = &cobra.Command{
	Use:   "expand <resource_type:resource_id> <permission>",
	Short: "List subjects that have access to a resource",
	Long: `Expand a resource to see all subjects that have a specific permission.

Examples:
  authgraph expand document:readme read
  authgraph expand project:main admin
  authgraph expand org:acme owner`,
	Args: cobra.ExactArgs(2),
	RunE: runExpand,
}

var expandLimit int

func init() {
	expandCmd.Flags().IntVar(&expandLimit, "limit", 100, "Maximum results to return")
}

func runExpand(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	resource, err := parseEntity(args[0])
	if err != nil {
		return fmt.Errorf("invalid resource: %w", err)
	}
	permission := args[1]

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

	resp, err := client.ListSubjects(ctx, authgraph.ListSubjectsRequest{
		Resource:   authgraph.Resource{Type: resource.Type, ID: resource.ID},
		Permission: permission,
		Limit:      expandLimit,
	})
	if err != nil {
		return fmt.Errorf("expand failed: %w", err)
	}

	if len(resp.Subjects) == 0 {
		fmt.Printf("No subjects have %q on %s:%s\n", permission, resource.Type, resource.ID)
		return nil
	}

	fmt.Printf("Subjects with %q on %s:%s:\n\n", permission, resource.Type, resource.ID)
	for _, s := range resp.Subjects {
		fmt.Printf("  %s:%s  (via %s)\n", s.SubjectType, s.SubjectID, s.Relation)
	}
	fmt.Printf("\n  Total: %d\n", resp.Page.Total)
	return nil
}
