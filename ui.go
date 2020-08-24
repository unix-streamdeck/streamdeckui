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

	win fyne.Window
)

func loadEditor() fyne.CanvasObject {
	entry := widget.NewEntry()
	entry.SetText(currentButton.key.Text)

	entry.OnChanged = func(text string) {
		currentButton.text.Text = text
		currentButton.text.Refresh()

		currentButton.key.Text = text
		currentButton.updateKey()
	}

	icon := widget.NewEntry()
	icon.SetText(currentButton.key.Icon)

	icon.OnChanged = func(text string) {
		currentButton.icon.File = text
		currentButton.icon.Refresh()

		currentButton.key.Icon = text
		currentButton.updateKey()
	}

	url := widget.NewEntry()
	url.SetText(currentButton.key.Url)

	url.OnChanged = func(text string) {
		currentButton.key.Url = text
		currentButton.updateKey()
	}

	return widget.NewForm(
		widget.NewFormItem("Text", entry),
		widget.NewFormItem("Icon", icon),
		widget.NewFormItem("Url", url),
	)
}

func loadToolbar(w fyne.Window) *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			err := conn.CommitConfig()
			if err != nil {
				dialog.ShowError(err, w)
			}
		}),
	)
}

func loadUI(info *api.StreamDeckInfo, w fyne.Window) fyne.CanvasObject {
	win = w
	var buttons []fyne.CanvasObject
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
		btn := newButton(key, size)
		buttons = append(buttons, btn.loadUI())

		if i == 0 {
			currentButton = btn
		}
	}

	toolbar := loadToolbar(w)
	editor := loadEditor()
	grid := fyne.NewContainerWithLayout(layout.NewGridLayout(info.Cols),
		buttons...)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(toolbar, editor, nil, nil),
		toolbar, editor, grid)
}
