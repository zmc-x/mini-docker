package container

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

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

func ListContainer() {
	dirPath := fmt.Sprintf(DefaultInfoPath, "")
	dirPath = dirPath[:len(dirPath)-1]

	// ls dirPath
	files, err := os.ReadDir(dirPath)
	if err != nil {
		zap.L().Sugar().Errorf("read the directory error %v", err)
		return
	}
	containers := []*ContainerMeta{}
	for _, file := range files {
		meta, err := getContainerInfo(file)
		if meta == nil && err == nil {
			continue
		}
		if err != nil {
			zap.L().Sugar().Error("get container information error")
			continue
		}
		containers = append(containers, meta)
	}
	
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			item.ID,
			item.Name,
			item.PID,
			item.Status,
			item.Command,
			item.CreateAt.Format(time.DateTime),
		)
	}
	if err := w.Flush(); err != nil {
		zap.L().Sugar().Errorf("flush error %v", err)
		return
	}
}

func getContainerInfo(file fs.DirEntry) (*ContainerMeta, error) {
	if file.IsDir() {
		cfgPath := fmt.Sprintf(DefaultInfoPath, file.Name())
		f, err := os.ReadFile(filepath.Join(cfgPath, ConfigName))
		if err != nil {
			zap.L().Sugar().Errorf("read the container config file error %v", err)
			return nil, err
		}
		// decode
		containerMeta := new(ContainerMeta)
		err = json.Unmarshal(f, containerMeta)
		if err != nil {
			zap.L().Sugar().Error("unmarshal json error %v", err)
			return nil, err
		}
		return containerMeta, nil
	}
	return nil, nil
}

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

func RemoveContainer(containerName string) {
	meta, err := getContaineryContainerName(containerName)
	if err != nil {
		zap.L().Sugar().Errorf("get container meta by container name error %v", err)
		return
	}
	
	if meta.Status != STOP {
		zap.L().Sugar().Errorf("the container %s status isn't stopped", containerName)
		return
	}
	err = DeleteConfig(containerName)
	if err != nil {
		zap.L().Sugar().Errorf("delete container config error %v", err)
		return
	}
	DeleteWorkSpace(meta.Image, containerName, meta.Volume)
}

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