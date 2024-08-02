package cgroup

import (
	"mini-docker/cgroup/subsystems"

	"go.uber.org/zap"
)

type CgroupManager struct {
	Path string 
	Resource subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	for _, sub := range subsystems.SubSystems {
		sub.Apply(c.Path, pid)
	}
	return nil
}

func (c *CgroupManager) Set(cfg *subsystems.ResourceConfig) error {
	for _, sub := range subsystems.SubSystems {
		sub.Set(c.Path, cfg)
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	for _, sub := range subsystems.SubSystems {
		if err := sub.Remove(c.Path); err != nil {
			zap.L().Warn("remove cgroup failed", zap.String("error", err.Error()))
		}
	}
	return nil
}