//go:build linux

package scanner

import (
	"io/fs"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/charlievieth/fastwalk"
)

func getDirSize(path string) (int64, error) {
	var blocks atomic.Int64
	var mu sync.Mutex
	seen := make(map[DevIno]struct{})

	walk := func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		st := info.Sys().(*syscall.Stat_t)

		// If hardlinks, avoid double counting
		if st.Nlink > 1 {
			key := DevIno{Dev: uint64(st.Dev), Ino: uint64(st.Ino)}
			mu.Lock()
			if _, ok := seen[key]; ok {
				mu.Unlock()
				return nil
			}
			seen[key] = struct{}{}
			mu.Unlock()
		}

		// Count allocated blocks (includes files + dirs)
		blocks.Add(int64(st.Blocks))
		return nil
	}

	err := fastwalk.Walk(&fastwalk.Config{Follow: false}, path, walk)
	if err != nil {
		return 0, err
	}

	// POSIX st.Blocks is in 512-byte units
	return blocks.Load() * 512, nil
}

// Use modification time for consistency across platforms
func getLastModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
