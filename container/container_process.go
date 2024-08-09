package container

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

const prgPath = "/proc/self/exe"

var ErrCreateWorkSpace = errors.New("create overlayfs work space error")

// parent process
func NewParentProcess(tty bool, volumePath, containerName string) (*exec.Cmd, *os.File, error) {
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
	cmd.ExtraFiles = []*os.File{r}
	rootURL, mntURL := "/home/hellozmc/download", "/home/hellozmc/busybox"
	err = NewWorkSpace(rootURL, mntURL, volumePath)
	if err != nil {
		return nil, nil, err
	}
	cmd.Dir = mntURL
	return cmd, w, nil
}

// overlayfs
// lowerdir + upperdir + workdir + mergedir
func NewWorkSpace(rootURL, mntURL, volumeURL string) error {
	if err := createOverlayfsLower(rootURL); err != nil {
		zap.L().Error("create overlayfs lower error", zap.String("error", err.Error()))
		return ErrCreateWorkSpace
	}
	zap.L().Info("create overlayfs lower dir successful")
	if err := createOverlayfsDirs(rootURL); err != nil {
		zap.L().Error("create overlayfs uppper or work error", zap.String("error", err.Error()))
		deleteDirs(rootURL)
		return ErrCreateWorkSpace
	}
	zap.L().Info("create overlayfs upper and work dirs successful")
	if err := mountOverlayfs(rootURL, mntURL); err != nil {
		zap.L().Error("mount overlayfs error", zap.String("error", err.Error()))
		return ErrCreateWorkSpace
	}
	zap.L().Info("mount overlayfs successful")
	// mount volume
	if volumeURL != "" {
		mappingVolumePath := parseVolumeUrl(volumeURL)
		if len(mappingVolumePath) == 2 && mappingVolumePath[0] != "" && mappingVolumePath[1] != "" {
			if err := mountVolume(mntURL, mappingVolumePath); err != nil {
				zap.L().Error("mount volume error", zap.String("error", err.Error()))
				return ErrCreateWorkSpace
			}
		} else {
			zap.L().Warn("input volume path don't correct")
			return ErrCreateWorkSpace
		}
		zap.L().Info("mount volume successful")
	}
	return nil
}

// create overlayfs lower
func createOverlayfsLower(rootURL string) error {
	busyboxURL := filepath.Join(rootURL, "busybox")
	busyboxTarURL := filepath.Join(rootURL, "busybox.tar")
	exist, err := pathExists(busyboxURL)
	if err != nil {
		zap.L().Warn("faild to check whether dir exists", zap.String("dir", busyboxURL), zap.String("error", err.Error()))
	}
	if !exist {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			return fmt.Errorf("mkdir %s failed, error is %v", busyboxURL, err)
		}
	}
	if _, err := exec.Command("tar", "xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
		return fmt.Errorf("tar busybox.tar error, error is %v", err)
	}
	return nil
}

// create overlayfs upper and work
func createOverlayfsDirs(rootURL string) error {
	upper, work := filepath.Join(rootURL, "upper"), filepath.Join(rootURL, "work")
	if err := os.Mkdir(upper, 0777); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("mkdir %s failed, error is %v", upper, err)
	}
	if err := os.Mkdir(work, 0777); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("mkdir %s failed, error is %v", upper, err)
	}
	return nil
}

// mount to rootMnt
func mountOverlayfs(rootURL, rootMnt string) error {
	if err := os.Mkdir(rootMnt, 0777); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("mkdir %s failed, error is %v", rootMnt, err)
	}

	dirs := "lowerdir=" + filepath.Join(rootURL, "busybox") + ",upperdir=" + filepath.Join(rootURL, "upper") +
		",workdir=" + filepath.Join(rootURL, "work")
	// mount
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, rootMnt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mount overlayfs error, error is %v", err)
	}
	return nil
}

func DeleteWorkSpace(rootURL, mntURL, volumePath string) {
	if volumePath != "" {
		mappingVolumePath := parseVolumeUrl(volumePath)
		if len(mappingVolumePath) == 2 && mappingVolumePath[0] != "" && mappingVolumePath[1] != "" {
			if err := umountVolume(mntURL, mappingVolumePath[1]); err != nil {
				zap.L().Error("umount volume error", zap.String("error", err.Error()))
			}
		}
	}
	if err := umountOverfs(mntURL); err != nil {
		zap.L().Error("umount overlayfs error", zap.String("error", err.Error()))
	}

	if err := deleteDirs(rootURL); err != nil {
		zap.L().Error("delete overlayfs upper or work error", zap.String("error", err.Error()))
	}
}

// umount volume
func umountVolume(rootMnt, volumePath string) error {
	cmd := exec.Command("umount", filepath.Join(rootMnt, volumePath))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// umount overlayfs
func umountOverfs(rootMnt string) error {
	// umount
	cmd := exec.Command("umount", rootMnt)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	// delete mount point
	if err := os.RemoveAll(rootMnt); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("remove dir %s is error, error is %v", rootMnt, err)
	}
	return nil
}

// delete upper and work
func deleteDirs(rootURL string) error {
	upper, work := filepath.Join(rootURL, "upper"), filepath.Join(rootURL, "work")
	// upper
	if err := os.RemoveAll(upper); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("remove dir %s is error, error is %v", upper, err)
	}
	// work
	if err := os.RemoveAll(work); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("remove dir %s is error, error is %v", work, err)
	}
	return nil
}

// mount volume
func mountVolume(mntURL string, mappingVolumePath []string) error {
	hostPath, containerPath := mappingVolumePath[0], mappingVolumePath[1]
	if err := os.Mkdir(hostPath, 0777); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("mkdir dir %s error, error is %v", hostPath, err)
	}

	containerPath = filepath.Join(mntURL, containerPath)
	if err := os.Mkdir(containerPath, 0777); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("mkdir dir %s error, error is %v", containerPath, err)
	}

	// mount
	cmd := exec.Command("mount", "--bind", hostPath, containerPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// parse volumeurl
func parseVolumeUrl(path string) []string {
	mappingVolumePath := strings.Split(path, ":")
	return mappingVolumePath
}

// anonymous pipe
func createPipe() (*os.File, *os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return r, w, err
}

func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
