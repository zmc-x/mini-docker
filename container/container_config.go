package container

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// record the container information
func RecordContainer(pid int, args, volume []string, containerName, containerID, imageName string) (string, error) {
	createAt := time.Now()
	command := strings.Join(args, " ")

	containerMeta := &ContainerMeta{
		ID:       containerID,
		PID:      pid,
		Command:  command,
		Name:     containerName,
		Status:   RUNING,
		CreateAt: createAt,
		Volume:   strings.Join(volume, " "),
		Image:    imageName,
	}

	// write to config
	cfg, err := json.Marshal(containerMeta)
	if err != nil {
		zap.L().Sugar().Errorf("marshal container information error %v", err)
		return "", err
	}

	dirPath := fmt.Sprintf(DefaultInfoPath, containerName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		zap.L().Sugar().Errorf("mkdir dir %s error %v", dirPath, err)
		return "", err
	}

	cfgPath := filepath.Join(dirPath, ConfigName)
	f, err := os.Create(cfgPath)
	if err != nil {
		zap.L().Sugar().Errorf("create config file error %v", err)
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(cfg); err != nil {
		zap.L().Sugar().Errorf("write config file error %v", err)
		return "", err
	}
	return containerName, nil
}

func DeleteConfig(containerName string) error {
	dirPath := fmt.Sprintf(DefaultInfoPath, containerName)
	return os.RemoveAll(dirPath)
}

// generate container id
// container-id = sha256(uuid)
func GenerateContainerId() string {
	hash := sha256.New()
	hash.Write([]byte(uuid.NewString()))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
