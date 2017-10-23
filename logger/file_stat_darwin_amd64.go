package logger

import (
    "time"
    "sync"
    "os"
    "log"
)

type _FILE struct {
	dir      string
	filename string
	_dev     int32
	_ino     uint64
	_suffix  int
	isCover  bool
	_date    *time.Time
	mu       *sync.RWMutex
	logfile  *os.File
	lg       *log.Logger
}