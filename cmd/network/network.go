package network

import (
	"fmt"
	"mini-docker/network"

	"github.com/spf13/cobra"
)

var (
	CreateCmd = &cobra.Command{
		Use: "create",
		Short: "create container network",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			network.Init()
			if err := network.CreateNetwork(args[0], subnet, driver); err != nil {
				return fmt.Errorf("create network error %v", err)
			}
			return nil
		},
	}

	RemoveCmd = &cobra.Command{
		Use: "remove",
		Short: "remove container network",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := network.Init(); err != nil {
				return fmt.Errorf("net work init error %v", err)
			}
			if err := network.RemoveNetwork(args[0]); err != nil {
				return fmt.Errorf("remove network error %v", err)
			}
			return nil
		},
	}

	ListCmd = &cobra.Command{
		Use: "list",
		Short: "list container network",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := network.Init(); err != nil {
				return fmt.Errorf("net work init error %v", err)
			}
			network.ListNetwork()
			return nil
		},
	}
)

var (
	subnet string 
	driver string
)

func init() {
	CreateCmd.Flags().StringVar(&subnet, "subnet", "", "subnet cidr")
	CreateCmd.Flags().StringVar(&driver, "driver", "", "network driver")
}