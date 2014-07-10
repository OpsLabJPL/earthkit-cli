package fileset

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var SkipEntry = errors.New("skip this entry")

//func Build(rootPath string, allowExtLinks bool, genDigest bool) BuildResult {
func Build(builderCfg BuilderCfg, cachedFileSet *FileSet, patterns FileSetFilter) BuildResult {
	// Make the root path absolute if it isn't already
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	rootPath := builderCfg.RootPath
	rootPath = joinIfNotAbs(wd, rootPath)
	builderCfg.RootPath = rootPath
	rootInfo, err := os.Stat(rootPath)
	if err != nil {
		log.Fatal(err)
	}
	if (rootInfo.Mode() & os.ModeDir) == 0 {
		log.Fatal("root path of FileSet must be a directory: " + rootPath)
	}

	var cachedEntryMap EntryMap
	if cachedFileSet != nil {
		cachedEntryMap = cachedFileSet.Root.Flatten()
	}

	builder := &builder{
		&FileSet{CrTime: time.Now()},
		make([]string, 8, 8),
		make([]string, 16, 16),
		builderCfg,
		cachedEntryMap,
	}
	rootEntry := builder.visitFile(rootPath, rootInfo)

	if patterns != nil && len(patterns) > 0 && cachedFileSet != nil {
		log.Printf("creating fileset from a limited subset of the workspace...")
		newRootEntry := patterns.ApplyToEntryRoot(rootEntry)
		builder.fileSet.Root = newRootEntry
	} else {
		builder.fileSet.Root = rootEntry
	}
	return builder
}

func LoadJson(data []byte) (fileSet *FileSet, err error) {
	fileSet = new(FileSet)
	err = json.Unmarshal(data, fileSet)
	return
}

func LoadGzJson(data []byte) (fileSet *FileSet, err error) {
	fileSet = new(FileSet)
	buf := bytes.NewReader(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(gz)
	err = decoder.Decode(fileSet)
	return
}

// This method will fully consume the data from a reader and generate a hex string
// based on the SHA-256 digest of the bytes read.
func Hexdigest(reader io.Reader) (digest string, err error) {
	hash := sha256.New()
	_, err = io.Copy(hash, reader)
	if err == nil {
		digest = hex.EncodeToString(hash.Sum(nil))
	}
	return
}

func FileSetNameFromFile(filepath string) string {
	return strings.Replace(path.Base(filepath), ".json.gz", "", 1)
}

// Takes two paths and joins them if the second path is not an absolute path.
// If the second path is an absolute path, it is returned.
func joinIfNotAbs(basepath string, otherpath string) string {
	if filepath.IsAbs(otherpath) {
		return otherpath
	} else {
		return filepath.Join(basepath, otherpath)
	}
}
