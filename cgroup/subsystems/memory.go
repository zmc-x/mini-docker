package subsystems

import (
	"fmt"
	"os"
	"path/filepath"
)

type memorySubSystem struct{}

func (m *memorySubSystem) Name() string {
	return "memory"
}

func (m *memorySubSystem) Set(cgroupPath string, cfg *ResourceConfig) error {
	if subsysCgroupPath, err := getCgroupPath(m.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		if cfg.MemoryLimit != "" {
			if err = os.WriteFile(filepath.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(cfg.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("set cgroup memory error")
}

func (m *memorySubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := getCgroupPath(m.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		if err = os.WriteFile(filepath.Join(subsysCgroupPath, "tasks"), []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
	}
	return nil
}

func (m *memorySubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := getCgroupPath(m.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		return os.RemoveAll(filepath.Join(subsysCgroupPath, cgroupPath))
	}
}
