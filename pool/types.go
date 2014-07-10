package pool

import (
	"github.com/opslabjpl/earthkit-cli/workspace"
	"github.com/opslabjpl/goprovision"
)

type Pool struct {
	// WorkspaceName string
	Workspace   workspace.Workspace
	Provisioner goprovision.Provisioner
}
