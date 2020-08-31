package main

import (
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

	entry, icon, url *widget.Entry
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
	return &editor{config: c, iconSize: size,
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

	e.refresh()
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

	e.refresh()
}

func (e *editor) refresh() {
	e.entry.SetText(e.currentButton.key.Text)
	e.icon.SetText(e.currentButton.key.Icon)
	e.url.SetText(e.currentButton.key.Url)
}

func (e *editor) reset() {
	for _, b := range e.buttons {
		newKey := api.Key{}
		b.(*button).key = newKey
		e.config.Pages[0][b.(*button).keyID] = newKey
		b.Refresh()
	}
	e.refresh()

	err := conn.SetConfig(e.config)
	if err != nil {
		dialog.ShowError(err, e.win)
	}
}

func (e *editor) loadToolbar() *widget.Toolbar {
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
	)
}

func (e *editor) loadUI() fyne.CanvasObject {
	var page api.Page
	if len(e.config.Pages) >= 1 {
		page = e.config.Pages[0]
	}

	for i := 0; i < e.pageCols*e.pageRows; i++ {
		var key api.Key
		if i < len(page) {
			key = page[i]
		}
		btn := newButton(key, i, e)
		if i == 1 {
			e.currentButton = btn
		}
		e.buttons = append(e.buttons, btn)
	}

	toolbar := e.loadToolbar()
	editor := e.loadEditor()
	grid := fyne.NewContainerWithLayout(layout.NewGridLayout(e.pageCols),
		e.buttons...)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(toolbar, editor, nil, nil),
		toolbar, editor, grid)
}
