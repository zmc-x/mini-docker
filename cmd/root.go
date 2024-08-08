package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use: "mini-docker",
	Short: "mini-docker is a simple container implementation.",
	Run: func(cmd *cobra.Command, args []string) {},
}

func Execute() error {return rootCmd.Execute()}

func init() {
	rootCmd.AddCommand(initCmd, runCmd, commitCmd, psCmd, logCmd)
}