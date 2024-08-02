package subsystems

import (
	"fmt"
	"os"
	"path/filepath"
)

type cpuSetSubSystem struct{}

func (c *cpuSetSubSystem) Name() string {
	return "cpuset"
}

func (c *cpuSetSubSystem) Set(cgroupPath string, cfg *ResourceConfig) error {
	if subsysCgroupPath, err := getCgroupPath(c.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		if cfg.CpuShare != "" {
			if err = os.WriteFile(filepath.Join(subsysCgroupPath, "cpuset.cpus"), []byte(cfg.CpuShare), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset fail %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("set cgroup cpuset error")
}

func (c *cpuSetSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := getCgroupPath(c.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		if err = os.WriteFile(filepath.Join(subsysCgroupPath, "tasks"), []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
	}
	return nil
}

func (c *cpuSetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := getCgroupPath(c.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		return os.RemoveAll(filepath.Join(subsysCgroupPath, cgroupPath))
	}
}
