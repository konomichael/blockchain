package wallet

import (
	"github.com/spf13/cobra"

	"blockchain/pkg/command"
)

var RootCmd = &cobra.Command{
	Use:   "wallet",
	Short: "simple wallet implementation",
}

func init() {
	b := &command.Builder{}
	b.AddCommand(
		newCreateCmd(),
		newListCmd(),
	)
	b.Build(RootCmd)
}
