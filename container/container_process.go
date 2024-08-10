package container

import (
	"errors"
	"fmt"
	"mini-docker/config"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const prgPath = "/proc/self/exe"

var ErrCreateWorkSpace = errors.New("create overlayfs work space error")

// parent process
func NewParentProcess(tty bool, imageName, containerName string, env, volumePath []string) (*exec.Cmd, *os.File, error) {
	r, w, err := createPipe()
	if err != nil {
		return nil, nil, err
	}
	cmd := exec.Command(prgPath, "init")
	// set namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | 
		syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dirPath := fmt.Sprintf(DefaultInfoPath, containerName)
		if err := os.MkdirAll(dirPath, 0644); err != nil {
			return nil, nil, err
		}
		logPath := filepath.Join(dirPath, ContainerLog)
		f, err := os.Create(logPath)
		if err != nil {
			return nil, nil, err
		}
		cmd.Stdout = f
	}

	err = NewWorkSpace(imageName, containerName, volumePath)
	if err != nil {
		return nil, nil, err
	}
	cmd.ExtraFiles = []*os.File{r}
	cmd.Dir = filepath.Join(config.ContainerPath, containerName, "merged")
	cmd.Env = append(os.Environ(), env...)
	return cmd, w, nil
}


// anonymous pipe
func createPipe() (*os.File, *os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return r, w, err
}
