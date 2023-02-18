package command

import "github.com/spf13/cobra"

type Cmd interface {
	GetCommand() *cobra.Command
}

type Builder struct {
	commands []Cmd
}

func (b *Builder) AddCommand(cmds ...Cmd) {
	b.commands = append(b.commands, cmds...)
}

func (b *Builder) Build(rootCmd *cobra.Command) {
	for _, cmd := range b.commands {
		rootCmd.AddCommand(cmd.GetCommand())
	}
}
