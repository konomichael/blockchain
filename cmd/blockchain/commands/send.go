package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
)

var _ command.Cmd = (*sendCmd)(nil)

type sendCmd struct {
	From   string `validate:"required"`
	To     string `validate:"required"`
	Amount int    `validate:"gte=0"`

	baseCmd *cobra.Command
}

func (cmd *sendCmd) GetCommand() *cobra.Command {
	return cmd.baseCmd
}

func newSendCmd() command.Cmd {
	cmd := &sendCmd{}

	baseCmd := &cobra.Command{
		Use:   "send",
		Short: "send amount from one wallet to another",
		RunE: func(_ *cobra.Command, args []string) error {
			chain, err := blockchain.ContinueBlockChain()
			if err != nil {
				return err
			}
			defer chain.Close()

			tx, err := blockchain.NewTransaction(cmd.From, cmd.To, cmd.Amount, chain)
			if err != nil {
				return err
			}

			err = chain.AddBlock([]*blockchain.Transaction{tx})
			if err != nil {
				return err
			}

			fmt.Println("Sent", cmd.Amount, "from", cmd.From, "to", cmd.To)
			return nil
		},
	}
	baseCmd.Flags().StringVar(&cmd.From, "from", "", "source wallet Address")
	baseCmd.Flags().StringVar(&cmd.To, "to", "", "destination wallet Address")
	baseCmd.Flags().IntVar(&cmd.Amount, "amount", 0, "amount to send")

	cmd.baseCmd = baseCmd
	return cmd
}
