package commands

import (
	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
)

var _ command.Cmd = (*balanceCmd)(nil)

type balanceCmd struct {
	address string `validate:"required"` //btc address

	baseCmd *cobra.Command
}

func (cmd *balanceCmd) GetCommand() *cobra.Command {
	return cmd.baseCmd
}

func newBalanceCmd() command.Cmd {
	cmd := &balanceCmd{}

	baseCmd := &cobra.Command{
		Use:   "balance",
		Short: "get balance of a wallet",
		RunE: func(_ *cobra.Command, args []string) error {
			chain, err := blockchain.ContinueBlockChain()
			if err != nil {
				return err
			}
			defer chain.Close()

			balance := 0
			UTXOs := chain.FindUTXO(cmd.address)
			for _, out := range UTXOs {
				balance += out.Value
			}

			return nil
		},
	}
	baseCmd.Flags().StringVar(&cmd.address, "address", "", "wallet address")

	cmd.baseCmd = baseCmd
	return cmd
}
