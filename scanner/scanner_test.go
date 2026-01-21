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

	// Load cached results
	cached, err := s.LoadCachedResults()
	if err != nil {
		t.Logf("Error loading cached results: %v", err)
	} else {
		fmt.Printf("Loaded %d cached results\n", len(cached))
	}

	// Also check what GetAll returns
	if s.Cache() != nil {
		all, _ := s.Cache().GetAll()
		fmt.Printf("Cache has %d total entries\n", len(all))
	}

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
	<-s.Done()
	wg.Wait()
	s.Close()
}
