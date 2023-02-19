package blockchain

import (
	"fmt"

	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
)

var _ command.Cmd = (*reindexCmd)(nil)

type reindexCmd struct {
	baseCmd *cobra.Command
}

func (cmd *reindexCmd) GetCommand() *cobra.Command {
	return cmd.baseCmd
}

func newReindexCmd() command.Cmd {
	cmd := &reindexCmd{}

	baseCmd := &cobra.Command{
		Use:   "reindex",
		Short: "reindex rebuilds the UTXO set",
		RunE: func(_ *cobra.Command, args []string) error {
			chain, err := blockchain.ContinueBlockChain()
			if err != nil {
				return err
			}
			defer chain.Close()

			utxoSet := blockchain.NewUTXOSet(chain)
			if err = utxoSet.Reindex(); err != nil {
				return err
			}

			count, err := utxoSet.CountTransactions()
			if err != nil {
				return err
			}

			fmt.Println("Done! There are", count, "transactions in the UTXO set.")
			return nil
		},
	}

	cmd.baseCmd = baseCmd
	return cmd
}
