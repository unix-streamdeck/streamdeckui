package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/unix-streamdeck/api"
)

var (
	currentButton *button
	config        *api.Config

	entry, icon, url *widget.Entry
	buttons          []fyne.CanvasObject

	win fyne.Window
)

func loadEditor() fyne.CanvasObject {
	entry = widget.NewEntry()
	entry.OnChanged = func(text string) {
		currentButton.key.Text = text
		currentButton.Refresh()
		currentButton.updateKey()
	}

	icon = widget.NewEntry()
	icon.OnChanged = func(text string) {
		currentButton.key.Icon = text
		currentButton.Refresh()
		currentButton.updateKey()
	}

	url = widget.NewEntry()
	url.OnChanged = func(text string) {
		currentButton.key.Url = text
		currentButton.updateKey()
	}

	refreshEditor()
	return widget.NewForm(
		widget.NewFormItem("Text", entry),
		widget.NewFormItem("Icon", icon),
		widget.NewFormItem("Url", url),
	)
}

func editButton(b *button) {
	old := currentButton
	currentButton = b

	old.Refresh()
	b.Refresh()

	refreshEditor()
}

func refreshEditor() {
	entry.SetText(currentButton.key.Text)
	icon.SetText(currentButton.key.Icon)
	url.SetText(currentButton.key.Url)
}

func reset() {
	for _, b := range buttons {
		b.(*button).key = api.Key{}
		config.Pages[0][b.(*button).keyID] = b.(*button).key
		b.Refresh()
	}
	refreshEditor()

	err := conn.SetConfig(config)
	if err != nil {
		dialog.ShowError(err, win)
	}
}

func loadToolbar(w fyne.Window) *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			err := conn.CommitConfig()
			if err != nil {
				dialog.ShowError(err, w)
			}
		}),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			dialog.ShowConfirm("Reset config?", "Are you sure you want to reset?",
				func(ok bool) {
					if ok {
						reset()
					}
				}, w)
		}),
	)
}

func loadUI(info *api.StreamDeckInfo, w fyne.Window) fyne.CanvasObject {
	win = w
	size := int(float32(info.IconSize) / w.Canvas().Scale())

	c, err := conn.GetConfig()
	if err != nil {
		dialog.ShowError(err, w)
		c = &api.Config{}
	}
	config = c

	var page api.Page
	if len(config.Pages) >= 1 {
		page = config.Pages[0]
	}

	for i := 0; i < info.Cols*info.Rows; i++ {
		var key api.Key
		if i < len(page) {
			key = page[i]
		}
		btn := newButton(key, i, size)
		if i == 1 {
			currentButton = btn
		}
		buttons = append(buttons, btn)
	}

	toolbar := loadToolbar(w)
	editor := loadEditor()
	grid := fyne.NewContainerWithLayout(layout.NewGridLayout(info.Cols),
		buttons...)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(toolbar, editor, nil, nil),
		toolbar, editor, grid)
}
