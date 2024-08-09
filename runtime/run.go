package runtime

import (
	"mini-docker/cgroup"
	"mini-docker/cgroup/subsystems"
	"mini-docker/container"
	"os"
	"strings"

	"go.uber.org/zap"
)

func Run(tty bool, volumePath string, args []string, cfg *subsystems.ResourceConfig, containerName string) {
	var containerID string = container.GenerateContainerId()
	if containerName == "" {
		containerName = containerID
	}
	parent, writePipe, err := container.NewParentProcess(tty, volumePath, containerName)
	if err != nil {
		zap.L().Sugar().Errorf("new parent process error %v", err)
		return
	}
	if err := parent.Start(); err != nil {
		zap.L().Sugar().Errorf("parent process don't start. %v", err)
		return
	}
	// record the container information
	containerName, err = container.RecordContainer(parent.Process.Pid, args, containerName, containerID)
	if err != nil {
		zap.L().Sugar().Error("record the container information error")
		return
	}
	// set resource limit
	cgroupManager := cgroup.NewCgroupManager("mini-docker")
	defer cgroupManager.Destroy()
	cgroupManager.Set(cfg)
	cgroupManager.Apply(parent.Process.Pid)
	err = sendInitCMD(args, writePipe)
	if err != nil {
		zap.L().Sugar().Errorf("don't send command to child process. %v", err)
	}
	if tty {
		parent.Wait()
		if err := container.DeleteConfig(containerName); err != nil {
			zap.L().Sugar().Warnf("delete container config failed %v", err)
		}
		// delete overlayf
		rootURL, mntURL := "/home/hellozmc/download", "/home/hellozmc/busybox"
		container.DeleteWorkSpace(rootURL, mntURL, volumePath)
	}
}

// send command to child process(container)
func sendInitCMD(args []string, w *os.File) error {
	cmd := strings.Join(args, " ")
	zap.L().Sugar().Infof("the total command is %s", cmd)
	_, err := w.WriteString(cmd)
	if err != nil {
		return err
	}
	defer w.Close()
	return nil
}
