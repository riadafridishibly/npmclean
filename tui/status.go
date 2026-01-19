package tui

import (
	"fmt"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/dustin/go-humanize"
	"github.com/riadafridishibly/npmclean/scanner"
)

func (a *App) updateFinalStatus() {
	if a.scanner == nil {
		return
	}

	fileCount := a.scanner.FileCount()
	elapsed := a.scanner.ElapsedTime().Round(time.Second)

	status := fmt.Sprintf("[white] Found: %d items | Files scanned: %s | Elasped: %s | Total Claimable: %s ",
		len(a.items),
		humanize.Comma(fileCount),
		elapsed,
		humanize.Bytes(uint64(a.totalClaimableSize.Load())),
	)
	a.header.SetText(status)
	a.header.SetTextAlign(cview.AlignCenter)

	footerText := "[black] [s/S]: Start  ↑/↓: Navigate  i: Details  [d/D]: Delete  [q/Q]: Quit"
	a.footer.SetText(footerText)
	a.footer.SetTextAlign(cview.AlignCenter)
}

func (a *App) updateProgressStatus(progress *scanner.ScanResult) {
	if progress.Done {
		a.updateFinalStatus()
		return
	}

	if progress.Error != nil {
		a.header.SetText(fmt.Sprintf("[red] Error: %v", progress.Error))
		return
	}

	a.lastUpdate = time.Now()

	if progress.ScannedPath != "" {
		scanPath := a.replaceHomeWithTilde(progress.ScannedPath)
		w, _ := a.app.GetScreenSize()
		w = w - 10
		if len(scanPath) > w {
			scanPath = "..." + scanPath[len(scanPath)-w:]
		}
		a.footer.SetText(" [white]Scanning: [black]" + scanPath)
	}
}
