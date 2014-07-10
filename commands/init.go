package commands

import (
	"fmt"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"os"
)

func InitCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("You need to specify a name for the workspace.")
		fmt.Println("Usage:", os.Args[0], "init workspace_name [dir]")
		return
	}

	wsName := args[0]
	dir := "."

	if len(args) >= 2 {
		dir = args[1]
	}

	ws := workspace.Workspace{
		Name: wsName,
	}
	ws.Init(dir, false)
}
