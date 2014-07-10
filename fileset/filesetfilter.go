package fileset

import (
	"fmt"
	"log"
	"strings"
	"path/filepath"
)

const (
	FSF_DEBUG = false
)

// This function applies filters to a tree of Entry objects. It returns a new Entry
// tree.
func (filter FileSetFilter) ApplyToEntryRoot(root *Entry) (newRoot *Entry) {
	matchPaths := getMatchingFiles(filter, root)
	newRoot = makeNewEntryTree(root, matchPaths)
	return
}

// This function updates a FileSet's Entry tree based on a FileSetFilter and a
// tree of changed Entries that must be applied. The result is a new Entry tree.
// Because a user can selectively pull parts of a FileSet to his/her local workspace,
// this mechanism is needed to push changes that occured within this subset of files.
func (filter FileSetFilter) ApplyToFileSet(fullRoot *Entry, changesRoot *Entry) (newRoot *Entry) {
	staticPaths := getStaticFiles(filter, fullRoot)
	newRoot = makeNewEntryTree(fullRoot, staticPaths)
	mergeChangeTree(newRoot, changesRoot)
	return
}

func getMatchingFiles(filter FileSetFilter, root *Entry) (matchPaths []string) {
	matchPaths = make([]string, 0)
	walkFn := func(fullPath string, entry *Entry) error {
		if(isMatch(fullPath, filter)) {
			matchPaths = append(matchPaths, fullPath)
		}
		return nil
	}
	root.Walk(walkFn)
	return
}

func getStaticFiles(filter FileSetFilter, root *Entry) (staticPaths []string) {
	staticPaths = make([]string, 0)
	walkFn := func(fullPath string, entry *Entry) error {
		if(!entry.Mode.IsDir() || entry.Tree == nil || len(entry.Tree) == 0) {
			// files or leaf dirs, need to do a pattern test
			if(!isMatch(fullPath, filter)) {
				staticPaths = append(staticPaths, fullPath)
			}
		}
		return nil
	}
	root.Walk(walkFn)
	return
}

func makeNewEntryTree(srcRoot *Entry, staticPaths []string) (newRoot *Entry) {
	entries := srcRoot.Flatten()
	newRoot = srcRoot.DuplicateMetadata()
	for _, path := range staticPaths {
		pathSlice := strings.Split(path, string(filepath.Separator))
		err := entryTreeDeepInsert(newRoot, entries, pathSlice)
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}

func mergeChangeTree(dstTreeRoot *Entry, changeTreeRoot *Entry) {
	changeEntries := changeTreeRoot.Flatten()
	walkFn := func(fullPath string, entry *Entry) error {
		if(!entry.Mode.IsDir() || entry.Tree == nil || len(entry.Tree) == 0) {
			pathSlice := strings.Split(fullPath, string(filepath.Separator))
			err := entryTreeDeepInsert(dstTreeRoot, changeEntries, pathSlice)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	}
	changeTreeRoot.Walk(walkFn)
}

func entryTreeDeepInsert(root *Entry, srcEntries EntryMap, path []string) error {
	currEntry := root
	for i, name := range path {
		subPath := path[0:(i+1)]
		subPathStr := filepath.Join(subPath...)
		if(currEntry.Tree[name] == nil) {
			// this part of the path is not in the tree, need to insert it
			if FSF_DEBUG {
				fmt.Println("attempting insert of '" + name + "' entry")
				fmt.Println("lookup path is", subPath)
			}
			srcEntry := srcEntries[subPathStr]
			if srcEntry == nil {
				return fmt.Errorf("cannot find entry information for:", subPathStr)
			}
			newEntry := srcEntry.DuplicateMetadata()
			insertResult := entryTreeSingleInsert(root, newEntry, subPath)
			if insertResult != nil {
				return insertResult
			}
			currEntry = newEntry
			if FSF_DEBUG {
				fmt.Println("tree insert of '" + subPathStr + "' successful")
			}
		} else if(!currEntry.Tree[name].EqualMetadata(srcEntries[subPathStr])) {
			// the metadata is just out of date for this node, need to update
			if FSF_DEBUG {
				fmt.Println("node '" + subPathStr + "' is outdated, updating...")
			}
			currEntry.Mode = srcEntries[subPathStr].Mode
			currEntry.ModTime = srcEntries[subPathStr].ModTime
			currEntry.Size = srcEntries[subPathStr].Size
			currEntry.Digest = srcEntries[subPathStr].Digest
			currEntry.Target = srcEntries[subPathStr].Target
			currEntry = currEntry.Tree[name]
		} else {
			// this node is fine
			if FSF_DEBUG {
				fmt.Println("entry node for '" + name + "' already exists")
			}
			currEntry = currEntry.Tree[name]
		}
	}
	return nil
}

func entryTreeSingleInsert(root *Entry, newEntry *Entry, path []string) error {
	if len(path) == 0 {
		return nil
	}
	currEntry := root
	for _, elemName := range path[0:len(path)-1] {
		if currEntry.Tree[elemName] == nil {
			return fmt.Errorf("can't insert, parent entries do not all exist")
		} else {
			currEntry = currEntry.Tree[elemName]
		}
	}
	currEntry.Tree[path[len(path)-1]] = newEntry
	return nil
}

func debugTree(root *Entry) {
	flatRoot := root.Flatten()
	for path, _ := range flatRoot {
		fmt.Println(path)
	}
}

func isMatch(s string, patterns []string) (match bool) {
	match = false
	for _, pattern := range patterns {
		this_match, _ := filepath.Match(pattern, s)
		if this_match {
			match = true
		}
	}
	return
}