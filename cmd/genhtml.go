package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var genHtmlCmd = &cobra.Command{
	Use:   "genhtml",
	Short: "Generate html for saved race activities",
	Long:  "genhtml will generate static html for saved race activities.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented")
	},
}
