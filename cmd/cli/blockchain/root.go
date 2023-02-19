package blockchain

import (
	"github.com/spf13/cobra"

	"blockchain/pkg/command"
)

var RootCmd = &cobra.Command{
	Use:   "blockchain",
	Short: "simple blockchain implementation",
}

func init() {
	b := &command.Builder{}
	b.AddCommand(
		newCreateCmd(),
		newBalanceCmd(),
		newPrintCmd(),
		newSendCmd(),
		newReindexCmd(),
	)
	b.Build(RootCmd)
}
