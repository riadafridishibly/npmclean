package scanner

import (
	"time"
)

type DevIno struct {
	Dev uint64
	Ino uint64
}

type ScanResult struct {
	Path        string
	Size        int64
	LastAccess  time.Time
	ScannedPath string // Current file being scanned
	FileCount   int64  // Total files scanned so far
	Error       error
	Done        bool
}
