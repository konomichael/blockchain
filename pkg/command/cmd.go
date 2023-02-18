package command

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

type Cmd interface {
	GetCommand() *cobra.Command
}

type Builder struct {
	commands []Cmd
}

func (b *Builder) AddCommand(cmds ...Cmd) {
	validate := validator.New()
	for i := range cmds {
		c := cmds[i]
		baseCmd := c.GetCommand()
		if baseCmd.PreRunE == nil {
			baseCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				return validate.Struct(c)
			}
		}
	}

	b.commands = append(b.commands, cmds...)
}

func (b *Builder) Build(rootCmd *cobra.Command) {
	for _, cmd := range b.commands {
		rootCmd.AddCommand(cmd.GetCommand())
	}
}
