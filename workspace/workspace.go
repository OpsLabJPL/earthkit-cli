package workspace

import (
	"encoding/json"
	"fmt"
	"github.com/opslabjpl/earthkit-cli/config"
	"github.com/opslabjpl/earthkit-cli/fileset"
	"github.com/opslabjpl/earthkit-cli/workspace/remote"
	"github.com/opslabjpl/goamz.git/s3"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Handles init command
func (workspace *Workspace) Init(dir string, cloning bool) {
	// Sanity checks
	if stat, err := os.Stat(dir); err != nil {
		log.Fatal("No such directory: " + dir)
	} else if !stat.Mode().IsDir() {
		log.Fatal("%s is not a directory" + dir)
	}
	workspace.LocalRootDir = dir
	earthKitDir := filepath.Join(dir, EarthkitDir)

	if _, err := os.Stat(earthKitDir); err == nil {
		fmt.Printf("%s is already an Earthkit workspace.\n", dir)
		return
	}

	// check against S3 to verify workspace does not already exist
	if !cloning && workspace.Remote().Exists() {
		log.Fatal("There is already an existing workspace with the same name. Please use a different name.")
	}

	// Create earthkitrc
	os.Mkdir(earthKitDir, 0700)
	myjson, err := json.Marshal(workspace)
	if err != nil {
		fmt.Println("error:", err)
	}

	f, err := os.Create(filepath.Join(earthKitDir, "earthkitrc"))
	if err != nil {
		fmt.Println("error:", err)
	}
	f.Write(myjson)

	err = os.Mkdir(workspace.FilesetsDir(), 0700)
	if err != nil {
		log.Fatal(err)
	}

	workspace.SetUpDiscoveryUrl(cloning)
}

func (workspace *Workspace) FilesetsDir() string {
	return filepath.Join(workspace.LocalRootDir, EarthkitDir, "filesets")
}

// Returns the corresponding remote.Remote struct for this workspace.
func (workspace *Workspace) Remote() *remote.Remote {
	if workspace.remote_ == nil {
		auth := config.AWSAuth()
		myS3 := s3.New(auth, config.Region)
		bucket := myS3.Bucket(*config.S3_BUCKET)
		workspace.remote_ = remote.New(workspace.Name, bucket)
	}
	return workspace.remote_
}

// Handles push command for the given workspace and fileset
func (workspace *Workspace) Push(filesetName string, comment string, patterns fileset.FileSetFilter) {
	cache_exists := true
	var cachedFileSet *fileset.FileSet
	mdata, err := ioutil.ReadFile(filepath.Join(workspace.FilesetsDir(), "_current"))
	if err != nil {
		cache_exists = false
		// It's only a "real" error if the fileset exists and we couldn't load it
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}
	if cache_exists {
		cachedFileSet, err = fileset.LoadGzJson(mdata)
		if err != nil {
			log.Fatal(err)
		}
	}

	baseDir := workspace.LocalRootDir
	builderCfg := fileset.BuilderCfg{baseDir, false, true, []string{EarthkitDir}}
	result := fileset.Build(builderCfg, cachedFileSet, patterns)

	fileSet := result.FileSet()
	fileSet.Comment = comment
	digestMap := fileSet.Root.DigestMap()

	files := make(map[string]string)
	for digest, entries := range digestMap {
		fileName := filepath.Join(baseDir, entries[0].Path)
		files[fileName] = digest
	}

	remoteWs := workspace.Remote()
	remoteWs.Upload(files)

	// upload the fileset json
	data, err := fileSet.GzJson()
	if err != nil {
		panic(err)
	}
	filesetName = filesetName + ".json.gz"
	err = remoteWs.PutFileset(filesetName, data)
	if err != nil {
		panic(err)
	}
	err = workspace.cacheFileset(filesetName, data)
	if err != nil {
		log.Fatal(err)
	}
	err = workspace.setCurrentFileset(filesetName)
	if err != nil {
		log.Fatal(err)
	}
}

func (workspace *Workspace) Pull(filesetName string, patterns fileset.FileSetFilter) {
	// generate local fileset (without calculating checksum)
	builderCfg := fileset.BuilderCfg{workspace.LocalRootDir, false, false, []string{EarthkitDir}}
	buildRes := fileset.Build(builderCfg, nil, nil)
	localFileSet := buildRes.FileSet()
	localEntryMap := localFileSet.Root.Flatten()
	var diff fileset.EntryMapDiff

	// Get cached fileset if it exists
	cache_exists := true
	var cachedFileSet *fileset.FileSet
	var cachedEntryMap fileset.EntryMap
	data, err := ioutil.ReadFile(filepath.Join(workspace.FilesetsDir(), "_current"))
	if err != nil {
		cache_exists = false
		// It's only a "real" error if the fileset exists and we couldn't load it
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}
	if cache_exists {
		cachedFileSet, err = fileset.LoadGzJson(data)
		if err != nil {
			log.Fatal(err)
		}
		cachedEntryMap = cachedFileSet.Root.Flatten()
		// See what changes have been made by the user to the current local
		// workspace and warn user that they will be deleted and/or modified
		diff = cachedEntryMap.Diff(localEntryMap)
		if patterns != nil {
			// if patterns are specified, only files matching the patterns
			// will be overwritten
			selectiveDiff := diff.Filter(patterns)
			warnUser(workspace.LocalRootDir, selectiveDiff)
		} else {
			warnUser(workspace.LocalRootDir, diff)
		}
	}

	// record the current pattern if one is used
	if patterns != nil && len(patterns) > 0 {
		workspace.ClearPatternCache()
		workspace.UpdatePatternCache(patterns)
	}

	// Now get the fileset from S3
	auth := config.AWSAuth()
	myS3 := s3.New(auth, config.Region)
	bucket := myS3.Bucket(*config.S3_BUCKET)
	s3Prefix := config.WorkspacePrefix(workspace.Name)

	filesetGzFile := filesetName + ".json.gz"
	path := s3Prefix + "/filesets/" + filesetGzFile

	exists, _ := bucket.Exists(path)
	if !exists {
		log.Fatal("Fileset " + filesetName + " does not exist.")
	}

	data, err = bucket.Get(path)
	if err != nil {
		log.Fatal("Unable to fetch fileset")
	}
	remoteFileSet, err := fileset.LoadGzJson(data)
	if err != nil {
		log.Fatal("Unable to parse remote fileset")
	}
	remoteEntryMap := remoteFileSet.Root.Flatten()
	remoteEntryMap = workspace.Filter(remoteEntryMap, patterns)
	if !cache_exists {
		// If there was no cached fileset, give a warning to the user about what files will
		// go away with the incoming pull
		diff = remoteEntryMap.Diff(localEntryMap)
		if patterns != nil {
			// if patterns are specified, only files matching the patterns
			// will be overwritten
			selectiveDiff := diff.Filter(patterns)
			warnUser(workspace.LocalRootDir, selectiveDiff)
		} else {
			warnUser(workspace.LocalRootDir, diff)
		}
	}
	workspace.Cache(localEntryMap, cachedEntryMap, diff)
	workspace.DownloadNewDigests(remoteEntryMap)
	workspace.Wipe()
	workspace.Rebuild(remoteEntryMap)

	err = workspace.cacheFileset(filesetGzFile, data)
	if err != nil {
		log.Fatal(err)
	}
	err = workspace.setCurrentFileset(filesetGzFile)
	if err != nil {
		log.Fatal(err)
	}

	workspace.cleanCache(*config.CACHE_LIMIT)
}

// filters out fileset entries that don't conform to a specified slice
// of pattern globs
func (workspace *Workspace) Filter(remoteEntryMap fileset.EntryMap, patterns []string) fileset.EntryMap {
	if patterns == nil || len(patterns) < 1 {
		return remoteEntryMap
	}

	filteredEntryMap := make(fileset.EntryMap)

	for k, _ := range remoteEntryMap {
		match := false
		for _, pattern := range patterns {
			this_match, _ := filepath.Match(pattern, k)
			if this_match {
				match = true
			}
		}

		if match {
			// we need to back-add all the parent directories
			// as well as the target entry
			dirs := strings.Split(k, string(filepath.Separator))

			for i, _ := range dirs {
				curr_key := filepath.Join(dirs[:(i + 1)]...)
				filteredEntryMap[curr_key] = remoteEntryMap[curr_key]
			}
		}
	}

	numTotalEntries := len(remoteEntryMap)
	numRemainingEntries := len(filteredEntryMap)
	log.Println("filters specified,", numRemainingEntries, "of", numTotalEntries, "fileset entries match")

	return filteredEntryMap
}

// wipe out only the things that will be pulled
func (workspace *Workspace) SelectiveWipe(remoteEntryMap fileset.EntryMap) {
	for k, v := range remoteEntryMap {
		if !v.Mode.IsDir() {
			os.Remove(filepath.Join(workspace.LocalRootDir, k))
		}
	}
}

// wipe out everything
func (workspace *Workspace) Wipe() {
	earthkitPath := filepath.Join(workspace.LocalRootDir)
	file, err := os.Open(earthkitPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for filenames, err := file.Readdir(1000); err == nil; filenames, err = file.Readdir(1000) {
		for _, fileStat := range filenames {
			if fileStat.Name() == EarthkitDir {
				continue
			} else {
				os.RemoveAll(filepath.Join(workspace.LocalRootDir, fileStat.Name()))
			}
		}
	}
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func (workspace *Workspace) Rebuild(entryMap fileset.EntryMap) {
	cacheDir := filepath.Join(workspace.LocalRootDir, EarthkitDir, "cache")

	// This map is used for keeping track of digest file that has been moved
	// from cache to final workspace
	movedDigestMap := make(map[string]string)

	// Need to sort the entries since we need to create parent dirs first before
	// creating children entries
	keys := make([]string, len(entryMap))
	i := 0
	for k, _ := range entryMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		path := filepath.Join(workspace.LocalRootDir, k)
		entry := entryMap[k]
		// fmt.Println("creating", path, "for entry", entry)
		if entry.Mode.IsDir() {
			os.MkdirAll(path, entry.Mode)
		} else if entry.Target != "" {
			os.Symlink(entry.Target, path)
		} else {
			srcFile := filepath.Join(cacheDir, entry.Digest)

			// empty file. Just create
			if entry.Size == 0 {
				os.Create(path)
				// if src file doesn't exist, it's because we already
				// move it for another file. Let's just copy it then
			} else if _, err := os.Stat(srcFile); os.IsNotExist(err) {
				cp(movedDigestMap[srcFile], path)
			} else {
				os.Rename(srcFile, path)
				movedDigestMap[srcFile] = path
			}
		}
		os.Chmod(path, entry.Mode)
	}

	// Have to set mtime here at the end and not
	// in the loop above because creating entries will undesirely update the
	// parent dirs' mtime and atime
	for k, entry := range entryMap {
		path := filepath.Join(workspace.LocalRootDir, k)
		if !entry.Mode.IsRegular() {
			continue
		}
		err := os.Chtimes(path, entry.ModTime, entry.ModTime)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (workspace *Workspace) DownloadNewDigests(remoteEntryMap fileset.EntryMap) {
	cacheDir := filepath.Join(workspace.LocalRootDir, EarthkitDir, "cache")
	if _, err := os.Stat(cacheDir); err != nil {
		os.MkdirAll(cacheDir, 0700)
	}

	// Get list of digest to download
	digestSet := make(map[string]bool)
	for _, entry := range remoteEntryMap {
		if entry.Digest == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(cacheDir, entry.Digest)); err != nil {
			if os.IsNotExist(err) {
				digestSet[entry.Digest] = true
			} else {
				log.Fatal(err)
			}
		}
	}
	// Convert the set to a flat array
	digests := make([]string, len(digestSet))
	i := 0
	for digest, _ := range digestSet {
		digests[i] = digest
		i++
	}
	// download the needed digests files
	fmt.Println(digests)
	workspace.Remote().Download(cacheDir, digests)
}

func (workspace *Workspace) Cache(localEntryMap fileset.EntryMap, cachedEntryMap fileset.EntryMap, diff fileset.EntryMapDiff) {
	fmt.Println("Caching local workspace")
	cacheDir := filepath.Join(workspace.LocalRootDir, EarthkitDir, "cache")
	if _, err := os.Stat(cacheDir); err != nil {
		os.MkdirAll(cacheDir, 0700)
	}
	cached := make(map[string]bool)
	var digest string
	for k, v := range localEntryMap {

		if diff.Added[k] != nil || diff.Updated[k] != nil {
			println("Skip caching", k, "because it needs to be deleted")
			continue
		}
		if v.Size == 0 {
			continue
		}
		var srcFile string
		if v.Digest == "" {
			srcFile = filepath.Join(workspace.LocalRootDir, k)
			cachedEntry := cachedEntryMap[k]
			if cachedEntry != nil {
				v.Digest = cachedEntry.Digest
			} else {
				fmt.Println("Gen digest for new file", srcFile, "-", digest)
				v.Digest = genDigest(srcFile)
			}
		}
		// already backup a copy
		if cached[v.Digest] == true {
			continue
		}
		dstFile := filepath.Join(cacheDir, v.Digest)
		// Theoretically this should only fail if the destination already exists
		err := os.Rename(srcFile, dstFile)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		cached[v.Digest] = true
	}
}

func (workspace *Workspace) SetUpDiscoveryUrl(isClone bool) {
	var discoveryUrl []byte
	earthKitDir := filepath.Join(workspace.LocalRootDir, EarthkitDir)
	if isClone {
		discoveryUrl = workspace.Remote().GetDiscoveryURL()
	} else {
		discoveryUrl = genEtcdDiscoveryURL("https://discovery.etcd.io/new")
		workspace.Remote().PutDiscoveryURL(discoveryUrl)
	}
	f, err := os.Create(filepath.Join(earthKitDir, "discovery_url"))
	if err != nil {
		fmt.Println("error:", err)
	}
	f.Write(discoveryUrl)
}

func (workspace *Workspace) GetCurrentFileSetName() string {
	linkname := filepath.Join(workspace.FilesetsDir(), "_current")
	mpath, _ := os.Readlink(linkname)
	return fileset.FileSetNameFromFile(mpath)
}

// Write the fileset data out to the to the fileset cache directory using the given filename.
// The argument 'filename' is expected to be the name on disk of the fileset (e.g. "fileset1.json.gz").
func (workspace *Workspace) cacheFileset(filename string, data []byte) error {
	filesetPath := filepath.Join(workspace.FilesetsDir(), filename)
	fp, err := os.Create(filesetPath)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	_, err = fp.Write(data)
	return err
}

// Given a path to a fileset, creates a symbolic link called "_current" that points to the fileset.
// The argument 'filename' is expected to be the name on disk of the fileset (e.g. "fileset1.json.gz").
func (workspace *Workspace) setCurrentFileset(filename string) error {
	linkname := filepath.Join(workspace.FilesetsDir(), "_current")
	os.Remove(linkname)
	err := os.Symlink(filename, linkname)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (workspace *Workspace) UpdatePatternCache(newPatterns []string) {
	var allPatterns []string
	var patternCachePath string
	var patternJson []byte

	patternCachePath = filepath.Join(workspace.LocalRootDir, EarthkitDir, "patterns.cache")
	allPatterns = make([]string, 0)
	allPatterns = append(allPatterns, newPatterns...)
	allPatterns = append(allPatterns, workspace.patternCache_...)
	patternJson, jsonErr := json.Marshal(allPatterns)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	ioutil.WriteFile(patternCachePath, patternJson, 0644)
	workspace.patternCache_ = allPatterns
}

func (workspace *Workspace) ClearPatternCache() {
	var patternCachePath string
	patternCachePath = filepath.Join(workspace.LocalRootDir, EarthkitDir, "patterns.cache")
	os.Remove(patternCachePath)
}

// Does not handle race conditions. Only here as a helper function for
// cleaning up workspaces while doing development
func (workspace *Workspace) DeleteFileset(filesetName string) {
	filesPrefix := workspace.Remote().FilesPrefix()
	filesetToDelete := workspace.Remote().GetFileset(filesetName)
	filesetKeyToDelete := workspace.Remote().FilesetsPrefix() + filesetName + ".json.gz"

	// Map of all digests for this workspace. Entries map to 1 will be delete. Entries
	// map to 0 will be skip (e.g. digests that are still referenced by other filesets)
	var digests = make(map[string]int)

	digestMap := filesetToDelete.Root.DigestMap()
	for digest, _ := range digestMap {
		digests[digest] = 1
	}

	println("Determining what to delete...")

	// Get all filesets
	filesetKeys, _ := workspace.Remote().Filesets()
	auth := config.AWSAuth()
	myS3 := s3.New(auth, config.Region)
	bucket := myS3.Bucket(*config.S3_BUCKET)

	for _, key := range filesetKeys {
		if key.Key == filesetKeyToDelete {
			continue
		}
		data, err := bucket.Get(key.Key)
		if err != nil {
			log.Fatal("Unable to fetch fileset")
		}
		remoteFileSet, err := fileset.LoadGzJson(data)
		digestMap := remoteFileSet.Root.DigestMap()
		for digest, _ := range digestMap {
			digests[digest] = 0
		}
	}

	for digest, deleteOrNot := range digests {
		if deleteOrNot == 1 {
			bucket.Del(filesPrefix + digest)
			println("Deleting", filesPrefix+digest)
		}
	}
	bucket.Del(filesetKeyToDelete)
}

func (workspace *Workspace) cleanCache(cacheLimit int64) {
	cacheDir := workspace.cacheDir()

	// See how big it is
	var totalSize int64 = 0
	files, _ := ioutil.ReadDir(cacheDir)
	sort.Sort(filesDateSort(files))
	for _, f := range files {
		totalSize += f.Size()
	}

	// Delete files until totalSize is smaller than limit
	var fileToDelete os.FileInfo
	for totalSize > cacheLimit {
		fileToDelete, files = files[len(files)-1], files[:len(files)-1]

		os.Remove(filepath.Join(cacheDir, fileToDelete.Name()))
		totalSize -= fileToDelete.Size()
	}
}

func (workspace *Workspace) cacheDir() string {
	return filepath.Join(workspace.LocalRootDir, EarthkitDir, "cache")
}
