package commands

import (
	"flag"
	"fmt"
	"strings"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"github.com/opslabjpl/goamz.git/aws"
)

// TODO: make these configurable
var region = aws.USGovWest
var bucketName = "comsolcloud"
var workspaceName = "workspace"

func PushCommand(args []string) {

	if len(args) < 1 {
		fmt.Println("You need to specify a name for the fileset.")
		fmt.Println("Usage: ekit push fileset_name [-c \"some helpful comment\"] [-filters \"pattern1,pattern2,…,patternN\"]")
		return
	}

	filesetName := args[0]

	flagSet := flag.NewFlagSet("ekit push fileset_name", flag.ExitOnError)
	comment := flagSet.String("c", "", "Comment to give to the fileset")
	patternString  := flagSet.String("filters", "", "upload only files from workspace that match given path patterns")
	flagSet.Parse(args[1:])

	ws := workspace.GetWorkspace(".")

	var patterns []string
	if *patternString != "" {
		patterns = strings.Split(*patternString, ",")
	}

	ws.Push(filesetName, *comment, patterns)
}
