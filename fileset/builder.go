package fileset

import (
	"errors"
	// "fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (bldr builder) Skipped() []string {
	return bldr.skipped
}

func (bldr builder) FileSet() *FileSet {
	return bldr.fileSet
}

// The visitFile method will determine the type of Entry to generate based on
// the provided os.FileInfo argument.  This method will return nil if the file was
// skipped due to it not being a directory, symbolic link, or regular file.
func (bldr *builder) visitFile(path string, info os.FileInfo) (entry *Entry) {
	switch {
	case info.IsDir():
		entry = bldr.visitDirectory(path, info)
	case (info.Mode() & os.ModeSymlink) != 0:
		entry = bldr.visitLink(path, info)
	case info.Mode().IsRegular():
		entry = bldr.visitRegularFile(path, info)
	default:
		bldr.skipped = append(bldr.skipped, path)
		log.Println("skipping irregular file: " + path)
	}
	if entry != nil {
		entry.Mode = info.Mode()
		entry.ModTime = info.ModTime()
		bldr.fileSet.Count++
	}
	return
}

func (bldr *builder) visitDirectory(path string, info os.FileInfo) (entry *Entry) {
	if bldr.shouldIgnore(info.Name()) {
		log.Println("Ignoring", info.Name())
		return
	}
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()
	entry = new(Entry)
	entry.Tree = make(EntryMap)
	for children, err := dir.Readdir(50); err == nil; children, err = dir.Readdir(50) {
		for _, childInfo := range children {
			childPath := filepath.Join(path, childInfo.Name())
			childEntry := bldr.visitFile(childPath, childInfo)
			if childEntry != nil {
				entry.Tree[childInfo.Name()] = childEntry
			}
		}
	}
	return
}

func (bldr *builder) visitRegularFile(path string, info os.FileInfo) (entry *Entry) {
	relPath, _ := filepath.Rel(bldr.cfg.RootPath, path)

	if bldr.shouldIgnore(info.Name()) {
		log.Println("Ignoring", info.Name())
		return
	}
	var (
		digest string
		err    error
	)
	if info.Size() > 0 && bldr.cfg.GenDigest == true {
		cachedEntry := bldr.CachedEntryMap[relPath]
		if cachedEntry != nil && cachedEntry.ModTime == info.ModTime() {
			digest = bldr.CachedEntryMap[relPath].Digest
		} else {
			var fp *os.File
			fp, err = os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
			digest, err = Hexdigest(fp)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	entry = new(Entry)
	entry.Size = info.Size()
	entry.Digest = digest
	bldr.fileSet.Size += entry.Size
	return
}

func (bldr *builder) visitLink(path string, info os.FileInfo) (entry *Entry) {
	if bldr.shouldIgnore(info.Name()) {
		log.Println("Ignoring", info.Name())
		return
	}
	target, err := os.Readlink(path)
	if err != nil {
		log.Fatal(err)
	}
	absTarget := joinIfNotAbs(bldr.cfg.RootPath, target)
	if !strings.HasPrefix(absTarget, bldr.cfg.RootPath) {
		if bldr.cfg.AllowExtLinks {
			log.Println("WARNING: Symbolic link points to external resource. The referenced file will not be included in the fileset.")
		} else {
			panic(errors.New("Workspace contains symbolic links pointing to external paths. They will not be included in the fileset. If this is okay, run with -f."))
		}
	}
	entry = new(Entry)
	entry.Target = target
	return
}

func (bldr *builder) shouldIgnore(name string) bool {
	for _, key := range bldr.cfg.IgnoreList {
		if key == name {
			return true
		}
	}
	return false
}
