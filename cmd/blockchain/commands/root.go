package commands

import (
	"os"

	"github.com/spf13/cobra"

	"blockchain/pkg/command"
)

var rootCmd = &cobra.Command{
	Use:   "blockchain",
	Short: "simple blockchain implementation",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
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
		newBalanceCmd(),
		newPrintCmd(),
		newSendCmd(),
	)
	b.Build(rootCmd)
}
