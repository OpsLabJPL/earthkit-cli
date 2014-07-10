package fileset

import (
	"os"
	"time"
)

type EntryMap map[string]*Entry

type EntryMapDiff struct {
	Added   EntryMap
	Removed EntryMap
	Updated EntryMap
}

// The Entry struct doesn't need a type field since the Mode field encodes
// mode bits which indicate whether it is a link, dir, etc.
type Entry struct {
	// Fields common to all entries
	Mode os.FileMode `json:"mode"`
	// User    string      `json:"user"`
	// Group   string      `json:"group"`
	ModTime time.Time `json:"mtime"`
	// Directory attributes
	Tree EntryMap `json:"_tree,omitempty"`
	// File attributes
	Size   int64  `json:"size,omitempty"`
	Digest string `json:"_digest,omitempty"`
	// Target attributes
	Target string `json:"target,omitempty"`
}

type FileSet struct {
	Size    int64     `json:"size"`
	Count   int64     `json:"count"`
	Root    *Entry    `json:"root"`
	CrTime  time.Time `json:"crtime"`
	Comment string    `json:"comment,omitempty"`
}

type BuilderCfg struct {
	RootPath      string
	AllowExtLinks bool
	GenDigest     bool
	IgnoreList    []string
}

type builder struct {
	fileSet        *FileSet
	skipped        []string
	visited        []string
	cfg            BuilderCfg
	CachedEntryMap EntryMap
}

type BuildResult interface {
	FileSet() *FileSet
	Skipped() []string
}

// Allows you to associate an Entry with its full path
type EntryPath struct {
	Path  string
	Entry *Entry
}

type DigestMap map[string][]EntryPath

type WalkFunc func(fullPath string, entry *Entry) error

type FileSetFilter []string

type FileSubSet struct {
	Patterns []string
	FullTree *Entry
}