package storage

import (
	"time"
)

type FileMeta struct {
	Name         string
	Size         int64
	LastModified time.Time
	ContentType  string
}
