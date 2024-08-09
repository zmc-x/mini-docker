package container

import "time"

// container information
type ContainerMeta struct {
	PID      int       `json:"pid"`
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	CreateAt time.Time `json:"create_at"`
	Command  string    `json:"command"`
	Status   string    `json:"status"`
}

const (
	// container status
	RUNING = "runing"
	STOP   = "stopped"
	EXIT   = "exited"

	// constant
	ConfigName      = "config.json"
	ContainerLog    = "container.log"
	DefaultInfoPath = "/var/run/mini-docker/%s/"

	// environment
	ENV_EXEC_PID = "mini_docker_pid"
	ENV_EXEC_CMD = "mini_docker_cmd"
)
