package cmd

import (
	"context"
	"fmt"
	"time"

	authgraph "github.com/authgraph/authgraph-go"
	"github.com/authgraph/cli/internal/config"
	"github.com/spf13/cobra"
)

var grantCmd = &cobra.Command{
	Use:   "grant <subject_type:subject_id> <relation> <resource_type:resource_id>",
	Short: "Grant a permission (create a relationship tuple)",
	Long: `Create a relationship tuple granting a subject a relation on a resource.

Examples:
  authgraph grant user:alice reader document:readme
  authgraph grant user:bob editor project:main --expires 2026-12-31T23:59:59Z
  authgraph grant group:admins owner org:acme`,
	Args: cobra.ExactArgs(3),
	RunE: runGrant,
}

var (
	grantExpires  string
	grantValidFrom string
)

func init() {
	grantCmd.Flags().StringVar(&grantExpires, "expires", "", "Expiration time (RFC3339)")
	grantCmd.Flags().StringVar(&grantValidFrom, "valid-from", "", "Valid from time (RFC3339)")
}

func runGrant(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	subject, err := parseEntity(args[0])
	if err != nil {
		return fmt.Errorf("invalid subject: %w", err)
	}
	relation := args[1]
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

	req := authgraph.GrantRequest{
		Subject:  authgraph.Subject{Type: subject.Type, ID: subject.ID},
		Relation: relation,
		Resource: authgraph.Resource{Type: resource.Type, ID: resource.ID},
	}

	if grantExpires != "" || grantValidFrom != "" {
		req.Condition = &authgraph.GrantCondition{
			ExpiresAt: grantExpires,
			ValidFrom: grantValidFrom,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Grant(ctx, req)
	if err != nil {
		return fmt.Errorf("grant failed: %w", err)
	}

	fmt.Printf("✓ Granted %s:%s → %s → %s:%s\n",
		subject.Type, subject.ID, relation, resource.Type, resource.ID)
	fmt.Printf("  Tuple ID: %s\n", resp.ID)
	if grantExpires != "" {
		fmt.Printf("  Expires: %s\n", grantExpires)
	}
	return nil
}

var revokeCmd = &cobra.Command{
	Use:   "revoke <subject_type:subject_id> <relation> <resource_type:resource_id>",
	Short: "Revoke a permission (delete a relationship tuple)",
	Long: `Delete a relationship tuple, removing a subject's relation on a resource.

Examples:
  authgraph revoke user:alice reader document:readme
  authgraph revoke user:bob editor project:main`,
	Args: cobra.ExactArgs(3),
	RunE: runRevoke,
}

func runRevoke(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	subject, err := parseEntity(args[0])
	if err != nil {
		return fmt.Errorf("invalid subject: %w", err)
	}
	relation := args[1]
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

	err = client.Revoke(ctx, authgraph.RevokeRequest{
		Subject:  authgraph.Subject{Type: subject.Type, ID: subject.ID},
		Relation: relation,
		Resource: authgraph.Resource{Type: resource.Type, ID: resource.ID},
	})
	if err != nil {
		return fmt.Errorf("revoke failed: %w", err)
	}

	fmt.Printf("✓ Revoked %s:%s → %s → %s:%s\n",
		subject.Type, subject.ID, relation, resource.Type, resource.ID)
	return nil
}
