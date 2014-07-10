package commands

import (
	"flag"
	"fmt"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"os"
	"strings"
)

func CloneCommand(args []string) {

	flagSet := flag.NewFlagSet("ekit clone workspace_name [fileset_name]", flag.ExitOnError)
	patternString := flagSet.String("filters", "", "only pull down files matched against a set of path patterns")

	if len(args) < 1 {
		fmt.Println("You need to specify the workspace you want to clone.")
		fmt.Println("Usage:", os.Args[0], "clone workspace_name [fileset_name] [-addonly pattern1,pattern2,…,patternN]")
		return
	}

	wsName := args[0]
	ws := workspace.Workspace{
		Name: wsName,
	}

	var fileSet string
	var err error

	if len(args) >= 2 && args[1] != "-addonly" {
		fileSet = args[1]
		flagSet.Parse(args[2:])
	} else {
		flagSet.Parse(args[1:])
		fileSet, err = ws.Remote().LatestFileset()
		if err != nil {
			fmt.Printf("Unable to determine latest fileset: %s\n", err)
			return
		}
	}

	var patterns []string
	if *patternString != "" {
		patterns = strings.Split(*patternString, ",")
	}

	println("Cloning workspace", wsName, "fileset", fileSet)
	println("patternString", *patternString)

	os.Mkdir(wsName, 0700)
	ws.Init(wsName, true)
	ws.Pull(fileSet, patterns)
}
