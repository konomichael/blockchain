package commands

import (
	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
)

var _ command.Cmd = (*createCmd)(nil)

type createCmd struct {
	address string `validate:"required"` //btc address

	baseCmd *cobra.Command
}

func (cmd *createCmd) GetCommand() *cobra.Command {
	return cmd.baseCmd
}

func newCreateCmd() command.Cmd {
	cmd := &createCmd{}

	baseCmd := &cobra.Command{
		Use:   "create",
		Short: "creates a new blockchain",
		RunE: func(_ *cobra.Command, args []string) error {
			chain, err := blockchain.InitBlockChain(cmd.address)
			if err != nil {
				return err
			}
			defer chain.Close()

			return nil
		},
	}
	baseCmd.Flags().StringVar(&cmd.address, "address", "", "genesis wallet address")

	cmd.baseCmd = baseCmd
	return cmd
}
