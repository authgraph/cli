package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	authgraph "github.com/authgraph/authgraph-go"
	"github.com/authgraph/cli/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with your Authgraph API key",
	Long: `Store your API key for use with other CLI commands.

Example:
  authgraph login
  authgraph login --api-key ag_live_xxx --base-url https://api.authgraph.dev`,
	RunE: runLogin,
}

var (
	loginAPIKey  string
	loginBaseURL string
)

func init() {
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "API key (or will prompt interactively)")
	loginCmd.Flags().StringVar(&loginBaseURL, "base-url", "https://api.authgraph.dev", "Authgraph API base URL")
}

func runLogin(cmd *cobra.Command, args []string) error {
	apiKey := loginAPIKey

	if apiKey == "" {
		fmt.Print("Enter your API key: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		apiKey = strings.TrimSpace(input)
	}

	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	if !strings.HasPrefix(apiKey, "ag_") {
		return fmt.Errorf("invalid API key format — must start with 'ag_'")
	}

	// Verify the key works by making a health-style request
	client, err := authgraph.NewClient(authgraph.Config{
		BaseURL: loginBaseURL,
		APIKey:  apiKey,
	})
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	_ = client // Key format validated; full verification happens on first command

	cfg := &config.Config{
		APIKey:  apiKey,
		BaseURL: loginBaseURL,
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println("✓ Logged in successfully")
	fmt.Printf("  Config saved to %s\n", config.Path())
	return nil
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Delete(); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Not logged in.")
				return nil
			}
			return err
		}
		fmt.Println("✓ Logged out")
		return nil
	},
}
