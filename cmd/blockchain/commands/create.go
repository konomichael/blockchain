package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
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
			chain, err := blockchain.InitBlockChain(cmd.Address)
			if err != nil {
				return err
			}
			defer chain.Close()

			fmt.Println("Created a new blockchain")
			return nil
		},
	}
	baseCmd.Flags().StringVar(&cmd.Address, "Address", "", "genesis wallet Address")

	cmd.baseCmd = baseCmd
	return cmd
}
