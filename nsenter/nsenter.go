package nsenter

/*
#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <sched.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>
#define N 1024

__attribute__((constructor)) void enter_namespace(void) {
	char* mini_docker_pid;
	// get value from environment
	mini_docker_pid = getenv("mini_docker_pid");
	if(!mini_docker_pid) {
		return;
	}
	char* mini_docker_cmd;
	mini_docker_cmd = getenv("mini_docker_cmd");
	if(!mini_docker_cmd) {
		return;
	}
	char *namespace[] = {"ipc", "uts", "net", "pid", "mnt"};
	char nspath[N];
	for(int i = 0; i < 5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", mini_docker_pid, namespace[i]);
		int fd = open(nspath, O_RDONLY);
		// entry namespace
		int t = setns(fd, 0);
		close(fd);
	}
	int status = system(mini_docker_cmd);
	exit(0);
	return;
}

*/
import "C"
