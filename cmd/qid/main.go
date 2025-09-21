package main

import (
	"fmt"
	"github.com/turtacn/QuantaID/cmd/qid/commands"
	"github.com/spf13/cobra"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:   "qid",
		Short: "qid is the command-line interface for managing a QuantaID instance.",
		Long: `QuantaID CLI (qid) is a unified tool to manage your QuantaID server,
users, policies, and configurations from the command line.`,
	}
)

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

//Personal.AI order the ending
