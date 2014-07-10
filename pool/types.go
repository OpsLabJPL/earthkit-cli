package pool

import (
	"github.com/opslabjpl/earthkit-cli/workspace"
	"github.com/opslabjpl/goprovision.git"
)

type Pool struct {
	// WorkspaceName string
	Workspace   workspace.Workspace
	Provisioner goprovision.Provisioner
}
