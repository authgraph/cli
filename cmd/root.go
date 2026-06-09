package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "authgraph",
	Short: "Authgraph — Permission Engine CLI",
	Long:  `Manage permissions, test access policies, and push schemas from the command line.`,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(grantCmd)
	rootCmd.AddCommand(revokeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(expandCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("authgraph %s\n", version)
	},
}
