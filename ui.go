package main

import (
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/unix-streamdeck/api"
)

type editor struct {
	currentButton      *button
	config             *api.Config
	iconSize           int
	pageCols, pageRows int
	currentPage        int

	entry, icon, url *widget.Entry
	pageLabel        *toolbarLabel
	buttons          []fyne.CanvasObject

	win fyne.Window
}

func newEditor(info *api.StreamDeckInfo, w fyne.Window) *editor {
	c, err := conn.GetConfig()
	if err != nil {
		dialog.ShowError(err, w)
		c = &api.Config{}
	}

	size := int(float32(info.IconSize) / w.Canvas().Scale())
	return &editor{config: c, iconSize: size, currentPage: info.Page,
		pageCols: info.Cols, pageRows: info.Rows, win: w}
}

func (e *editor) loadEditor() fyne.CanvasObject {
	e.entry = widget.NewEntry()
	e.entry.OnChanged = func(text string) {
		e.currentButton.key.Text = text
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}

	e.icon = widget.NewEntry()
	e.icon.OnChanged = func(text string) {
		e.currentButton.key.Icon = text
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}

	e.url = widget.NewEntry()
	e.url.OnChanged = func(text string) {
		e.currentButton.key.Url = text
		e.currentButton.updateKey()
	}

	e.refreshEditor()
	return widget.NewForm(
		widget.NewFormItem("Text", e.entry),
		widget.NewFormItem("Icon", e.icon),
		widget.NewFormItem("Url", e.url),
	)
}

func (e *editor) editButton(b *button) {
	old := e.currentButton
	e.currentButton = b

	old.Refresh()
	b.Refresh()

	e.refreshEditor()
}

func (e *editor) emptyPage() api.Page {
	var keys api.Page
	for i := 0; i < e.pageCols*e.pageRows; i++ {
		keys = append(keys, api.Key{})
	}

	return keys
}

func (e *editor) refreshEditor() {
	e.entry.SetText(e.currentButton.key.Text)
	e.icon.SetText(e.currentButton.key.Icon)
	e.url.SetText(e.currentButton.key.Url)
}

func (e *editor) refresh() {
	for _, b := range e.buttons {
		if e.currentButton == nil {
			e.currentButton = b.(*button)
		}
		if b.(*button).keyID >= len(e.config.Pages[e.currentPage]) {
			e.config.Pages[e.currentPage] = append(e.config.Pages[e.currentPage], api.Key{})
		}
		b.(*button).key = e.config.Pages[e.currentPage][b.(*button).keyID]
		b.Refresh()
	}

	e.refreshEditor()
}

func (e *editor) reset() {
	for _, b := range e.buttons {
		newKey := api.Key{}
		b.(*button).key = newKey
		e.config.Pages[e.currentPage][b.(*button).keyID] = newKey
		b.Refresh()
	}
	e.refreshEditor()

	err := conn.SetConfig(e.config)
	if err != nil {
		dialog.ShowError(err, e.win)
	}
}

func (e *editor) setPage(page int) {
	err := conn.SetPage(page)
	if err != nil {
		dialog.ShowError(err, e.win)
		return
	}

	text := fmt.Sprintf("%d/%d", page+1, len(e.config.Pages))
	e.pageLabel.label.SetText(text)
	e.currentPage = page
	e.currentButton = nil
	e.refresh()
}

func (e *editor) loadToolbar() *widget.Toolbar {
	e.pageLabel = newToolbarLabel()
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			err := conn.CommitConfig()
			if err != nil {
				dialog.ShowError(err, e.win)
			}
		}),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			dialog.ShowConfirm("Reset config?", "Are you sure you want to reset?",
				func(ok bool) {
					if ok {
						e.reset()
					}
				}, e.win)
		}),

		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MediaSkipPreviousIcon(), func() {
			if e.currentPage == 0 {
				return
			}

			e.setPage(e.currentPage - 1)
		}),
		e.pageLabel,
		widget.NewToolbarAction(theme.MediaSkipNextIcon(), func() {
			if e.currentPage == len(e.config.Pages)-1 {
				return
			}

			e.setPage(e.currentPage + 1)
		}),
		widget.NewToolbarSpacer(),

		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			if e.currentPage == len(e.config.Pages)-1 {
				e.config.Pages = append(e.config.Pages, e.emptyPage())
			} else {
				e.config.Pages = append(e.config.Pages, api.Page{}) // dummy value
				for i := len(e.config.Pages) - 1; i > e.currentPage; i-- {
					e.config.Pages[i] = e.config.Pages[i-1]
				}
				e.config.Pages[e.currentPage+1] = e.emptyPage()
			}
			err := conn.SetConfig(e.config)
			if err != nil {
				dialog.ShowError(err, e.win)
				return
			}
			e.setPage(e.currentPage + 1)
		}),
		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {

		}),
	)
}

func (e *editor) loadUI() fyne.CanvasObject {
	toolbar := e.loadToolbar()
	var page api.Page
	if len(e.config.Pages) >= 1 {
		page = e.config.Pages[e.currentPage]
	}

	for i := 0; i < e.pageCols*e.pageRows; i++ {
		var key api.Key
		if i < len(page) {
			key = page[i]
		}
		btn := newButton(key, i, e)
		if i == 0 {
			e.currentButton = btn
		}
		e.buttons = append(e.buttons, btn)
	}

	editor := e.loadEditor()
	e.setPage(e.currentPage)

	grid := fyne.NewContainerWithLayout(layout.NewGridLayout(e.pageCols),
		e.buttons...)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(toolbar, editor, nil, nil),
		toolbar, editor, grid)
}

type toolbarLabel struct {
	label *widget.Label
}

func (t *toolbarLabel) ToolbarObject() fyne.CanvasObject {
	return t.label
}

func newToolbarLabel() *toolbarLabel {
	return &toolbarLabel{label: widget.NewLabel("0")}
}
