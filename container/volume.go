package container

import (
	"fmt"
	"mini-docker/config"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// overlayfs
// lowerdir + upperdir + workdir + mergedir
func NewWorkSpace(imageName, containerName string, volumePath []string) error {
	if err := createOverlayfsLower(imageName); err != nil {
		zap.L().Sugar().Errorf("create overlayfs lower error %v", err)
		return ErrCreateWorkSpace
	}
	zap.L().Info("create overlayfs lower dir successful")
	if err := createOverlayfsDirs(containerName); err != nil {
		zap.L().Sugar().Errorf("create overlayfs uppper or work error %v", err)
		return ErrCreateWorkSpace
	}
	zap.L().Sugar().Info("create overlayfs upper and work dirs successful")
	if err := mountOverlayfs(imageName, containerName); err != nil {
		zap.L().Sugar().Errorf("mount overlayfs error %v", err)
		return ErrCreateWorkSpace
	}
	zap.L().Sugar().Info("mount overlayfs successful")
	// mount volume
	for _, volumeUrl := range volumePath {
		if volumeUrl != "" {
			mappingVolumePath := parseVolumeUrl(volumeUrl)
			if len(mappingVolumePath) == 2 && mappingVolumePath[0] != "" && mappingVolumePath[1] != "" {
				if err := mountVolume(containerName, mappingVolumePath); err != nil {
					DeleteWorkSpace(imageName, containerName, volumePath)
					zap.L().Sugar().Errorf("mount volume error %v", err)
					return ErrCreateWorkSpace
				}
			} else {
				DeleteWorkSpace(imageName, containerName, volumePath)
				zap.L().Sugar().Warn("input volume path don't correct")
				return ErrCreateWorkSpace
			}
			zap.L().Sugar().Info("mount volume successful")
		}
	}
	return nil
}

// create overlayfs lower
func createOverlayfsLower(imageName string) error {
	imageURL := filepath.Join(config.ImagePath, imageName)
	imageTarURL := filepath.Join(config.ImagePath, imageName+".tar")
	exist, err := fileExists(imageURL)
	if err != nil {
		zap.L().Sugar().Warnf("faild to check whether dir %s exists. %v", imageURL, err)
	}
	if !exist {
		if err := os.Mkdir(imageURL, 0777); err != nil {
			return fmt.Errorf("mkdir %s failed, error is %v", imageURL, err)
		}
	}
	exist, err = fileExists(imageTarURL)
	if err != nil || !exist {
		imageTarURL = filepath.Join(config.ImagePath, imageName+".tar.gz")
	}
	if _, err := exec.Command("tar", "xvf", imageTarURL, "-C", imageURL).CombinedOutput(); err != nil {
		return fmt.Errorf("tar %s error, error is %v", imageTarURL, err)
	}
	return nil
}

// create overlayfs upper and work
func createOverlayfsDirs(containerName string) error {
	containerDir := filepath.Join(config.ContainerPath, containerName)
	if err := os.Mkdir(containerDir, 0777); err != nil {
		return fmt.Errorf("mkdir %s failed, error is %v", containerDir, err)
	}
	diff, work := filepath.Join(containerDir, "diff"), filepath.Join(containerDir, "work")
	if err := os.Mkdir(diff, 0777); err != nil {
		return fmt.Errorf("mkdir %s failed, error is %v", diff, err)
	}
	if err := os.Mkdir(work, 0777); err != nil {
		return fmt.Errorf("mkdir %s failed, error is %v", work, err)
	}
	return nil
}

// mount overlayfs
func mountOverlayfs(imageName, containerName string) error {
	imageUrl := filepath.Join(config.ImagePath, imageName)
	containerUrl := filepath.Join(config.ContainerPath, containerName)

	mnt := filepath.Join(containerUrl, "merged")
	upper := filepath.Join(containerUrl, "diff")
	work := filepath.Join(containerUrl, "work")

	if err := os.Mkdir(mnt, 0777); err != nil {
		return fmt.Errorf("mkdir %s failed, error is %v", mnt, err)
	}

	dirs := "lowerdir=" + imageUrl + ",upperdir=" + upper + ",workdir=" + work
	// mount
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mnt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mount overlayfs error, error is %v", err)
	}
	return nil
}

func DeleteWorkSpace(imageName, containerName string, volumePath []string) {
	umountVolume(containerName, volumePath)

	if err := umountOverfs(containerName); err != nil {
		zap.L().Sugar().Errorf("umount overlayfs error %v", err)
	}

	if err := deleteDirs(imageName, containerName); err != nil {
		zap.L().Sugar().Errorf("delete overlayfs upper or work error %v", err)
	}
}

// umount volume
func umountVolume(containerName string, volumePath []string) {
	for _, volumeUrl := range volumePath {
		if volumeUrl != "" {
			mappingVolumePath := parseVolumeUrl(volumeUrl)
			if len(mappingVolumePath) == 2 && mappingVolumePath[0] != "" && mappingVolumePath[1] != "" {
					cmd := exec.Command("umount", filepath.Join(config.ContainerPath, containerName, "merged", mappingVolumePath[1]))
					cmd.Stderr = os.Stderr
					cmd.Stdout = os.Stdout
					if err := cmd.Run(); err != nil {
						zap.L().Sugar().Errorf("umount volume error %v", err)
					}
			}
		}
	}
}

// umount overlayfs
func umountOverfs(containerName string) error {
	// umount
	mnt := filepath.Join(config.ContainerPath, containerName, "merged")
	cmd := exec.Command("umount", mnt)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	// delete mount point
	if err := os.RemoveAll(mnt); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove dir %s is error, error is %v", mnt, err)
	}
	return nil
}

// delete upper and work
func deleteDirs(imageName, containerName string) error {
	containerUrl := filepath.Join(config.ContainerPath, containerName)
	imageUrl := filepath.Join(config.ImagePath, imageName)
	// container(upper and work)
	if err := os.RemoveAll(containerUrl); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove dir %s is error, error is %v", containerUrl, err)
	}
	// low
	if err := os.RemoveAll(imageUrl); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove dir %s is error, error is %v", imageUrl, err)
	}
	return nil
}

// mount volume
func mountVolume(containerName string, mappingVolumePath []string) error {
	hostPath, containerPath := mappingVolumePath[0], mappingVolumePath[1]
	if err := os.Mkdir(hostPath, 0777); err != nil && !os.IsExist(err) {
		return fmt.Errorf("mkdir dir %s error, error is %v", hostPath, err)
	}

	mnt := filepath.Join(config.ContainerPath, containerName, "merged")
	containerPath = filepath.Join(mnt, containerPath)
	if err := os.Mkdir(containerPath, 0777); err != nil && !os.IsExist(err) {
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

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
