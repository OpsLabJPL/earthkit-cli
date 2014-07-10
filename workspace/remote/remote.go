package remote

import (
	"fmt"
	"github.com/opslabjpl/earthkit-cli/fileset"
	"github.com/opslabjpl/earthkit-cli/s3utils"
	"github.com/opslabjpl/goamz/s3"
	"github.com/opslabjpl/gotx/tx"
	txFile "github.com/opslabjpl/gotx/tx/file"
	txS3 "github.com/opslabjpl/gotx/tx/s3"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// Determine if the workspace already exists in S3 by listing
// for object that matches the workspace prefix
func (this *Remote) Exists() bool {
	wsPrefix := this.WorkspacePrefix()
	data, err := this.bucket.List(wsPrefix, "", "", 1)

	if err != nil {
		panic(err.Error())
	}

	return len(data.Contents) > 0
}

func (this *Remote) WorkspacePrefix() string {
	return fmt.Sprintf(WorkspacePrefix, this.name)
}

func (this *Remote) FilesPrefix() string {
	return fmt.Sprintf(WorkspaceFilesPrefix, this.name)
}

func (this *Remote) FilesetsPrefix() string {
	return fmt.Sprintf(WorkspaceFilesetsPrefix, this.name)
}

func (this *Remote) Filesets() ([]s3.Key, error) {
	prefix := this.FilesetsPrefix()
	keys := make([]s3.Key, 0, 64)
	results, errs := this.bucket.ListAllAsync(prefix)
	for key := range results {
		keys = append(keys, key)
	}
	err := <-errs
	return keys, err
}

func (this *Remote) LatestFileset() (filesetName string, err error) {
	s3Keys, err := this.Filesets()
	if err != nil {
		return
	}

	timestamp := ""

	for _, key := range s3Keys {
		if timestamp < key.LastModified {
			filesetName = strings.Replace(path.Base(key.Key), ".json.gz", "", 1)
			timestamp = key.LastModified
		}
	}
	return
}

func (this *Remote) Upload(files map[string]string) {
	txMgr := tx.NewTxMgr(32, false, false)
	cTxDone := txMgr.NewTxChan(16)

	prefix := this.FilesPrefix()
	for fileName, digest := range files {
		key := path.Join(prefix, digest)

		// No need to upload if the object is already there
		if s3utils.S3ObjectExist(this.bucket, key) {
			println("Skipping", key)
			continue
		}

		f, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		src, err := txFile.NewSource(f, 5*1024*1024)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		dst := txS3.NewDestination(this.bucket, key)
		txMgr.Add(src, dst)
	}

	// 'quit' channel used to coordinate progress goroutine and main goroutine.
	quit := make(chan bool)
	// Start a goroutine for outputting progress
	go progressPrinter(txMgr.KnownSize(), txMgr.Wx, quit)

	txMgr.Start()

	// Wait for all transfers to finish, aborting if there is even one error.
	var err error = nil
	for t := range cTxDone {
		if err != nil {
			continue
		}
		if t.Error() != nil {
			err = t.Error()
			log.Printf("Aborting upload due to error (this could take a few seconds).\n\terror: %s", err)
			txMgr.Abort()
		}
	}
	quit <- true
	<-quit
	if err != nil {
		log.Fatalf("Upload failed.")
	}
	fmt.Println("Progress: Completed")
}

func (this *Remote) Download(localPath string, digests []string) {
	txMgr := tx.NewTxMgr(32, false, false)
	cTxDone := txMgr.NewTxChan(16)

	prefix := this.FilesPrefix()
	for _, digest := range digests {
		key := path.Join(prefix, digest)
		fileName := path.Join(localPath, digest)
		// Add as many files as you want here
		src, err := txS3.NewSource(this.bucket, key, 5*1024*1024)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 666)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		dst, err := txFile.NewDestination(f)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		txMgr.Add(src, dst)
	}

	// 'quit' channel used to coordinate progress goroutine and main goroutine.
	quit := make(chan bool)
	// Start a goroutine for outputting progress
	go progressPrinter(txMgr.KnownSize(), txMgr.Rx, quit)

	txMgr.Start()

	// Wait for all transfers to finish, aborting if there is even one error.
	var err error = nil
	for t := range cTxDone {
		if err != nil {
			continue
		}
		if t.Error() != nil {
			err = t.Error()
			log.Printf("Aborting download due to error (this could take a few seconds).\n\terror: %s", err)
			txMgr.Abort()
		}
	}
	quit <- true
	<-quit
	if err != nil {
		// If the download was aborted, try to remove all the digests we were attempting to download.
		for _, digest := range digests {
			fileName := path.Join(localPath, digest)
			os.Remove(fileName)
		}
		log.Fatalf("Download failed.")
	}
	fmt.Println("Progress: Completed")
}

func (this *Remote) PutFileset(name string, data []byte) error {
	key := path.Join(this.FilesetsPrefix(), name)
	return this.bucket.Put(key, data, "application/json", s3.Private, s3.Options{})
}

func (this *Remote) PutDiscoveryURL(data []byte) error {
	key := path.Join(this.WorkspacePrefix(), "discovery_url")
	return this.bucket.Put(key, data, "", s3.Private, s3.Options{})
}

func (this *Remote) GetDiscoveryURL() []byte {
	key := path.Join(this.WorkspacePrefix(), "discovery_url")
	discoveryUrl, err := this.bucket.Get(key)
	if err != nil {
		log.Fatal("Unable to fetch discovery url")
	}
	return discoveryUrl
}

func (this *Remote) GetFileset(filesetName string) *fileset.FileSet {
	prefix := this.FilesetsPrefix()
	filesetGzFile := filesetName + ".json.gz"
	path := prefix + filesetGzFile

	exists, _ := this.bucket.Exists(path)
	if !exists {
		log.Fatal("Fileset " + filesetName + " does not exist.")
	}

	data, err := this.bucket.Get(path)
	if err != nil {
		log.Fatal("Unable to fetch fileset")
	}
	remoteFileSet, err := fileset.LoadGzJson(data)
	if err != nil {
		log.Fatal("Unable to parse remote fileset")
	}
	return remoteFileSet
}

// A function for printing the progress of a transfer while it's happening. This is intented to be
// launched as a goroutine.  'quitChan' should be an unbuffered channel used for coordinating the
// shutdown of the goroutine.  'quitChan' should be sent a value when the goroutine should stop.
// It will send a value back to verify that it is stopping.
func progressPrinter(knownSize int64, progress func() int64, quitChan chan bool) {
	lastWx := int64(-1)
	for {
		select {
		case <-quitChan:
			quitChan <- true
			return
		default:
			wx := progress()
			if wx != lastWx {
				lastWx = wx
				fmt.Printf("Progress: %d/%d (%.2f%%)\n", wx, knownSize, float64(wx)/float64(knownSize)*100)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
