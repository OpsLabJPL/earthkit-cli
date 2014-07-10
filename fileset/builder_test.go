package fileset

import (
	"io/ioutil"
	"testing"
	"time"
)

const (
	tree1Path string = "testdata/tree1"
)

// Sets the FileSet's creation time to 0 and zeroes out all entries' ModTime fields
func zeroTime(fileSet *FileSet) {
	zero := time.Time{}
	fileSet.CrTime = zero
	walkFn := func(fullPath string, entry *Entry) error {
		entry.ModTime = zero
		return nil
	}
	fileSet.Root.Walk(walkFn)
}

func loadReference() (fileSet *FileSet, err error) {
	data, err := ioutil.ReadFile(tree1Path + ".json")
	if err != nil {
		return
	}
	fileSet, err = LoadJson(data)
	return
}

func TestBuilder_DisallowExtLinks(t *testing.T) {
	// This should fail due to the existence of external links, so
	// defer a function that fails the test if there is no error.
	defer func() {
		if err := recover(); err != nil {
			// Success
		} else {
			t.Fail()
		}
	}()
	buildCfg := BuilderCfg{RootPath: tree1Path, GenDigest: true}
	Build(buildCfg, nil)
}

func TestBuilder_AllowExtLinks(t *testing.T) {
	// Load reference fileset
	refFileSet, err := loadReference()
	if err != nil {
		t.Fatal(err)
	}
	// Now build a FileSet allowing external links
	buildCfg := BuilderCfg{RootPath: tree1Path, AllowExtLinks: true, GenDigest: true}
	result := Build(buildCfg, nil)
	realFileSet := result.FileSet()
	// Zero out all the time fields
	zeroTime(realFileSet)
	if !realFileSet.Equal(refFileSet) {
		t.Log("The generated FileSet does not match the one loaded from JSON")
		t.Fail()
	}
}
