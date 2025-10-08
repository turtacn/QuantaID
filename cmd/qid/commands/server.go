package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

// NewServerCmd creates the root `server` command and its subcommands.
// This command acts as a namespace for all server process management operations.
func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Manage the QuantaID server process",
		Long:  `Use the server command to start, stop, and check the status of the QuantaID server.`,
	}

	cmd.AddCommand(newServerStartCmd())

	return cmd
}

// newServerStartCmd creates the `server start` subcommand.
// This command is responsible for locating and launching the `qid-server` daemon process.
// It runs the server in the foreground, streaming its stdout and stderr, and waits for it to exit.
func newServerStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the QuantaID server",
		Long:  `Starts the QuantaID server daemon. By default, it runs in the foreground.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			serverPath, err := exec.LookPath("qid-server")
			if err != nil {
				return fmt.Errorf("could not find qid-server executable: %w. Ensure it is built and in your PATH", err)
			}

			fmt.Println("Starting QuantaID server...")

			serverCmd := exec.Command(serverPath)
			serverCmd.Stdout = os.Stdout
			serverCmd.Stderr = os.Stderr

			if err := serverCmd.Start(); err != nil {
				return fmt.Errorf("failed to start qid-server: %w", err)
			}

			fmt.Printf("QuantaID server started with PID: %d\n", serverCmd.Process.Pid)
			fmt.Println("Waiting for server to exit...")

			if err := serverCmd.Wait(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return fmt.Errorf("qid-server exited with error: %s", exitErr)
				}
				return fmt.Errorf("failed to wait for qid-server: %w", err)
			}

			fmt.Println("QuantaID server exited.")
			return nil
		},
	}
	return cmd
}

