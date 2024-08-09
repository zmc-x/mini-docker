package container

import "go.uber.org/zap"

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
}