package runtime

import (
	"mini-docker/config"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
)

func CommitContainer(containerName, imageName string) {
	mntUrl := filepath.Join(config.ContainerPath, containerName, "merged")
	imageTar := filepath.Join(config.ImagePath, imageName+".tar.gz")
	if _, err := exec.Command("tar", "czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		zap.L().Sugar().Errorf("commit container to image error %v", err)
	}
}
