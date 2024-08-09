package runtime

import (
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
)

func CommitContainer(imageName string) {
	mntUrl := "/home/hellozmc/busybox"
	imageTar := filepath.Join("/home/hellozmc/download", imageName+".tar.gz")
	if _, err := exec.Command("tar", "czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		zap.L().Sugar().Errorf("commit container to image error %v", err)
	}
}
