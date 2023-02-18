package commands

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"

	"blockchain/pkg/blockchain"
	"blockchain/pkg/command"
)

var _ command.Cmd = (*sendCmd)(nil)

type sendCmd struct {
	from   string `validate:"required"`
	to     string `validate:"required"`
	amount int    `validate:"required,gte=0"`

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
			validate := validator.New()
			if err := validate.Struct(cmd); err != nil {
				return err
			}

			chain, err := blockchain.ContinueBlockChain()
			if err != nil {
				return err
			}
			defer chain.Close()

			tx, err := blockchain.NewTransaction(cmd.from, cmd.to, cmd.amount, chain)
			if err != nil {
				return err
			}

			err = chain.AddBlock([]*blockchain.Transaction{tx})
			if err != nil {
				return err
			}

			return nil
		},
	}
	baseCmd.Flags().StringVar(&cmd.from, "from", "", "source wallet address")
	baseCmd.Flags().StringVar(&cmd.to, "to", "", "destination wallet address")
	baseCmd.Flags().IntVar(&cmd.amount, "amount", 0, "amount to send")

	cmd.baseCmd = baseCmd
	return cmd
}
