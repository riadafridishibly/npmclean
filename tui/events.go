package tui

import (
	"context"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/riadafridishibly/npmclean/scanner"
)

func (a *App) trySendUIUpdate(f func()) {
	select {
	case a.uiUpdates <- f:
	default:
	}
}

// setRoot queues a SetRoot operation to avoid data races
func (a *App) setRoot(primitive cview.Primitive, focus bool) {
	a.app.QueueUpdateDraw(func() {
		a.app.SetRoot(primitive, focus)
	})
}

func (a *App) processProgressEvents(ctx context.Context) {
	progressChan := a.scanner.Progress()
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	var progress *scanner.ScanResult
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.scanner.Done():
			a.trySendUIUpdate(a.updateFinalStatus)
			return
		case progress = <-progressChan:
			if progress != nil && progress.Done {
				a.trySendUIUpdate(a.updateFinalStatus)
				return
			}
		case <-ticker.C:
			if a.scanner.IsRunning() && progress != nil {
				p := *progress
				a.trySendUIUpdate(func() { a.updateProgressStatus(&p) })
			}
		}
	}
}

func (a *App) processResultEvents(ctx context.Context) {
	resultsChan := a.scanner.Results()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.scanner.Done():
			a.trySendUIUpdate(a.updateFinalStatus)
			return
		case result, ok := <-resultsChan:
			if !ok {
				return
			}
			// TODO: We probably don't need to queue every update
			a.trySendUIUpdate(func() { a.handleResult(result) })
		case <-ticker.C:
			if a.scanner.IsRunning() {
				a.trySendUIUpdate(func() { a.updateStatus() })
			}
		}
	}
}
