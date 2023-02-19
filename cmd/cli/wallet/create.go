package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"blockchain/pkg/command"
	"blockchain/pkg/wallet"
)

var _ command.Cmd = (*createCmd)(nil)

type createCmd struct {
	baseCmd *cobra.Command
}

func (cmd *createCmd) GetCommand() *cobra.Command {
	return cmd.baseCmd
}

func newCreateCmd() command.Cmd {
	cmd := &createCmd{}

	baseCmd := &cobra.Command{
		Use:   "create",
		Short: "create a new wallet",
		RunE: func(_ *cobra.Command, args []string) error {
			address, err := wallet.CreateWallet()
			if err != nil {
				return err
			}

			fmt.Printf("created wallet with address: %s\n", address)
			return nil
		},
	}
	cmd.baseCmd = baseCmd

	return cmd
}
