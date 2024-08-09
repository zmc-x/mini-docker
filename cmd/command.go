package cmd

import (
	"fmt"
	"mini-docker/cgroup/subsystems"
	"mini-docker/container"
	"mini-docker/runtime"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	runCmd = &cobra.Command{
		Use:   "run [Command]",
		Short: "run command creates container with Namespace and Cgroup",
		RunE: func(cmd *cobra.Command, args []string) error {
			// check --ti and --d
			if ti && daemon {
				return fmt.Errorf("ti and d paramter can't both provided")
			}
			cfg := &subsystems.ResourceConfig{
				MemoryLimit: m,
				CpuSet:      cpuset,
				CpuShare:    cpushare,
			}
			runtime.Run(ti, volume, args, cfg, name)
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}

	commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "commit container into image",
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime.CommitContainer(args[0])
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}

	initCmd = &cobra.Command{
		Use:    "init",
		Short:  "init command init the container, don't call outside",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return container.ContainerInit()
		},
	}

	psCmd = &cobra.Command{
		Use: "ps",
		Short: "list the container",
		Run: func(cmd *cobra.Command, args []string) {
			container.ListContainer()
		},
	}

	logCmd = &cobra.Command{
		Use: "log",
		Short: "print logs of the container",
		Run: func(cmd *cobra.Command, args []string) {
			container.GetContainerLog(args[0])
		},
		Args: cobra.MinimumNArgs(1),
	}

	execCmd = &cobra.Command{
		Use: "exec",
		Short: "exec a command into container",
		RunE: func(cmd *cobra.Command, args []string) error {
			if os.Getenv(container.ENV_EXEC_PID) != "" {
				zap.L().Sugar().Infof("pid call back pid %d", os.Getpid())
				return nil
			}
			// check args length
			if len(args) < 2 {
				return fmt.Errorf("missing container name or command")
			}
			container.ExecContainer(args[0], args[1: ])
			return nil
		},
	}

	stopCmd = &cobra.Command{
		Use: "stop",
		Short: "stop the container",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			container.StopContainer(args[0])
		},
	}
)

var (
	// generic
	ti     bool
	volume string
	daemon bool
	name   string
	// cgroup
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
	runCmd.Flags().StringVarP(&volume, "volume", "v", "", "set the volume of the container")
	runCmd.Flags().BoolVar(&daemon, "d", false, "detach container")
	runCmd.Flags().StringVar(&name, "name", "", "set the container name")
}
