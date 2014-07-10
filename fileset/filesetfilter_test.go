package fileset

import (
	"fmt"
	"time"
	"os"
	"testing"
	"path/filepath"
)

const (
	STATIC_PATH string = "testdata/tree-remote"
	CHANGES_PATH string = "testdata/tree-local"
)

func GetDirectoryRootEntry(path string) *Entry {
	absPath, _ := filepath.Abs(path)
	fmt.Println(absPath)
	rootInfo, rootErr := os.Stat(absPath)
	if rootErr != nil {
		fmt.Println("uhhh")
	}
	builderCfg := BuilderCfg{path, false, true, []string{path + "/.earthkit"}}
	builder := &builder{
		&FileSet{CrTime: time.Now()},
		make([]string, 8, 8),
		make([]string, 16, 16),
		builderCfg,
		nil,
	}
	rootEntry := builder.visitFile(path, rootInfo)
	return rootEntry
}

func TestFileSubSet(t *testing.T) {
	filter := FileSetFilter{"*.txt", "test/sub2/*"}
	staticRoot := GetDirectoryRootEntry(STATIC_PATH)
	changesRoot := GetDirectoryRootEntry(CHANGES_PATH)

	newRoot := filter.ApplyToFileSet(staticRoot, changesRoot)
	fmt.Println(">>>>>")
	debugTree(newRoot)
	fmt.Println("<<<<<")
}