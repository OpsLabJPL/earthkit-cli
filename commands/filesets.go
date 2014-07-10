package commands

import (
	"fmt"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"log"
	"path"
	"strings"
)

func FilesetsCommand(args []string) {
	var wsName string
	var ws *workspace.Workspace

	if len(args) >= 1 {
		wsName = args[0]
		ws = workspace.New(wsName, "")
	} else {
		ws_ := workspace.GetWorkspace(".")
		ws = &ws_
	}

	keys, err := ws.Remote().Filesets()
	if err != nil {
		log.Fatal(err)
	}

	if len(keys) == 0 {
		fmt.Println("No remote filesets found for workspace '" + ws.Name + "'.")
		return
	}

	fmt.Printf("Remote filesets for '%s':\n", ws.Name)
	for _, key := range keys {
		fmt.Printf("  %s\n", strings.Replace(path.Base(key.Key), ".json.gz", "", 1))
	}
}

func FilesetDeleteCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("You need to specify a fileset")
		return
	}
	ws := workspace.GetWorkspace(".")
	filesetName := args[0]
	ws.DeleteFileset(filesetName)
}
