package tui

import (
	"fmt"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/dustin/go-humanize"
	"github.com/riadafridishibly/npmclean/scanner"
)

func headerStartupStatus(theme *Theme, path string) string {
	return fmt.Sprintf(" Ready to scan. Press '[%s]s[-]' to start scanning: [::b][%s]%s[::-][-]",
		theme.darkGray.String(), theme.orange.String(), path)
}

func headerStatus(theme *Theme, items, fileCount, totalClaimableSize int64, elapsed time.Duration, done bool) string {
	if elapsed.Seconds() > 1 {
		elapsed = elapsed.Round(time.Second)
	} else {
		elapsed = elapsed.Round(time.Millisecond)
	}
	s := "Scanning"
	if done {
		s = "Found"
	}
	return fmt.Sprintf(" %s: [%s]%d[-] items | Files scanned: [%s]%s[-] | Elasped: [%s]%s[-] | Total Claimable: [::b][%s]%s[::-][-] ",
		s,
		theme.darkGray.String(), items,
		theme.darkGray.String(), humanize.Comma(fileCount),
		theme.darkGray.String(), elapsed,
		theme.darkGray.String(), humanize.Bytes(uint64(totalClaimableSize)),
	)
}

func headerStatusError(theme *Theme, err error) string {
	return fmt.Sprintf("[%s] Error: %v", theme.darkGray.String(), err)
}

func footerStatusMenu(theme *Theme) string {
	return fmt.Sprintf("[%s] r: Restart Scan  ↑/↓: Navigate  i: Details  d: Delete  t: Theme  q: Quit", theme.fg.String())
}

func footerStatusScanning(theme *Theme, path string) string {
	return fmt.Sprintf(" Scanning: [%s]%s", theme.purple.String(), path)
}

func (a *App) updateFinalStatus() {
	if a.scanner == nil {
		return
	}

	fileCount := a.scanner.FileCount()

	a.header.SetTextAlign(cview.AlignCenter)
	a.header.SetText(headerStatus(&a.currentTheme, int64(len(a.items)), fileCount, a.totalClaimableSize.Load(), a.scanner.ElapsedTime(), a.scanner.IsRunning()))

	a.footer.SetTextAlign(cview.AlignCenter)
	a.footer.SetText(footerStatusMenu(&a.currentTheme))
}

func (a *App) updateProgressStatus(progress *scanner.ScanResult) {
	if progress.Done {
		a.updateFinalStatus()
		return
	}

	if progress.Error != nil {
		a.header.SetText(headerStatusError(&a.currentTheme, progress.Error))
		return
	}

	theme := a.currentTheme

	a.header.SetTextAlign(cview.AlignCenter)
	a.header.SetText(headerStatus(&theme, int64(len(a.items)), progress.FileCount, a.totalClaimableSize.Load(), a.scanner.ElapsedTime(), progress.Done))

	a.lastUpdate = time.Now()

	if progress.ScannedPath != "" {
		scanPath := a.replaceHomeWithTilde(progress.ScannedPath)
		w, _ := a.app.GetScreenSize()
		w = w - 10
		if len(scanPath) > w {
			scanPath = "..." + scanPath[len(scanPath)-w:]
		}
		a.footer.SetTextAlign(cview.AlignLeft)
		a.footer.SetText(footerStatusScanning(&a.currentTheme, scanPath))
	}
}
