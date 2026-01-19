package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestScanner(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	s := NewScanner(filepath.Join(home, "git-clones"))
	wg := sync.WaitGroup{}
	wg.Add(2) // Wait for both goroutines
	go func() {
		defer wg.Done()
		for res := range s.Results() {
			fmt.Println(res.Size, res.Path)
		}
	}()
	go func() {
		defer wg.Done()
		for range s.Progress() {
			// Discard progress updates to prevent blocking
		}
	}()
	s.Start()
	wg.Wait()
}
