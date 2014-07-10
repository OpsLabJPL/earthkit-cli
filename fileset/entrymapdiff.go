package fileset

import(
	"path/filepath"
)

// Returns a new EntryMapDiff containing only entries that match against at least
// one pattern in a slice of glob patterns
func (diff EntryMapDiff) Filter(patterns []string) (newDiff EntryMapDiff) {
	newDiff.Added = filterEntryMap(diff.Added, patterns)
	newDiff.Updated = filterEntryMap(diff.Updated, patterns)
	newDiff.Removed = filterEntryMap(diff.Removed, patterns)
	return
}

// Returns a new EntryMap that contains only values that match one of the entries
// in a slice of glob patterns
func filterEntryMap(entryMap EntryMap, patterns []string) (newEntryMap EntryMap) {
	newEntryMap = make(EntryMap)
	for k, v := range entryMap {
		match := false
		for _, pattern := range patterns {
			result, _ := filepath.Match(pattern, k)
			if result {
				match = true
			}
		}
		if match {
			newEntryMap[k] = v
		}
	}
	return
}