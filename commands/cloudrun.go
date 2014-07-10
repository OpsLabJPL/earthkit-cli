package commands

import (
	"flag"
	"fmt"
	"github.com/opslabjpl/earthkit-cli/cloudrun"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"log"
	"os"
	"strings"
)

func CloudRunCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("You need to specify the app and command you want to run.")
		fmt.Println("Usage: ", os.Args[0], "cloudrun [-fileset fileset] app-name [container cmd]")
		return
	}

	flagSet := flag.NewFlagSet("ekit cloudrun [-fileset fileset] app cmd", flag.ExitOnError)

	filesetFlag := flagSet.String("fileset", "", "fileset to use")
	flagSet.Parse(args)
	leftoverArgs := args[len(args)-flagSet.NArg():]

	workingDir, _ := os.Getwd()
	ws := workspace.GetWorkspace(".")
	workspaceDir := ws.LocalRootDir
	if !strings.HasPrefix(workingDir, workspaceDir) {
		log.Fatal(fmt.Sprintf("cannot find EarthKit workspace above current directory"))
	}
	relWorkingDir := workingDir[len(workspaceDir):]

	var fileset string
	if *filesetFlag == "" {
		fileset = ws.GetCurrentFileSetName()
	} else {
		fileset = *filesetFlag
	}

	app := leftoverArgs[0]
	cmd := leftoverArgs[1:]

	println("Using fileset", fileset)

	// Example of how we would submit a job to the cloud
	cloudrun.Run(app, cmd, fileset, []string{}, relWorkingDir)
}

func CloudRunStatusCommand(args []string) {
	cloudRun := cloudrun.New()
	if len(args) == 0 {
		cloudRun.GetJobStatuses()
	} else if len(args) == 1 {
		cloudRun.GetJobStatus(args[0])
	}
}
