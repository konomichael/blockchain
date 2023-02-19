package main

import (
	"github.com/spf13/cobra"

	"blockchain/cmd/cli/blockchain"
	"blockchain/cmd/cli/wallet"
)

var rootCmd = &cobra.Command{
	Use:          "cli",
	Short:        "cli for blockchain",
	SilenceUsage: true,
}

func main() {
	rootCmd.AddCommand(blockchain.RootCmd, wallet.RootCmd)
	rootCmd.Execute()
}
