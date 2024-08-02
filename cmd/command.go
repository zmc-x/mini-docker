package cmd

import (
	"mini-docker/cgroup/subsystems"
	"mini-docker/container"
	"mini-docker/runtime"

	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:   "run [Command]",
		Short: "run command creates container with Namespace and Cgroup",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &subsystems.ResourceConfig{
				MemoryLimit: m,
				CpuSet:      cpuset,
				CpuShare:    cpushare,
			}
			runtime.Run(ti, args, cfg)
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "init command init the container, don't call outside",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return container.ContainerInit()
		},
	}
)

var (
	ti       bool
	m        string
	cpuset   string
	cpushare string
)

func init() {
	// flag
	runCmd.Flags().BoolVar(&ti, "ti", false, "enable tty")
	runCmd.Flags().StringVar(&m, "m", "", "set memory limit")
	runCmd.Flags().StringVar(&cpuset, "cpuset", "", "set the cgroup process can be used in the CPU and memory")
	runCmd.Flags().StringVar(&cpushare, "cpushare", "", "set the cpu schedule for the processes in cgroup")
}
