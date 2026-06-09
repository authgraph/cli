package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/authgraph/cli/internal/config"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage permission schemas",
	Long:  `Push, validate, and view permission schemas.`,
}

var schemaPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push a schema to Authgraph",
	Long: `Upload a permission schema YAML file to Authgraph.

Examples:
  authgraph schema push --file permissions.yaml
  authgraph schema push -f ./schemas/production.yaml`,
	RunE: runSchemaPush,
}

var schemaValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a schema without deploying",
	Long: `Check a schema file for errors without pushing it.

Examples:
  authgraph schema validate --file permissions.yaml`,
	RunE: runSchemaValidate,
}

var schemaGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current active schema",
	RunE:  runSchemaGet,
}

var schemaFile string

func init() {
	schemaCmd.AddCommand(schemaPushCmd)
	schemaCmd.AddCommand(schemaValidateCmd)
	schemaCmd.AddCommand(schemaGetCmd)

	schemaPushCmd.Flags().StringVarP(&schemaFile, "file", "f", "", "Schema YAML file path (required)")
	schemaPushCmd.MarkFlagRequired("file")

	schemaValidateCmd.Flags().StringVarP(&schemaFile, "file", "f", "", "Schema YAML file path (required)")
	schemaValidateCmd.MarkFlagRequired("file")
}

func runSchemaPush(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}

	body := map[string]interface{}{
		"schema": string(content),
	}
	jsonBody, _ := json.Marshal(body)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "POST", cfg.BaseURL+"/v1/schemas", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("pushing schema: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		json.Unmarshal(respBody, &apiErr)
		return fmt.Errorf("push failed [%d]: %s — %s", resp.StatusCode, apiErr.Code, apiErr.Message)
	}

	var result struct {
		Version int `json:"version"`
	}
	json.Unmarshal(respBody, &result)

	fmt.Printf("✓ Schema pushed successfully (version %d)\n", result.Version)
	return nil
}

func runSchemaValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}

	body := map[string]interface{}{
		"schema": string(content),
	}
	jsonBody, _ := json.Marshal(body)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "POST", cfg.BaseURL+"/v1/schemas/validate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("validating schema: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Code    string   `json:"code"`
			Message string   `json:"message"`
			Errors  []string `json:"errors"`
		}
		json.Unmarshal(respBody, &apiErr)
		fmt.Printf("✗ Schema validation failed:\n")
		if len(apiErr.Errors) > 0 {
			for _, e := range apiErr.Errors {
				fmt.Printf("  • %s\n", e)
			}
		} else {
			fmt.Printf("  • %s\n", apiErr.Message)
		}
		os.Exit(1)
	}

	fmt.Println("✓ Schema is valid")
	return nil
}

func runSchemaGet(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", cfg.BaseURL+"/v1/schemas", nil)
	req.Header.Set("X-API-Key", cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("getting schema: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to get schema [%d]", resp.StatusCode)
	}

	var result struct {
		Version int    `json:"version"`
		Schema  string `json:"schema"`
	}
	json.Unmarshal(respBody, &result)

	fmt.Printf("# Schema (version %d)\n", result.Version)
	fmt.Println(result.Schema)
	return nil
}
