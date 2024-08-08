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
		zap.L().Error("read the container error", zap.String("container name", containerName), zap.String("error", err.Error()))
		return
	}
	fmt.Print(string(content))
}