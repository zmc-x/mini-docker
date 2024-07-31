package container

import (
	"os"
	"os/exec"
	"syscall"
)

const PrgPath = "/proc/self/exe" 

// parent process
func NewParentProcess(tty bool) (*exec.Cmd, *os.File, error) {
	r, w, err := createPipe()
	if err != nil {
		return nil, nil, err
	}
	cmd := exec.Command(PrgPath, "init")
	// set namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWPID |
					syscall.CLONE_NEWNS,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout 
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{r}
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

