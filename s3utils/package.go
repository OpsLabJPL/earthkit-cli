package s3utils

import (
	"github.com/opslabjpl/goamz.git/s3"
)

func S3ObjectExist(bucket *s3.Bucket, key string) bool {
	exists, _ := bucket.Exists(key)
	return exists
}
