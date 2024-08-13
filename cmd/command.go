package cmd

import (
	"fmt"
	"mini-docker/cgroup/subsystems"
	netcmd "mini-docker/cmd/network"
	"mini-docker/container"
	"mini-docker/runtime"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
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
			imageName, command := args[0], args[1:]
			runtime.Run(ti, command, env, volume, port, cfg, imageName, name, net)
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}

	commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "commit container into image",
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime.CommitContainer(args[0], args[1])
			return nil
		},
		Args: cobra.MinimumNArgs(2),
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
		Use:   "ps",
		Short: "list the container",
		Run: func(cmd *cobra.Command, args []string) {
			container.ListContainer()
		},
	}

	logCmd = &cobra.Command{
		Use:   "log",
		Short: "print logs of the container",
		Run: func(cmd *cobra.Command, args []string) {
			container.GetContainerLog(args[0])
		},
		Args: cobra.MinimumNArgs(1),
	}

	execCmd = &cobra.Command{
		Use:   "exec",
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
			container.ExecContainer(args[0], args[1:])
			return nil
		},
	}

	stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stop the container",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			container.StopContainer(args[0])
		},
	}

	removeCmd = &cobra.Command{
		Use:   "rm",
		Short: "remove the container",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			container.RemoveContainer(args[0])
		},
	}

	networkCmd = &cobra.Command{
		Use:   "network",
		Short: "container network commands",
		Run: func(cmd *cobra.Command, args []string) {},
	}
)

var (
	// generic
	ti     bool
	volume []string
	daemon bool
	name   string
	env    []string
	// network
	net  string
	port []string
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
	runCmd.Flags().StringArrayVarP(&volume, "volume", "v", []string{}, "set the volume of the container")
	runCmd.Flags().BoolVarP(&daemon, "detach", "d", false, "detach container")
	runCmd.Flags().StringVar(&name, "name", "", "set the container name")
	runCmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "set environment")
	runCmd.Flags().StringArrayVarP(&port, "port", "p", []string{}, "port mapping")
	runCmd.Flags().StringVar(&net, "net", "", "set the container network")
	// child command
	networkCmd.AddCommand(netcmd.CreateCmd, netcmd.ListCmd, netcmd.RemoveCmd)
}
