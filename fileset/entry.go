package fileset

import (
	"log"
	"path/filepath"
)

func (this Entry) Equal(other *Entry) bool {
	var equal bool
	equal = (this.Mode == other.Mode)
	equal = equal && this.ModTime.Equal(other.ModTime)
	equal = equal && (this.Size == other.Size)
	equal = equal && (this.Digest == other.Digest)
	equal = equal && (this.Target == other.Target)
	equal = equal && (len(this.Tree) == len(other.Tree))
	if equal {
		for name, thisEntry := range this.Tree {
			if otherEntry, ok := other.Tree[name]; ok {
				// Other tree contains same key, perform recursive step
				if !thisEntry.Equal(otherEntry) {
					// ... but the mapped value is not equal
					equal = false
					break
				}
			} else {
				// Other tree does not contain same key
				equal = false
				break
			}
		}
	}
	return equal
}

// Turns an Entry into a flattened map, mapping each entry's full path to
// the actual entry.  This does not "completely" flatten it, as directory
// entries will still have their Tree maps in tact.
func (entry *Entry) Flatten() EntryMap {
	m := make(EntryMap)
	flatten(``, entry, m)
	return m
}

func flatten(path string, entry *Entry, entryMap EntryMap) {
	for childName, childEntry := range entry.Tree {
		childPath := filepath.Join(path, childName)
		entryMap[childPath] = childEntry
		if len(childEntry.Tree) > 0 {
			flatten(childPath, childEntry, entryMap)
		}
	}
	return
}

// Returns a collection mapping file digests to the entries, allowing you to
// determine entries that are identical according to their digests.  The actual
// mapping is to a slice of EntryPath structs.  Empty files (size == 0) are
// omitted.
func (entry *Entry) DigestMap() DigestMap {
	digestMap := make(DigestMap)
	walkFn := func(fullPath string, entry *Entry) error {
		if len(entry.Digest) == 0 {
			return nil
		}
		ep := EntryPath{fullPath, entry}
		if entrySet, ok := digestMap[entry.Digest]; ok {
			digestMap[entry.Digest] = append(entrySet, ep)
		} else {
			entrySet = make([]EntryPath, 2)
			entrySet[0] = ep
			digestMap[entry.Digest] = entrySet
		}
		return nil
	}
	entry.Walk(walkFn)
	return digestMap
}

// Returns a DigestMap of entries from the given entry [tree] that don't
// yet exist in the provided DigestMap.  Only checks the existence of files
// that have a size > 0.
func (entry *Entry) MissingDigests(digests DigestMap) DigestMap {
	missing := make(DigestMap)
	walkFn := func(fullPath string, entry *Entry) error {
		if len(entry.Digest) == 0 {
			return nil
		}
		if _, digestExists := digests[entry.Digest]; !digestExists {
			ep := EntryPath{fullPath, entry}
			if entrySet, ok := missing[entry.Digest]; ok {
				missing[entry.Digest] = append(entrySet, ep)
			} else {
				entrySet = make([]EntryPath, 2)
				entrySet[0] = ep
				missing[entry.Digest] = entrySet
			}
		}
		return nil
	}
	entry.Walk(walkFn)
	return missing
}

// Performs a depth-first walk on an Entry, calling the provided WalkFunc for
// each child Entry (including the receiver Entry).  You can skip directory
// entries by having your walk function return SkipEntry.
func (entry *Entry) Walk(walkFn WalkFunc) {
	walk("", entry, walkFn)
}

// Returns a duplicate entry object containing the same metadata as the given
// entry object. The content of Tree is not duplicated, but a new EntryMap
// is initialized if the given entry object refers to a directory.
func (entry *Entry) DuplicateMetadata() (newEntry *Entry) {
	newEntry = new(Entry)
	newEntry.Mode = entry.Mode
	newEntry.ModTime = entry.ModTime
	newEntry.Size = entry.Size
	newEntry.Digest = entry.Digest
	newEntry.Target = entry.Target
	if entry.Mode.IsDir() {
		newEntry.Tree = make(EntryMap)
	}
	return
}

func (entry *Entry) EqualMetadata(entryTwo *Entry) bool {
	var equal bool
	equal = (entry.Mode == entryTwo.Mode)
	equal = equal && entry.ModTime.Equal(entryTwo.ModTime)
	equal = equal && (entry.Size == entryTwo.Size)
	equal = equal && (entry.Digest == entryTwo.Digest)
	equal = equal && (entry.Target == entryTwo.Target)
	return equal
}

func walk(fullPath string, entry *Entry, walkFn WalkFunc) {
	walkFn(fullPath, entry)
	for name, child := range entry.Tree {
		childPath := filepath.Join(fullPath, name)
		err := walkFn(childPath, child)
		if err == nil {
			if len(child.Tree) > 0 {
				walk(childPath, child, walkFn)
			}
		} else if err != SkipEntry {
			log.Printf("Ignoring invalid walk error:  %s\n", err)
		}
	}
	return
}
