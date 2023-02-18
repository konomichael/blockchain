package commands

import (
	"os"

	"github.com/spf13/cobra"

	"blockchain/pkg/command"
)

var rootCmd = &cobra.Command{
	Use:   "wallet",
	Short: "simple wallet implementation",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		rootCmd.Usage()
		os.Exit(1)
	}
}

func init() {
	b := &command.Builder{}
	b.AddCommand(
		newCreateCmd(),
		newListCmd(),
	)
	b.Build(rootCmd)
}
