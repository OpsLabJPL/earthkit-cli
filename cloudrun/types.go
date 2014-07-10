package cloudrun

import (
	"github.com/coreos/go-etcd/etcd"
)

type CloudRun struct {
	EtcdClient *etcd.Client
}
