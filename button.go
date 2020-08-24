package main

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"github.com/unix-streamdeck/api"
)

type button struct {
	border, bg *canvas.Rectangle
	text       *canvas.Text
	icon       *canvas.Image

	key api.Key
}

func newButton(key api.Key, size int) *button {
	icon := canvas.NewImageFromFile(key.Icon)
	text := canvas.NewText(key.Text, color.White)
	text.Alignment = fyne.TextAlignCenter

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeWidth = 2
	border.StrokeColor = &color.Gray{128}
	border.SetMinSize(fyne.NewSize(size, size))

	return &button{border: border, bg: canvas.NewRectangle(color.Black),
		text: text, icon: icon, key: key}
}

func (b *button) loadUI() fyne.CanvasObject {
	return fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		b.bg, b.icon, b.text, b.border)
}

func (b *button) updateKey() {
	if len(config.Pages) == 0 {
		config.Pages = append(config.Pages, api.Page{api.Key{}})
	}
	config.Pages[0][0] = b.key
	err := conn.SetConfig(config)
	if err != nil {
		dialog.ShowError(err, win)
	}
}
