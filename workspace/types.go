package workspace

import (
	"github.com/opslabjpl/earthkit-cli/workspace/remote"
	"os"
)

type Workspace struct {
	Name          string `json:"workspace"`
	LocalRootDir  string
	remote_       *remote.Remote
	patternCache_ []string
}

type filesDateSort []os.FileInfo
