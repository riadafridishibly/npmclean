package tui

import "github.com/gdamore/tcell/v3"

func (a *App) handleInput(event *tcell.EventKey) *tcell.EventKey {
	// TODO: Fix the modal handling
	if a.showDetail || a.showConfirm || a.showTheme {
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
	// TODO: We don't want this step, we want immediate invokation
	case "s", "S":
		if !a.IsScanning() {
			a.startScanning()
		}
		return nil
	case "q", "Q":
		a.Stop()
		a.app.Stop()
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
