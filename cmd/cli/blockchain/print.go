package blockchain

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
)

var _ command.Cmd = (*printCmd)(nil)

type printCmd struct {
	baseCmd *cobra.Command
}

func (cmd *printCmd) GetCommand() *cobra.Command {
	return cmd.baseCmd
}

func newPrintCmd() command.Cmd {
	cmd := &printCmd{}

	baseCmd := &cobra.Command{
		Use:   "print",
		Short: "prints the blockchain",
		RunE: func(_ *cobra.Command, args []string) error {
			chain, err := blockchain.ContinueBlockChain()
			if err != nil {
				return nil
			}
			defer chain.Close()

			iter := chain.Iterator()
			for iter.HasNext() {
				block := iter.Next()
				fmt.Printf("Prev. hash: %x\n", block.PrevHash)
				fmt.Printf("Hash: %x\n", block.Hash)
				pow := blockchain.NewProof(block)
				fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
				for _, tx := range block.Transactions {
					fmt.Println(tx)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.baseCmd = baseCmd
	return cmd
}
