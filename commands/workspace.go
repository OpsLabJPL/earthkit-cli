package commands

import (
	"fmt"
	"github.com/opslabjpl/earthkit-cli/config"
	"github.com/opslabjpl/earthkit-cli/fileset"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"github.com/opslabjpl/earthkit-cli/workspace/remote"
	"github.com/opslabjpl/goamz/s3"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

func WorkspaceCommand(args []string) {
	var action string

	if len(args) < 1 {
		action = "list"
	} else {
		action = args[0]
	}

	switch action {
	case "list":
		listRemoteWorkspaces()
	case "status":
		localWorkspaceStatus()
	default:
		panic("Unsupported action for workspace command")
	}
}

func listRemoteWorkspaces() {
	auth := config.AWSAuth()
	myS3 := s3.New(auth, config.Region)
	bucket := myS3.Bucket(*config.S3_BUCKET)

	workspaces, err := remote.Workspaces(bucket)
	if err != nil {
		log.Fatal(err)
	}

	if len(workspaces) == 0 {
		fmt.Println("No remote workspaces found.")
		return
	}

	fmt.Println("Remote workspaces:")
	for _, wsName := range workspaces {
		fmt.Printf("  %s\n", path.Base(wsName))
	}
}

func localWorkspaceStatus() {
	ws := workspace.GetWorkspace(".")
	fmt.Println("Current workspace:", ws.Name)

	currentFileSetFile, err := os.Readlink(filepath.Join(ws.FilesetsDir(), "_current"))
	if err != nil {
		log.Fatal(err)
	}

	ext := ".json.gz"
	currentFileSetName := currentFileSetFile[0 : len(currentFileSetFile)-len(ext)]
	fmt.Println("Current fileset:", currentFileSetName)

	var cachedFileSet *fileset.FileSet
	data, err := ioutil.ReadFile(filepath.Join(ws.FilesetsDir(), "_current"))
	if err != nil {
		log.Fatal(err)
	} else {
		cachedFileSet, err = fileset.LoadGzJson(data)
		if err != nil {
			log.Fatal(err)
		}

	}

	baseDir := ws.LocalRootDir
	builderCfg := fileset.BuilderCfg{baseDir, false, true, []string{workspace.EarthkitDir}}
	result := fileset.Build(builderCfg, nil, nil)
	localFileSet := result.FileSet()
	diff := cachedFileSet.Root.Flatten().Diff(localFileSet.Root.Flatten())

	fmt.Println("Outstanding changes:")
	if len(diff.Added) == 0 && len(diff.Removed) == 0 && len(diff.Updated) == 0 {
		fmt.Println("none")
	}
	for k, _ := range diff.Added {
		fmt.Println("  (a)", k)
	}
	for k, _ := range diff.Removed {
		fmt.Println("  (d)", k)
	}
	for k, _ := range diff.Updated {
		fmt.Println("  (m)", k)
	}
}
