package localrun

import (
	"os"
	"os/exec"
	"log"
	// "fmt"
	"github.com/opslabjpl/earthkit-cli/config"
)

// executes a command through the Docker command-line client. it blocks until
// the command finishes.
func DockerHijack(args []string) {
	docker_cmd := exec.Command(*config.DOCKER_PATH, args...)
	docker_cmd.Stdin = os.Stdin
	docker_cmd.Stdout = os.Stdout
	docker_cmd.Stderr = os.Stderr
	err := docker_cmd.Run()
	if err != nil {
		log.Print(err)
	}
}

// checks if a workspace is valid by looking for a .earthkit/earthkitrc
// file within. it's safe to handle fpath as if it is a directory because
// <fpath>/.earthkit/earthkitrc will still cause a file open error if fpath
// is a file.
func IsWorkspace(fpath string) bool {
	rcPath := fpath + "/.earthkit/earthkitrc"
	f, err := os.Open(rcPath)
	if err != nil {
		return false
	}
	defer f.Close()
	return true
}
