package s3utils

import (
	"fmt"
	"github.com/opslabjpl/goamz/s3"
	"io"
	"os"
)

type S3Task interface {
	Perform()
}

// Contains info needed to do an upload
type S3DownloadTask struct {
	Bucket   *s3.Bucket
	FileName string
	S3ObjKey string
}

// Contains info needed to do an upload
type S3UploadTask struct {
	Bucket   *s3.Bucket
	FileName string
	S3ObjKey string
	Replace  bool
}

func (task S3DownloadTask) Perform() {
	//fmt.Println("Need to download", task.S3ObjKey, "to", task.FileName)

	obj, err1 := task.Bucket.GetReader(task.S3ObjKey)
	fo, err2 := os.Create(task.FileName)
	if err1 != nil {
		panic(err1)
	}
	if err2 != nil {
		panic(err2)
	}

	defer obj.Close()
	defer fo.Close()

	buf := make([]byte, 16384)
	for {
		// read a chunk
		n, err := obj.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := fo.Write(buf[:n]); err != nil {
			panic(err)
		}
	}
}

// Upload the file to S3
// TODO: error handling
// TODO: only use multi upload if file is large
func (task S3UploadTask) Perform() {
	exists, _ := task.Bucket.Exists(task.S3ObjKey)
	if exists && task.Replace != true {
		return
	}
	fmt.Println("Uploading", task.FileName)
	file, _ := os.Open(task.FileName)
	multi, _ := task.Bucket.Multi(task.S3ObjKey, "", s3.Private)
	parts, _ := multi.PutAll(file, 5242880)
	err := multi.Complete(parts)
	if err != nil {
		fmt.Println(err.Error())
	}
}
