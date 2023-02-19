package blockchain

import (
	"fmt"

	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
	"blockchain/pkg/wallet"
)

var _ command.Cmd = (*createCmd)(nil)

type createCmd struct {
	Address string `validate:"required"` //btc address

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
			_, err := wallet.PubKeyHashFromAddress(cmd.Address)
			if err != nil {
				return err
			}

			chain, err := blockchain.InitBlockChain(cmd.Address)
			if err != nil {
				return err
			}
			defer chain.Close()

			utxoSet := blockchain.NewUTXOSet(chain)
			if err = utxoSet.Reindex(); err != nil {
				return err
			}

			fmt.Println("Created a new blockchain")
			return nil
		},
	}
	baseCmd.Flags().StringVar(&cmd.Address, "address", "", "genesis wallet Address")

	cmd.baseCmd = baseCmd
	return cmd
}
