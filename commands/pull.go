package commands

import (
	"flag"
	"fmt"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"strings"
)

func PullCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("You need to specify a name for the fileset.")
		fmt.Println("Usage: ekit pull fileset_name [-filters \"pattern1,pattern2,…,patternN\"]")
		return
	}

	filesetName := args[0]

	flagSet := flag.NewFlagSet("ekit pull fileset_name", flag.ExitOnError)
	patternString := flagSet.String("filters", "", "only pull down files matched against a set of path patterns")
	flagSet.Parse(args[1:])

	var patterns []string
	if *patternString != "" {
		patterns = strings.Split(*patternString, ",")
	}

	ws := workspace.GetWorkspace(".")
	ws.Pull(filesetName, patterns)
}
