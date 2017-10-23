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
	_dev     uint64
	_ino     uint64
	_suffix  int
	isCover  bool
	_date    *time.Time
	mu       *sync.RWMutex
	logfile  *os.File
	lg       *log.Logger
}

func filedev(file string) (uint64, uint64) {
	fileinfo, _ := os.Stat(file)
	stat := fileinfo.Sys().(*syscall.Stat_t)
	return stat.Dev, stat.Ino
}