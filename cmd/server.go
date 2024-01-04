package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "http server for saved race activities",
	Long:  "server will start an http server for saved race activities.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}
