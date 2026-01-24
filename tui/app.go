package tui

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sync/atomic"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/riadafridishibly/npmclean/scanner"
)

type App struct {
	app     *cview.Application
	scanner *scanner.Scanner

	header       *cview.TextView
	footer       *cview.TextView
	table        *cview.Table
	panels       *cview.Panels
	detailModal  *cview.Modal
	confirmModal *cview.Modal
	themeModal   *cview.Modal
	quitModal    *cview.Modal

	items       []*scanner.NodeModuleInfo
	rootPath    string
	lastUpdate  time.Time
	showDetail  bool
	showConfirm bool
	showTheme   bool
	showQuit    bool

	uiUpdates chan func()

	userHomeDir        string
	totalClaimableSize atomic.Int64

	currentTheme  Theme
	shouldRestart bool
	isRestarting  atomic.Bool

	deleteQueue    chan *scanner.NodeModuleInfo
	deleteDone     chan *deleteResult
	activeDeletes  atomic.Int64
	pendingDeletes atomic.Int64
}

type deleteResult struct {
	module  *scanner.NodeModuleInfo
	path    string
	err     error
	success bool
}

func defaultTheme() Theme {
	return themes["nord"]
}

func (a *App) switchTheme(themeName string) {
	if th, ok := themes[themeName]; ok {
		a.currentTheme = th
	}
}

func (a *App) applyTheme() {
	theme := a.currentTheme

	a.header.SetBackgroundColor(theme.headerBg)
	a.header.SetTitleColor(theme.headerFg)
	a.header.SetTextColor(theme.headerFg)

	a.footer.SetBackgroundColor(theme.footerBg)
	a.footer.SetTitleColor(theme.footerFg)
	a.footer.SetTextColor(theme.footerFg)

	a.detailModal.SetBackgroundColor(theme.modalBg)
	a.detailModal.SetTextColor(theme.modalFg)
	a.detailModal.SetButtonBackgroundColor(theme.buttonBg)
	a.detailModal.SetButtonTextColor(theme.buttonFg)

	a.confirmModal.SetBackgroundColor(theme.modalBg)
	a.confirmModal.SetTextColor(theme.modalFg)
	a.confirmModal.SetButtonBackgroundColor(theme.buttonBg)
	a.confirmModal.SetButtonTextColor(theme.buttonFg)

	a.themeModal.SetBackgroundColor(theme.modalBg)
	a.themeModal.SetTextColor(theme.modalFg)
	a.themeModal.SetButtonBackgroundColor(theme.buttonBg)
	a.themeModal.SetButtonTextColor(theme.buttonFg)

	a.quitModal.SetBackgroundColor(theme.modalBg)
	a.quitModal.SetTextColor(theme.modalFg)
	a.quitModal.SetButtonBackgroundColor(theme.buttonBg)
	a.quitModal.SetButtonTextColor(theme.buttonFg)

	a.table.SetBackgroundColor(theme.bg)

	a.panels.SetBackgroundColor(theme.bg)

	a.trySendUIUpdate(func() {
		a.header.SetText(headerStartupStatus(&theme, a.rootPath))
		a.footer.SetText(footerStatusMenu(&theme))
		a.updateFinalStatus()
		a.buildTable()
	})
}

func NewApp(scanPath string) *App {
	app := cview.NewApplication()

	theme := defaultTheme()

	header := cview.NewTextView()
	header.SetDynamicColors(true)

	footer := cview.NewTextView()
	footer.SetDynamicColors(true)

	detailModal := cview.NewModal()
	detailModal.SetText("")
	detailModal.AddButtons([]string{"Okay"})
	detailModal.SetBackgroundColor(theme.bg)
	detailModal.SetTextColor(theme.fg)
	detailModal.SetButtonBackgroundColor(theme.orange)
	detailModal.SetButtonTextColor(theme.darkGray)

	confirmModal := cview.NewModal()
	confirmModal.SetText("")
	confirmModal.AddButtons([]string{"Delete", "Cancel", "Don't ask again"})

	themeModal := cview.NewModal()
	themeModal.SetText("")
	themeNames := getThemeNames()
	themeModal.AddButtons(themeNames)

	quitModal := cview.NewModal()
	quitModal.SetText("")
	quitModal.AddButtons([]string{"Wait", "Force Quit"})

	panels := cview.NewPanels()
	table := cview.NewTable()
	panels.AddPanel("table", table, true, true)

	a := &App{
		app:           app,
		header:        header,
		footer:        footer,
		detailModal:   detailModal,
		confirmModal:  confirmModal,
		themeModal:    themeModal,
		quitModal:     quitModal,
		rootPath:      scanPath,
		panels:        panels,
		table:         table,
		items:         make([]*scanner.NodeModuleInfo, 0),
		showDetail:    false,
		showConfirm:   false,
		showTheme:     false,
		showQuit:      false,
		uiUpdates:     make(chan func(), 128),
		currentTheme:  theme,
		shouldRestart: false,
		deleteQueue:   make(chan *scanner.NodeModuleInfo, 100),
		deleteDone:    make(chan *deleteResult, 100),
	}

	flex := cview.NewFlex()
	flex.SetDirection(cview.FlexRow)
	flex.AddItem(header, 1, 0, false)
	flex.AddItem(panels, 0, 1, true)
	flex.AddItem(footer, 1, 0, false)

	app.SetInputCapture(a.handleInput)

	detailModal.SetDoneFunc(func(_ int, _ string) {
		a.showDetail = false
		a.setRoot(flex, true)
	})

	confirmModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.showConfirm = false
		a.setRoot(flex, true)

		switch buttonLabel {
		case "Delete":
			a.deleteSelectedItem()
		case "Don't ask":
			// TOOD: remember not to ask again
			a.deleteSelectedItem()
		}
	})

	themeModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.showTheme = false
		a.setRoot(flex, true)

		if buttonIndex >= 0 && buttonIndex < len(themeNames) {
			a.switchTheme(buttonLabel)
			a.applyTheme()
		}
	})

	quitModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.showQuit = false
		a.setRoot(flex, true)

		if buttonLabel == "Force Quit" {
			a.Stop()
			a.app.Stop()
		}
	})

	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicln("Error getting home:", err)
	}

	a.userHomeDir = home

	header.SetTextAlign(cview.AlignCenter)
	header.SetText(headerStartupStatus(&theme, a.rootPath))
	footer.SetTextAlign(cview.AlignCenter)
	footer.SetText(footerStatusMenu(&theme))

	a.setRoot(flex, true)

	a.applyTheme()

	return a
}

func (a *App) showThemeSelector() {
	if a.themeModal == nil {
		return
	}
	theme := a.currentTheme
	text := fmt.Sprintf("Select Theme (Current: [%s]%s[-])", theme.orange.String(), theme.Name)
	a.themeModal.SetText(text)
	a.showTheme = true
	a.setRoot(a.themeModal, false)
}

func (a *App) ShouldRestart() bool {
	return a.shouldRestart
}

func (a *App) Scanner() *scanner.Scanner {
	return a.scanner
}

func (a *App) IsDeleting() bool {
	return a.activeDeletes.Load() > 0 || a.pendingDeletes.Load() > 0
}

func (a *App) startDeleteWorkers(workers int) {
	for i := 0; i < workers; i++ {
		go a.deleteWorker()
	}
	go a.processDeleteResults()
}

func (a *App) deleteWorker() {
	for module := range a.deleteQueue {
		a.activeDeletes.Add(1)

		displayPath := a.replaceHomeWithTilde(module.Path)
		a.trySendUIUpdate(func() { a.footer.SetText(footerStatusDeleting(&a.currentTheme, displayPath)) })

		err := os.RemoveAll(module.Path)

		a.activeDeletes.Add(-1)

		result := &deleteResult{
			module:  module,
			path:    module.Path,
			err:     err,
			success: err == nil,
		}

		a.deleteDone <- result
	}
}

func (a *App) processDeleteResults() {
	for result := range a.deleteDone {
		a.pendingDeletes.Add(-1)
		displayPath := a.replaceHomeWithTilde(result.path)

		if result.err != nil {
			log.Printf("Error deleting dir: %s: error: %v", result.path, result.err)
			a.trySendUIUpdate(func() { a.footer.SetText(footerStatusDeleteError(&a.currentTheme, displayPath, result.err)) })
		} else {
			if a.scanner != nil && a.scanner.Cache() != nil {
				a.scanner.Cache().Delete(result.path)
			}

			a.items = slices.DeleteFunc(a.items, func(mod *scanner.NodeModuleInfo) bool { return mod.Path == result.module.Path })
			a.totalClaimableSize.Add(-result.module.Size)
			a.trySendUIUpdate(func() {
				a.buildTable()
				a.updateFinalStatus()
				a.footer.SetText(footerStatusDeleted(&a.currentTheme, displayPath))
			})
		}
	}
}
