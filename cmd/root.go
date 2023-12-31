package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "races"}

func Execute() error {
	rootCmd.AddCommand(newTokenCmd, fetchCmd, genHtmlCmd)
	return rootCmd.Execute()
}
