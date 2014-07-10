package remote

import (
	"github.com/opslabjpl/goamz.git/s3"
	"path"
)

const (
	Prefix                  = ".earthkit/"
	WorkspacePrefix         = ".earthkit/%s/"
	WorkspaceFilesPrefix    = ".earthkit/%s/files/"
	WorkspaceFilesetsPrefix = ".earthkit/%s/filesets/"
)

func New(name string, bucket *s3.Bucket) *Remote {
	return &Remote{name, bucket}
}

func Workspaces(bucket *s3.Bucket) ([]string, error) {
	workspaces := make([]string, 0, 256)
	prefixes, errs := bucket.ListAllChildrenAsync(Prefix, "/")
	for prefix := range prefixes {
		workspaces = append(workspaces, path.Base(prefix))
	}
	err := <-errs
	return workspaces, err
}
