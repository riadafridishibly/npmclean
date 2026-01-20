//go:build windows

// TODO: Test this on windows

package scanner

import (
	"io/fs"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charlievieth/fastwalk"
	"golang.org/x/sys/windows"
)

type fileID struct {
	VolumeSerial uint32
	IndexHigh    uint32
	IndexLow     uint32
}

func getDirSize(path string) (result, error) {
	var (
		total       atomic.Int64
		fileScanned atomic.Int64
		mu          sync.Mutex
	)
	seen := make(map[fileID]struct{})
	walk := func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fileScanned.Add(1)
		if d.IsDir() {
			return nil
		}
		// Use os.Open so we can query a HANDLE
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}

		// Use GetFileInformationByHandle to retrieve volume + index identity
		h := windows.Handle(f.Fd())
		var info windows.ByHandleFileInformation
		if err := windows.GetFileInformationByHandle(h, &info); err != nil {
			// On failure, fall back to naive size count
			total.Add(fi.Size())
			return nil
		}

		// Only dedupe on files with >1 link count or when identity matches seen
		id := fileID{
			VolumeSerial: info.VolumeSerialNumber,
			IndexHigh:    info.FileIndexHigh,
			IndexLow:     info.FileIndexLow,
		}

		mu.Lock()
		if _, exists := seen[id]; exists {
			mu.Unlock()
			// skip duplicate
			return nil
		}
		seen[id] = struct{}{}
		mu.Unlock()

		total.Add(fi.Size())

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
