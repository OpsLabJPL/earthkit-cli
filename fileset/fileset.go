package fileset

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
)

func (fileSet FileSet) Json() (data []byte, err error) {
	data, err = json.Marshal(fileSet)
	return
}

func (fileSet FileSet) GzJson() (data []byte, err error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	encoder := json.NewEncoder(gz)
	err = encoder.Encode(fileSet)
	if err != nil {
		return
	}
	err = gz.Close()
	data = buf.Bytes()
	return
}

func (this FileSet) Equal(other *FileSet) bool {
	var equal bool
	equal = (this.Count == other.Count)
	equal = equal && (this.Size == other.Size)
	equal = equal && this.CrTime.Equal(other.CrTime)
	equal = equal && (this.Root.Equal(other.Root))
	return equal
}
