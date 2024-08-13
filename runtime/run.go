package runtime

import (
	"mini-docker/cgroup"
	"mini-docker/cgroup/subsystems"
	"mini-docker/container"
	"mini-docker/network"
	"os"
	"strings"

	"go.uber.org/zap"
)

func Run(tty bool, args, env, volumePath, port []string, cfg *subsystems.ResourceConfig, imageName, containerName, net string) {
	var containerID string = container.GenerateContainerId()
	if containerName == "" {
		containerName = containerID
	}
	parent, writePipe, err := container.NewParentProcess(tty, imageName, containerName, env, volumePath)
	if err != nil {
		zap.L().Sugar().Errorf("new parent process error %v", err)
		return
	}
	if err := parent.Start(); err != nil {
		zap.L().Sugar().Errorf("parent process don't start. %v", err)
		return
	}
	// record the container information
	containerName, err = container.RecordContainer(parent.Process.Pid, args, volumePath, port, containerName, containerID, imageName)
	if err != nil {
		zap.L().Sugar().Error("record the container information error")
		return
	}
	// set resource limit
	cgroupManager := cgroup.NewCgroupManager("mini-docker")
	defer cgroupManager.Destroy()
	cgroupManager.Set(cfg)
	cgroupManager.Apply(parent.Process.Pid)
	// set network
	if net != "" {
		if err := network.Init(); err != nil {
			zap.L().Sugar().Errorf("init network error %v", err)
			return 
		}
		containerMeta := &container.ContainerMeta{
			PID:  parent.Process.Pid,
			ID:   containerID,
			Name: containerName,
			Port: strings.Join(port, " "),
		}
		if err := network.Connect(net, containerMeta); err != nil {
			zap.L().Sugar().Errorf("container connect network error %v", err)
			return
		}
	}
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
		container.DeleteWorkSpace(imageName, containerName, volumePath)
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
