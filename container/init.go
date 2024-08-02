package container

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

func ContainerInit() error {
	receiveCMD, err := readCMD()
	if err != nil {
		zap.L().Error("read command from pipe error", zap.String("error", err.Error()))
		return fmt.Errorf("read command from pipe error %v", err)
	}
	if len(receiveCMD) == 0 {
		zap.L().Error("run container get user command error")
		return fmt.Errorf("run container get user command error")
	}

	defaultMountFlag := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlag), "")
	path, err := exec.LookPath(receiveCMD[0])
	if err != nil {
		zap.L().Error("exec look path error", zap.String("error", err.Error()))
		return err 
	}
	zap.L().Info("find path", zap.String("path", path))
	// override
	if err := syscall.Exec(path, receiveCMD, os.Environ()); err != nil {
		zap.L().Error(err.Error())
		return err
	}
	return nil
}

func readCMD() ([]string, error) {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := io.ReadAll(pipe)
	if err != nil {
		return nil, err 
	}
	return strings.Split(string(msg), " "), nil
}