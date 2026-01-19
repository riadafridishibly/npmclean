package tui

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/gdamore/tcell/v3"
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

	items         []*scanner.NodeModuleInfo
	rootPath      string // TODO: We'll support multiple roots later
	lastUpdate    time.Time
	showDetail    bool
	showConfirm   bool
	pendingDelete int

	uiUpdates chan func()

	userHomeDir        string
	totalClaimableSize atomic.Int64
}

func NewApp(scanPath string) *App {
	app := cview.NewApplication()

	header := cview.NewTextView()
	header.SetDynamicColors(true)
	header.SetTextAlign(cview.AlignLeft)

	footer := cview.NewTextView()
	footer.SetDynamicColors(true)
	footer.SetTextAlign(cview.AlignLeft)

	detailModal := cview.NewModal()
	detailModal.SetText("")
	detailModal.AddButtons([]string{"Okay"})

	confirmModal := cview.NewModal()
	confirmModal.SetText("")
	confirmModal.AddButtons([]string{"Delete", "Cancel", "Don't ask again"})

	header.SetBackgroundColor(tcell.ColorDarkBlue)
	header.SetTitleColor(tcell.ColorWhite)

	footer.SetBackgroundColor(tcell.ColorDarkGray)
	footer.SetTitleColor(tcell.ColorWhite)

	panels := cview.NewPanels()
	table := cview.NewTable()
	panels.AddPanel("table", table, true, true)

	a := &App{
		app:          app,
		header:       header,
		footer:       footer,
		detailModal:  detailModal,
		confirmModal: confirmModal,
		rootPath:     scanPath,
		panels:       panels,
		table:        table,
		items:        make([]*scanner.NodeModuleInfo, 0),
		showDetail:   false,
		showConfirm:  false,
		uiUpdates:    make(chan func(), 128),
	}

	flex := cview.NewFlex()
	flex.SetDirection(cview.FlexRow)
	flex.AddItem(header, 1, 0, false)
	flex.AddItem(panels, 0, 1, true)
	flex.AddItem(footer, 1, 0, false)

	app.SetInputCapture(a.handleInput)

	detailModal.SetDoneFunc(func(_ int, _ string) {
		a.showDetail = false
		app.SetRoot(flex, true)
	})

	confirmModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.showConfirm = false
		app.SetRoot(flex, true)

		switch buttonLabel {
		case "Delete":
			a.deleteItem(a.pendingDelete)
		case "Don't ask":
			// TOOD: remember not to ask again
			a.deleteItem(a.pendingDelete)
		}
	})

	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicln("Error getting home:", err)
	}

	a.userHomeDir = home

	// TODO: Remove the first confirmation step
	header.SetText(fmt.Sprintf("[white]Ready to scan. Press 's' to start scanning: %s", scanPath))
	footer.SetText("[white]s: Start scan  ↑/↓: Navigate  i: Details  d: Delete  q: Quit")

	app.SetRoot(flex, true)
	return a
}
