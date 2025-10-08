package main

import (
	"fmt"
	"github.com/turtacn/QuantaID/cmd/qid/commands"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands.
// It's the entry point for the entire CLI application, configured using the Cobra library.
var (
	rootCmd = &cobra.Command{
		Use:   "qid",
		Short: "qid is the command-line interface for managing a QuantaID instance.",
		Long: `QuantaID CLI (qid) is a unified tool to manage your QuantaID server,
users, policies, and configurations from the command line.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// If the command execution fails, it prints the error to stderr and exits.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your command '%s'", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(commands.NewServerCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
}

func main() {
	Execute()
}

