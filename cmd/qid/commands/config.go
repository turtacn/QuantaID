package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// NewConfigCmd creates the `config` command
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage QuantaID configuration",
		Long:  `Use the config command to validate, view, and manage the server configuration.`,
	}

	cmd.AddCommand(newConfigValidateCmd())

	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration file",
		Long:  `Parses the configuration file and checks for syntax errors and required fields.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")

			dummyLogger, _ := utils.NewZapLogger(&utils.LoggerConfig{
				Level:   "error",
				Console: utils.ConsoleConfig{Enabled: true},
			})

			_, err := utils.NewConfigManager(configPath, "server", "yaml", dummyLogger)
			if err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			fmt.Printf("Configuration at '%s' is valid.\n", configPath)
			return nil
		},
	}

	cmd.Flags().StringP("config", "c", "./configs", "Path to the configuration directory")

	return cmd
}

//Personal.AI order the ending
