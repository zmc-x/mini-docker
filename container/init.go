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

	if err := setMount(); err != nil {
		zap.L().Error("set mount is error", zap.String("error", err.Error()))
		return fmt.Errorf("container set mount error")
	}

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

// pivot_root
func pivotRoot(root string) error {
	// prevents propagation to other mount namespaces
	if err := syscall.Mount("", "/", "", syscall.MS_SLAVE | syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("prevents propagation error: %v", err)
	}

	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND | syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	if err := syscall.Chdir(root); err != nil {
		return fmt.Errorf("chdir %v error: %v", root, err)
	}

	if err := syscall.PivotRoot(".", "."); err != nil {
		return fmt.Errorf("pivot_root error: %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / error: %v", err)
	}
	return nil
}

func setMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	zap.L().Info("current location was found", zap.String("path", pwd))
	if err := pivotRoot(pwd); err != nil {
		return err
	}

	defaultMountFlag := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlag), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_STRICTATIME | syscall.MS_NOSUID, "mode=755")
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