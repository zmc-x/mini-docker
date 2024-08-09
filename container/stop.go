package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"go.uber.org/zap"
)

func StopContainer(containerName string) {
	pid, err := getPidByContainerName(containerName)
	if err != nil {
		zap.L().Sugar().Errorf("get pid by container name error %v", err)
		return
	}

	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		zap.L().Sugar().Errorf("kill the pid %d error %v", pid, err)
		return
	}

	// override the config.json
	meta, err := getContaineryContainerName(containerName)
	if err != nil {
		zap.L().Sugar().Errorf("get container meta by container name error %v", err)
		return
	}
	meta.Status = STOP
	meta.PID = -1
	cfg, err := json.Marshal(meta)
	if err != nil {
		zap.L().Sugar().Errorf("marshal container information error %v", err)
		return
	}
	dirPath := fmt.Sprintf(DefaultInfoPath, containerName)
	cfgPath := filepath.Join(dirPath, ConfigName)
	err = os.WriteFile(cfgPath, cfg, 0622)
	if err != nil {
		zap.L().Sugar().Errorf("write file error %v", err)
		return
	}
}
