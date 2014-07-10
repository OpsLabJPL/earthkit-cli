package fileset

// Diff the two entry maps. Generate list of entries that are added, removed and updated
// in the second entry map
func (thisMap EntryMap) Diff(otherMap EntryMap) (diff EntryMapDiff) {
	diff.Added = make(EntryMap)
	diff.Removed = make(EntryMap)
	diff.Updated = make(EntryMap)
	for key, entry := range otherMap {
		if thisMap[key] == nil {
			diff.Added[key] = entry
		} else if isUpdated(thisMap[key], entry) {
			diff.Updated[key] = entry
		}
		// delete(thisMap, key)
	}
	for key, entry := range thisMap {
		if otherMap[key] == nil {
			diff.Removed[key] = entry
		}
	}
	return
}

// TODO: Possibly ditch this function and just refactor Entry.Equal() to allow one to specify whitelist or blacklist fields when checking equality
func isUpdated(entry1 *Entry, entry2 *Entry) bool {
	if entry1.Target != entry2.Target {
		return true
	}
	if entry1.Mode != entry2.Mode {
		return true
	}

	// for mtime, we only care about regular file
	if entry2.Mode.IsRegular() && entry2.ModTime.Unix() != entry1.ModTime.Unix() {
		return true
	}

	return false
}
