package subsystems

type ResourceConfig struct {
	MemoryLimit string 
	CpuSet string 
	CpuShare string 
}


type SubSystem interface {
	// return subsystem name
	Name() string 
	// set cgroup to hierarchy
	Set(path string, cfg *ResourceConfig) error 
	// append pid in the cgroup
	Apply(path string, pid int) error 
	// remove cgroup
	Remove(path string) error
}

var SubSystems []SubSystem = []SubSystem{
	&cpuSetSubSystem{},
	&cpuSubSystem{},
	&memorySubSystem{},
}