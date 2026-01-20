//go:build !darwin && !windows && !linux

package scanner

import (
	"io/fs"
	"os"
	"sync/atomic"
	"time"

	"github.com/charlievieth/fastwalk"
)

func getDirSize(path string) (result, error) {
	var total atomic.Int64
	var fileScanned atomic.Int64
	walk := func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fileScanned.Add(1)
		info, err := d.Info()
		if err != nil {
			return err
		}
		total.Add(info.Size())
		return nil
	}
	err := fastwalk.Walk(&fastwalk.Config{Follow: false}, path, walk)
	return result{
		Size:         total.Load(),
		FilesScanned: fileScanned.Load(),
	}, err
}

func getLastModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
