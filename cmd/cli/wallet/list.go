package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"blockchain/pkg/command"
	"blockchain/pkg/wallet"
)

var _ command.Cmd = (*listCmd)(nil)

type listCmd struct {
	baseCmd *cobra.Command
}

func (cmd *listCmd) GetCommand() *cobra.Command {
	return cmd.baseCmd
}

func newListCmd() command.Cmd {
	cmd := &listCmd{}

	baseCmd := &cobra.Command{
		Use:   "list",
		Short: "lists all wallets",
		RunE: func(_ *cobra.Command, args []string) error {
			addresses := wallet.GetAllAddresses()
			for _, address := range addresses {
				fmt.Println(address)
			}

			return nil
		},
	}

	cmd.baseCmd = baseCmd
	return cmd
}
