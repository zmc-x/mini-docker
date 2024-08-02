package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func findCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	prefix := filepath.Join("/cgroup", subsystem)
	for scan.Scan() {
		txt := scan.Text()
		if !strings.Contains(txt, prefix) {continue}
		return strings.Split(txt, " ")[4]
	}
	if err := scan.Err(); err != nil {
		return ""
	}
	return ""
}

func getCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := findCgroupMountPoint(subsystem)
	cgroupPath = filepath.Join(cgroupRoot, cgroupPath)
	if _, err := os.Stat(cgroupPath); err == nil || autoCreate && os.IsNotExist(err) {
		if os.IsNotExist(err) {
			err = os.Mkdir(cgroupPath, 0755)
			if err != nil {
				return "", fmt.Errorf("create cgroup error %v", err)
			}
		}
		return cgroupPath, nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}