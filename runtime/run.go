package runtime

import (
	"mini-docker/cgroup"
	"mini-docker/cgroup/subsystems"
	"mini-docker/container"
	"os"
	"strings"

	"go.uber.org/zap"
)

func Run(tty bool, args []string, cfg *subsystems.ResourceConfig) {
	parent, writePipe, err := container.NewParentProcess(tty)
	if err != nil {
		zap.L().Error("new parent process error", zap.String("error", err.Error()))
		return
	}
	if err := parent.Start(); err != nil {
		zap.L().Error("parent process don't start", zap.String("error", err.Error()))
		return
	}
	// set resource limit
	cgroupManager := cgroup.NewCgroupManager("mini-docker")
	defer cgroupManager.Destroy()
	cgroupManager.Set(cfg)
	cgroupManager.Apply(parent.Process.Pid)
	err = sendInitCMD(args, writePipe)
	if err != nil {
		zap.L().Error("don't send command to child process", zap.String("error", err.Error()))
	}
	parent.Wait()
	os.Exit(0)
}

// send command to child process(container)
func sendInitCMD(args []string, w *os.File) error {
	cmd := strings.Join(args, " ")
	zap.L().Info("the total command is", zap.String("command", cmd))
	_, err := w.WriteString(cmd)
	if err != nil {
		return err
	}
	defer w.Close()
	return nil
}
