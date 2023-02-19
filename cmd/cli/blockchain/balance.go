package blockchain

import (
	"fmt"

	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
	"blockchain/pkg/wallet"
)

var _ command.Cmd = (*balanceCmd)(nil)

type balanceCmd struct {
	Address string `validate:"required"` //btc address

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
			pubKeyHash, err := wallet.PubKeyHashFromAddress(cmd.Address)
			if err != nil {
				return err
			}

			chain, err := blockchain.ContinueBlockChain()
			if err != nil {
				return err
			}
			defer chain.Close()

			utxoSet := blockchain.NewUTXOSet(chain)

			balance := 0
			UTXOs, err := utxoSet.FindUTXOs(pubKeyHash)
			if err != nil {
				return err
			}

			for _, out := range *UTXOs {
				balance += out.Value
			}
			fmt.Println("Balance of", cmd.Address, "is", balance)
			return nil
		},
	}
	baseCmd.Flags().StringVar(&cmd.Address, "address", "", "wallet address")

	cmd.baseCmd = baseCmd
	return cmd
}
