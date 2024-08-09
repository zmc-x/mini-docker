package container

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func ExecContainer(containerName string, args []string) {
	pid, err := getPidByContainerName(containerName)
	if err != nil {
		zap.L().Sugar().Error("get pid error %v", err)
		return
	}
	command := strings.Join(args, " ")
	zap.L().Sugar().Infof("container pid %d", pid)
	zap.L().Sugar().Infof("command %s", command)

	cmd := exec.Command(prgPath, "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// set environment
	os.Setenv(ENV_EXEC_PID, fmt.Sprint(pid))
	os.Setenv(ENV_EXEC_CMD, command)
	if err := cmd.Run(); err != nil {
		zap.L().Sugar().Errorf("exec container %s error %s", containerName, err)
	}
}

func getPidByContainerName(containerName string) (int, error) {
	dirPath := fmt.Sprintf(DefaultInfoPath, containerName)
	cfgPath := filepath.Join(dirPath, ConfigName)
	cfg, err := os.ReadFile(cfgPath)
	if err != nil {
		zap.L().Sugar().Errorf("read the config file error %v", err)
		return -1, err
	}
	meta := new(ContainerMeta)
	if err := json.Unmarshal(cfg, meta); err != nil {
		zap.L().Sugar().Errorf("unmarshal config file error %v", err)
		return -1, err
	}
	return meta.PID, nil
}


func getContaineryContainerName(containerName string) (*ContainerMeta, error) {
	dirPath := fmt.Sprintf(DefaultInfoPath, containerName)
	cfgPath := filepath.Join(dirPath, ConfigName)
	cfg, err := os.ReadFile(cfgPath)
	if err != nil {
		zap.L().Sugar().Errorf("read the config file error %v", err)
		return nil, err
	}
	meta := new(ContainerMeta)
	if err := json.Unmarshal(cfg, meta); err != nil {
		zap.L().Sugar().Errorf("unmarshal config file error %v", err)
		return nil, err
	}
	return meta, nil
}