package scanner

import (
	"context"
	"io/fs"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charlievieth/fastwalk"
)

// FIXME: Fix all the experimental mess

type NodeModuleInfo struct {
	Path           string
	Size           int64
	LastModifiedAt time.Time
	ScannedAt      time.Time
}

const (
	statusIdle int32 = iota
	statusRunning
)

type Scanner struct {
	rootPath string

	// Node modules meta data with size of the directory
	results chan *NodeModuleInfo

	// Progress events, as we progress through file tree
	progress chan *ScanResult

	// Signaling the consumers, if we're done you can stop consuming from the
	// results and progress channel
	doneChan chan struct{}

	status int32 // idle | running

	// atomic total file processed
	fileCount int64

	// Scanner start time to track how much time does it take to complete the scan
	startTime time.Time

	backlogsMu sync.Mutex
	backlogs   []string

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

func NewScanner(rootPath string) *Scanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scanner{
		rootPath:  rootPath,
		results:   make(chan *NodeModuleInfo, 100),
		progress:  make(chan *ScanResult, 100),
		doneChan:  make(chan struct{}),
		status:    statusIdle,
		fileCount: 0,
		startTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *Scanner) Start() {
	if !atomic.CompareAndSwapInt32(&s.status, statusIdle, statusRunning) {
		return
	}
	s.startTime = time.Now()
	// Reset context for new scan
	s.ctx, s.cancel = context.WithCancel(context.Background())
	go s.scan()
}

func (s *Scanner) Stop() {
	if !atomic.CompareAndSwapInt32(&s.status, statusRunning, statusIdle) {
		return // not running
	}

	// Cancel the context
	if s.cancel != nil {
		s.cancel()
	}

	// Wait for completion
	<-s.doneChan
}

func (s *Scanner) IsRunning() bool {
	return atomic.LoadInt32(&s.status) == statusRunning
}

func (s *Scanner) Progress() <-chan *ScanResult {
	return s.progress
}

func (s *Scanner) Results() <-chan *NodeModuleInfo {
	return s.results
}

func (s *Scanner) Done() <-chan struct{} {
	return s.doneChan
}

func (s *Scanner) FileCount() int64 {
	return atomic.LoadInt64(&s.fileCount)
}

func (s *Scanner) ElapsedTime() time.Duration {
	if s.startTime.IsZero() {
		return 0
	}
	return time.Since(s.startTime)
}

func (s *Scanner) calculateSize(path string) {
	size, err := GetDirectorySize(path)
	if err != nil {
		select {
		case s.progress <- &ScanResult{Error: err}:
		case <-s.ctx.Done():
			return
		default:
		}
		return
	}

	lastModified, err := GetLastModifiedAt(path)
	if err != nil {
		lastModified = time.Now()
	}

	info := &NodeModuleInfo{
		Path:           path,
		Size:           size,
		LastModifiedAt: lastModified,
		ScannedAt:      time.Now(),
	}

	select {
	case s.results <- info:
	case <-s.ctx.Done():
		return
	}

	fileCount := atomic.LoadInt64(&s.fileCount)
	select {
	case s.progress <- &ScanResult{Path: path, Size: size, FileCount: fileCount}:
	case <-s.ctx.Done():
	default:
	}
}

func (s *Scanner) processSizeCalculation(paths <-chan string) {
	defer close(s.doneChan) // We're done when result is done
	defer close(s.progress)
	defer close(s.results)

	// 1. Process from chan
Loop:
	for {
		select {
		case path, ok := <-paths:
			if !ok {
				break Loop
			}
			s.calculateSize(path)

		case <-s.ctx.Done():
			return
		}
	}

	// 2. Process from backlogs
	s.backlogsMu.Lock()
	backlogs := s.backlogs
	s.backlogs = nil
	s.backlogsMu.Unlock()
	for _, p := range backlogs {
		if err := s.ctx.Err(); err != nil {
			return
		}
		s.calculateSize(p)
	}

	s.progress <- &ScanResult{Done: true, FileCount: s.fileCount}
}

func (s *Scanner) addToBacklogs(path string) {
	s.backlogsMu.Lock()
	defer s.backlogsMu.Unlock()
	s.backlogs = append(s.backlogs, path)
}

// TODO: We have to implement two staged search here
// 1. First we want to find the node_modules folders as fast as possible
// 2. For each folder we want to claculate it's size
func (s *Scanner) scan() {
	conf := fastwalk.Config{Follow: false}

	nodeModulePaths := make(chan string, 128)
	go s.processSizeCalculation(nodeModulePaths)

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err := s.ctx.Err(); err != nil {
			return fs.SkipAll
		}

		if err != nil {
			select {
			case s.progress <- &ScanResult{Error: err}:
			case <-s.ctx.Done():
				return fs.SkipAll
			}
			return nil
		}

		fileCount := atomic.AddInt64(&s.fileCount, 1)
		// Send progress update every 10 files
		// TODO: Use a ticker for this
		if fileCount%10 == 0 {
			select {
			case s.progress <- &ScanResult{ScannedPath: path, FileCount: fileCount}:
			case <-s.ctx.Done():
				return fs.SkipAll
			default:
			}
		}

		if d.IsDir() && d.Name() == "node_modules" {
			select {
			case nodeModulePaths <- path:
			case <-s.ctx.Done():
				return fs.SkipAll
			default:
				s.addToBacklogs(path)
			}
			return fastwalk.SkipDir
		}

		return nil
	}

	if err := fastwalk.Walk(&conf, s.rootPath, walkFn); err != nil {
		select {
		case s.progress <- &ScanResult{Error: err}:
		case <-s.ctx.Done():
		default:
			// fallback to close
		}
	}

	close(nodeModulePaths)
}
