package main

import (
	"flag"
	"fmt"
	"github.com/opslabjpl/earthkit-cli/commands"
	"github.com/opslabjpl/earthkit-cli/config"
	"os"
	"runtime"
	"strings"
)

var entryPoints = map[string]func([]string){
	"init":            commands.InitCommand,
	"push":            commands.PushCommand,
	"pull":            commands.PullCommand,
	"cloudrun":        commands.CloudRunCommand,
	"cloudrun-status": commands.CloudRunStatusCommand,
	"run":             commands.RunCommand,
	"fileset-delete":  commands.FilesetDeleteCommand,
	"filesets":        commands.FilesetsCommand,
	"clone":           commands.CloneCommand,
	"workspace":       commands.WorkspaceCommand,
	"pool-create":     commands.PoolCreateCommand,
	"pool-list":       commands.PoolListCommand,
	"pool-kill":       commands.PoolKillCommand,
	"pool-stop":       commands.PoolStopCommand,
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config.Load()

	cmdOpts := make([]string, len(entryPoints))
	i := 0
	for k, _ := range entryPoints {
		cmdOpts[i] = k
		i++
	}

	if (flag.NArg() < 1) || (flag.Arg(0) == "help") {
		fmt.Println("Usage: ", os.Args[0], "-[options] cmd")
		fmt.Println("Commands: ", strings.Join(cmdOpts, "|"))
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	var subCommand = flag.Arg(0)
	var subCommandArgs = flag.Args()[1:]

	cmd, ok := entryPoints[subCommand]
	if !ok {
		fmt.Println("Usage: ", os.Args[0], strings.Join(cmdOpts, "|"))
		flag.PrintDefaults()
		return
	}
	cmd(subCommandArgs)
}
