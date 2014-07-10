package workspace

import (
	"encoding/json"
	"fmt"
	"github.com/opslabjpl/earthkit-cli/fileset"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const EarthkitDir = ".earthkit"

// Starting at the given dir, traverse up to the root dir
// and generate a Workspace struct
func GetWorkspace(dir string) (workspace Workspace) {
	dir, _ = filepath.Abs(dir)
	earthkitRCPath, rootDir := findEarthkitRC(dir)

	earthkitRC, _ := ioutil.ReadFile(earthkitRCPath)
	json.Unmarshal(earthkitRC, &workspace)
	workspace.LocalRootDir = rootDir

	patternCachePath := filepath.Join(rootDir, EarthkitDir, "patterns.cache")
	var patternCache []string
	patternCacheJson, patternCacheErr := ioutil.ReadFile(patternCachePath)
	if patternCacheErr == nil {
		json.Unmarshal(patternCacheJson, &patternCache)
	}
	workspace.patternCache_ = patternCache

	return
}

func New(name, dir string) *Workspace {
	return &Workspace{name, dir, nil, []string{}}
}

// Traverse up the tree, loking for earthkitrc
func findEarthkitRC(dir string) (earthKitRC string, rootDir string) {
	earthKitRC = filepath.Join(dir, EarthkitDir, "earthkitrc")
	if _, err := os.Stat(earthKitRC); err == nil {
		return earthKitRC, dir
	} else if dir == "/" {
		return "", ""
	}

	return findEarthkitRC(filepath.Dir(dir))
}

func cp(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func warnUser(localRootDir string, diff fileset.EntryMapDiff) {
	need_confirmation := false
	if len(diff.Added) > 0 || len(diff.Updated) > 0 {
		fmt.Println("The following files/directories will be deleted or updated:")
		for k, _ := range diff.Added {
			toBeDeleted := filepath.Join(localRootDir, k)
			fmt.Println(toBeDeleted)
		}
		for k, _ := range diff.Updated {
			toBeUpdated := filepath.Join(localRootDir, k)
			fmt.Println(toBeUpdated)
		}
		need_confirmation = true
	}
	if need_confirmation {
		var input string
		fmt.Print("Do you want to continue (y/N)? ")
		fmt.Scanf("%s", &input)
		if input != "y" && input != "Y" {
			os.Exit(0)
		}
	}
}

func genDigest(path string) (digest string) {
	info, err := os.Stat(path)
	if err != nil {
		panic(err.Error())
	}

	if info.Size() > 0 {
		var fp *os.File
		fp, err = os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer fp.Close()
		digest, err = fileset.Hexdigest(fp)
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}

func genEtcdDiscoveryURL(newUrl string) []byte {
	resp, err := http.Get(newUrl)
	if err != nil {
		panic(err.Error())
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	return body
}

func (f filesDateSort) Len() int {
	return len(f)
}

func (f filesDateSort) Less(i, j int) bool {
	return f[i].ModTime().Unix() <= f[j].ModTime().Unix()
}

func (f filesDateSort) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
