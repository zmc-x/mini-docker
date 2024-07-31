package cmd

import (
	"mini-docker/container"
	"mini-docker/runtime"

	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use: "run [Command]",
		Short: "run command creates container with Namespace and Cgroup",
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime.Run(ti, args)
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}

	initCmd = &cobra.Command{
		Use: "init",
		Short: "init command init the container, don't call outside",
		RunE: func(cmd *cobra.Command, args []string) error {
			return container.ContainerInit()
		},
	}
)

var (
	ti bool
)

func init() {
	// flag
	runCmd.Flags().BoolVar(&ti, "ti", false, "enable tty")
}