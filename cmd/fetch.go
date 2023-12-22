package cmd

import "github.com/spf13/cobra"

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch and save Strava race activities",
	Long: "fetch will request activities from Strava and \n." +
		"save the race activities.",
	Run: func(cmd *cobra.Command, args []string) {},
}
