package container

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

func GetContainerLog(containerName string) {
	dirPath := fmt.Sprintf(DefaultInfoPath, containerName)
	logPath := filepath.Join(dirPath, ContainerLog)
	content, err := os.ReadFile(logPath)
	if err != nil {
		zap.L().Sugar().Errorf("read the container %s error %v", containerName, err)
		return
	}
	fmt.Print(string(content))
}