package main

import (
	"mini-docker/cmd"
	"mini-docker/logger"
	_ "mini-docker/nsenter"
	_ "mini-docker/config"

	"go.uber.org/zap"
)

func main() {
	logger := logger.CreateLogger()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)
	cmd.Execute()
}
