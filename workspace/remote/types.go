package remote

import (
	"github.com/opslabjpl/goamz/s3"
)

type Remote struct {
	name   string
	bucket *s3.Bucket
}
