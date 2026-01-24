package tui

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"codeberg.org/tslocum/cview"
	"github.com/dustin/go-humanize"
	"github.com/riadafridishibly/npmclean/scanner"
)

func (a *App) IsScanning() bool {
	return a.scanner != nil && a.scanner.IsRunning()
}

func (a *App) startScanning() {
	a.scanner = scanner.NewScanner(a.rootPath)

	// Load cached results first
	if cachedResults, err := a.scanner.LoadCachedResults(); err == nil {
		a.handleBatchResults(cachedResults)
	}

	a.scanner.Start()

	ctx := context.Background()

	go a.processProgressEvents(ctx)
	go a.processResultEvents(ctx)
}

func (a *App) replaceHomeWithTilde(p string) string {
	if after, ok := strings.CutPrefix(p, a.userHomeDir); ok {
		p = "~" + after
	}
	return p
}

func (a *App) buildTable() *cview.Table {
	theme := a.currentTheme
	table := a.table
	table.Clear()
	items := a.items[:]
	sort.Slice(items, func(i, j int) bool { return items[i].Size > items[j].Size })
	for row, item := range items {
		// Access
		accessCell := cview.NewTableCell(" " + humanize.Time(item.LastModifiedAt))
		accessCell.SetTextColor(theme.fg)
		accessCell.SetAlign(cview.AlignLeft)

		// TODO: Set the actual object as reference
		// We can probably add the actual path here
		// Then we'll just query it, and never going to need to lookup the a.items array
		// Then we'll get the cell and take the reference pathCell.GetReference().(string)
		//  TODO: how about the GC? Are we creating multiple reference?
		//  Though: we're only sorting them, but they are basically the same pointer
		//  Only restarting the application should clear both the list and table
		accessCell.SetReference(item)
		table.SetCell(row, 0, accessCell)

		// Size
		sizeCell := cview.NewTableCell(fmt.Sprintf(" %s ", humanize.Bytes(uint64(item.Size))))
		sizeCell.SetTextColor(theme.yellow)
		sizeCell.SetAlign(cview.AlignRight)
		table.SetCell(row, 1, sizeCell)

		// Path
		pathCell := cview.NewTableCell(a.replaceHomeWithTilde(item.Path))
		pathCell.SetTextColor(theme.fg)
		pathCell.SetAlign(cview.AlignLeft)
		pathCell.SetExpansion(1)
		table.SetCell(row, 2, pathCell)
	}

	table.SetBorder(false)
	table.SetBorders(false)
	table.SetSelectable(true, false)
	table.SetSeparator(' ')

	// table.SetScrollBarVisibility(cview.ScrollBarNever)

	return table
}

func (a *App) handleBatchResults(results []*scanner.NodeModuleInfo) {
	// 1. Build a reverse index
	ri := make(map[string]int)
	for i, p := range a.items {
		ri[p.Path] = i
	}

	// 2. Iterate through the new items and check if we already have the item
	for _, result := range results {
		// We have the path, update existing
		if idx, ok := ri[result.Path]; ok {
			oldSize := a.items[idx].Size

			a.items[idx].Size = result.Size
			a.items[idx].LastModifiedAt = result.LastModifiedAt
			a.items[idx].ScannedAt = result.ScannedAt

			// Adjust claimable size
			a.totalClaimableSize.Add(-oldSize + result.Size)
			continue
		}

		a.items = append(a.items, result)
		a.totalClaimableSize.Add(result.Size)
	}

	a.trySendUIUpdate(func() { a.buildTable() })
}

func (a *App) handleResult(result *scanner.NodeModuleInfo) {
	a.handleBatchResults([]*scanner.NodeModuleInfo{result})
}

func (a *App) showItemDetail() {
	if a.table == nil {
		return
	}

	row, _ := a.table.GetSelection()
	cell := a.table.GetCell(row, 0) // always bind the reference to 0th column
	if cell == nil {
		return
	}

	ref, ok := cell.GetReference().(*scanner.NodeModuleInfo)
	if !ok {
		log.Printf("Expected *scanner.NodeModuleInfo, but found %T", cell.GetReference())
		return
	}

	// TODO: can this ever be out of bound?
	// If we have inconsistency between table and items then probably yes
	item := ref

	var detail strings.Builder
	fmt.Fprintf(&detail, "Path: %s\n", item.Path)
	fmt.Fprintf(&detail, "Size: %s\n", humanize.Bytes(uint64(item.Size)))
	fmt.Fprintf(&detail, "Last Modified: %s\n", item.LastModifiedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(&detail, "Scanned At: %s\n", item.ScannedAt.Format(time.Kitchen))

	a.detailModal.SetText(detail.String())
	a.panels.AddPanel("detailModal", a.detailModal, false, true)
}

func (a *App) confirmDelete() {
	if a.table == nil {
		return
	}
	row, _ := a.table.GetSelection()
	cell := a.table.GetCell(row, 0)
	if cell == nil {
		return
	}
	module, ok := cell.GetReference().(*scanner.NodeModuleInfo)
	if !ok {
		return
	}
	baseName := module.Path
	text := fmt.Sprintf("Delete '%s'?\n\nSize: %s", baseName, humanize.Bytes(uint64(module.Size)))
	a.confirmModal.SetText(text)
	a.panels.AddPanel("confirmModal", a.confirmModal, false, true)
}

func (a *App) deleteSelectedItem() {
	row, _ := a.table.GetSelection()
	cell := a.table.GetCell(row, 0)
	if cell == nil {
		return
	}

	module, ok := cell.GetReference().(*scanner.NodeModuleInfo)
	if !ok {
		return
	}

	a.pendingDeletes.Add(1)
	a.deleteQueue <- module
}

func (a *App) Stop() {
	if a.scanner != nil {
		a.scanner.Stop()
	}
}

func (a *App) Run() error {
	go func() {
		for updateFn := range a.uiUpdates {
			a.app.QueueUpdateDraw(updateFn)
		}
	}()
	go a.startScanning()
	go a.startDeleteWorkers(2)
	return a.app.Run()
}
