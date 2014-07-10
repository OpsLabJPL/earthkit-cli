package commands

import (
	"flag"
	"fmt"
	"github.com/opslabjpl/earthkit-cli/config"
	"github.com/opslabjpl/earthkit-cli/localrun"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"log"
	"os"
	"path"
	"strings"
)

//
// Usage: earthkit run <myapp> <optional -w> <container command>
//

func RunCommand(args []string) {
	config.Load()

	workingDir, _ := os.Getwd()
	workspace := workspace.GetWorkspace(workingDir)
	workspaceDir := workspace.LocalRootDir

	fmt.Println(fmt.Sprintf("working dir: %s", workingDir))
	fmt.Println(fmt.Sprintf("workspace dir: %s", workspaceDir))

	if !strings.HasPrefix(workingDir, workspaceDir) {
		log.Fatal(fmt.Sprintf("cannot find EarthKit workspace above current directory"))
	}

	relWorkingDir := workingDir[len(workspaceDir):]

	flagSet := flag.NewFlagSet("localrun", flag.ExitOnError)

	if len(args) < 1 {
		fmt.Println("uUage:", os.Args[0], "run {app-name} [container commands]")
		flagSet.PrintDefaults()
		return
	}

	flagErr := flagSet.Parse(args[1:])
	if flagErr != nil {
		log.Println(flagErr)
	}

	var repo_name = args[0]
	var docker_args []string

	// add all the usual daemon port and interactive terminal flags
	if *config.DOCKER_HTTP_PORT != -1 {
		docker_args = []string{fmt.Sprintf("-H=:%d", *config.DOCKER_HTTP_PORT), "run", "-i", "-t"}
	} else {
		docker_args = []string{"run", "-i", "-t"}
	}
	// setup docker mounts and working directory
	mount_opts := []string{"-v", fmt.Sprintf("%s:/data", workspaceDir), "-w", path.Join("/data", relWorkingDir)}
	docker_args = append(docker_args, mount_opts...)
	// add the repository to run and any additional commands
	docker_args = append(docker_args, repo_name)
	// add the command to execute (default to bash if none given)
	if len(flagSet.Args()) > 0 {
		docker_args = append(docker_args, flagSet.Args()...)
	} else {
		docker_args = append(docker_args, "/bin/bash")
	}

	localrun.DockerHijack(docker_args)
}
