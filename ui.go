package main

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/unix-streamdeck/api"
)

var (
	currentButton fyne.CanvasObject
	currentKey    api.Key
	config        *api.Config

	win fyne.Window
)

func newButton(key api.Key, size int) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.Black)

	icon := canvas.NewImageFromFile(key.Icon)
	text := canvas.NewText(key.Text, color.White)
	text.Alignment = fyne.TextAlignCenter

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeWidth = 2
	border.StrokeColor = &color.Gray{128}
	border.SetMinSize(fyne.NewSize(size, size))

	return fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		bg, icon, text, border)
}

func loadEditor() fyne.CanvasObject {
	entry := widget.NewEntry()
	entry.SetText(currentKey.Text)

	entry.OnChanged = func(text string) {
		label := currentButton.(*fyne.Container).Objects[2].(*canvas.Text)
		label.Text = text
		label.Refresh()

		currentKey.Text = text
		setKey(currentKey)
	}

	icon := widget.NewEntry()
	icon.SetText(currentKey.Icon)

	icon.OnChanged = func(text string) {
		img := currentButton.(*fyne.Container).Objects[1].(*canvas.Image)
		img.File = text
		img.Refresh()

		currentKey.Icon = text
		setKey(currentKey)
	}

	url := widget.NewEntry()
	url.SetText(currentKey.Url)

	url.OnChanged = func(text string) {
		currentKey.Url = text
		setKey(currentKey)
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

func setKey(key api.Key) {
	if len(config.Pages) == 0 {
		config.Pages = append(config.Pages, api.Page{api.Key{}})
	}
	config.Pages[0][0] = key
	err := conn.SetConfig(config)
	if err != nil {
		dialog.ShowError(err, win)
	}
}

func loadUI(info *api.StreamDeckInfo, w fyne.Window) fyne.CanvasObject {
	win = w
	var buttons []fyne.CanvasObject
	size := int(float32(info.IconSize) / w.Canvas().Scale())

	c, err := conn.GetConfig()
	if err != nil {
		dialog.ShowError(err, w)
	} else {
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
		buttons = append(buttons, btn)

		if i == 0 {
			currentKey = key
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
