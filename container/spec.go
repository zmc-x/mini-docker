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
	Volume   string    `json:"volume,omitempty"`
	Image    string    `json:"image"`
	Port     string    `json:"port,omitempty"`
}

const (
	// container status
	RUNING = "runing"
	STOP   = "stopped"
	EXIT   = "exited"

	// constant
	ConfigName      = "config.json"
	ContainerLog    = "container.log"
	DefaultInfoPath = "/var/run/mini-docker/container/%s/"

	// environment
	ENV_EXEC_PID = "mini_docker_pid"
	ENV_EXEC_CMD = "mini_docker_cmd"
)
