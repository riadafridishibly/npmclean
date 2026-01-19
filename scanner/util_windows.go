//go:build windows

package scanner

import (
	"io/fs"
	"os"
	"sync/atomic"
	"time"

	"github.com/charlievieth/fastwalk"
)

// TODO: Probably use windows specific APIs for this one!

func getDirSize(path string) (int64, error) {
	var total int64
	walk := func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		atomic.AddInt64(&total, info.Size())
		return nil
	}
	err := fastwalk.Walk(&fastwalk.Config{Follow: false}, path, walk)
	return total, err
}

func getLastModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	// Portable systems don’t expose atime API reliably
	// So fallback to modification time
	return info.ModTime(), nil
}
