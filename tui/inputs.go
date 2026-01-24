package tui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

func containsFooterStatus(text string) bool {
	return strings.Contains(text, "Deleted:") || strings.Contains(text, "Error deleting")
}

func (a *App) handleInput(event *tcell.EventKey) *tcell.EventKey {
	// Clear deletion status on any key press after deletion completes
	if !a.IsDeleting() && !a.showDetail && !a.showConfirm && !a.showTheme && !a.showQuit {
		footerText := a.footer.GetText(false)
		if len(footerText) > 0 {
			if containsFooterStatus(footerText) {
				a.trySendUIUpdate(a.updateFinalStatus)
			}
		}
	}

	// TODO: Fix the modal handling
	if a.showDetail || a.showConfirm || a.showTheme || a.showQuit {
		// Let modals handle their own input
		switch event.Str() {
		case "l":
			return tcell.NewEventKey(tcell.KeyRight, tcell.KeyNames[tcell.KeyRight], tcell.ModNone)
		case "h":
			return tcell.NewEventKey(tcell.KeyLeft, tcell.KeyNames[tcell.KeyLeft], tcell.ModNone)
		}

		return event
	}

	// vi key binding for modal button selection
	if a.panels != nil && a.panels.HasPanel("confirm") {
		switch event.Str() {
		case "l":
			return tcell.NewEventKey(tcell.KeyRight, tcell.KeyNames[tcell.KeyRight], tcell.ModNone)
		case "h":
			return tcell.NewEventKey(tcell.KeyLeft, tcell.KeyNames[tcell.KeyLeft], tcell.ModNone)
		}

		return event
	}

	switch event.Str() {
	case "q", "Q":
		if a.IsDeleting() {
			a.quitModal.SetText("Deletion in progress. Wait for it to complete or force quit?")
			a.showQuit = true
			a.setRoot(a.quitModal, false)
			return nil
		}
		a.Stop()
		a.app.Stop()
		return nil
	case "r", "R":
		if !a.isRestarting.Load() {
			a.isRestarting.Store(true)
			a.shouldRestart = true
			a.Stop()
			a.app.Stop()
		}
		return nil
	case "i", "I":
		a.showItemDetail()
	case "d", "D":
		a.confirmDelete()
	case "t", "T":
		a.showThemeSelector()
	}

	return event
}
