package config

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

var (
	// mini-docker image path
	ImagePath string = "/mini-docker/images"
	// mini-docker container path
	ContainerPath string = "/mini-docker/containers"
)

func init() {
	home := os.Getenv("HOME")
	ImagePath, ContainerPath = filepath.Join(home, ImagePath), filepath.Join(home, ContainerPath)
	if err := os.MkdirAll(ImagePath, 0622); err != nil {
		zap.L().Sugar().Errorf("mkdir %s error %v", ImagePath, err)
	}
	if err := os.MkdirAll(ContainerPath, 0622); err != nil {
		zap.L().Sugar().Errorf("mkdir %s error %v", ContainerPath, err)
	}
}
