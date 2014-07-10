package pool

import (
	"github.com/opslabjpl/earthkit-cli/config"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"github.com/opslabjpl/goamz.git/ec2"
	"github.com/opslabjpl/goprovision.git"
	"net"
)

func New(workspace workspace.Workspace) *Pool {
	auth := config.AWSAuth()

	provisioner := goprovision.Provisioner{
		EC2: ec2.New(auth, config.Region),
	}

	return &Pool{Workspace: workspace, Provisioner: provisioner}
}

func etcdAndDockerUp(ipaddr string) bool {
	// etcd
	_, err := net.Dial("tcp", ipaddr+":4001")
	if err != nil {
		return false
	}

	// docker
	// _, err = net.Dial("tcp", ipaddr+":80")
	// if err != nil {
	// 	return false
	// }
	return true
}

func getStatus(instance ec2.Instance) string {
	// ec2 state codes
	//  0 (pending) | 16 (running) | 32 (shutting-down) | 48 (terminated) | 64 (stopping) | 80 (stopped)
	if instance.State.Code == 16 && etcdAndDockerUp(instance.IPAddress) {
		return "available"
	} else if instance.State.Code <= 16 {
		return "pending"
	} else {
		return instance.State.Name
	}
}
