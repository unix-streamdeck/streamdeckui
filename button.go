package main

import (
	"image"
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/unix-streamdeck/api"
)

type button struct {
	widget.BaseWidget
	editor *editor

	keyID int
	key   api.Key
}

func newButton(key api.Key, id int, e *editor) *button {
	b := &button{key: key, keyID: id, editor: e}
	b.ExtendBaseWidget(b)
	return b
}

func (b *button) CreateRenderer() fyne.WidgetRenderer {
	icon := canvas.NewImageFromFile(b.key.Icon)
	text := &canvas.Image{}

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeWidth = 2
	border.SetMinSize(fyne.NewSize(b.editor.currentDevice.IconSize, b.editor.currentDevice.IconSize))

	bg := canvas.NewRectangle(color.Black)
	render := &buttonRenderer{border: border, text: text, icon: icon, bg: bg,
		objects: []fyne.CanvasObject{bg, icon, text, border}, b: b}
	render.Refresh()
	return render
}

func (b *button) Tapped(ev *fyne.PointEvent) {
	b.editor.editButton(b)
}

func (b *button) updateKey() {
	if b.keyID >= len(b.editor.currentDeviceConfig.Pages[b.editor.currentDevice.Page]) {
		return
	}
	b.editor.currentDeviceConfig.Pages[b.editor.currentDevice.Page][b.keyID] = b.key
	if b.editor.currentDeviceConfig.Pages[b.editor.currentDevice.Page][b.keyID].IconHandler == "Default" {
		b.editor.currentDeviceConfig.Pages[b.editor.currentDevice.Page][b.keyID].IconHandler = ""
	}
	if b.editor.currentDeviceConfig.Pages[b.editor.currentDevice.Page][b.keyID].KeyHandler == "Default" {
		b.editor.currentDeviceConfig.Pages[b.editor.currentDevice.Page][b.keyID].KeyHandler = ""
	}
}

const (
	buttonInset = 5
)

type buttonRenderer struct {
	border, bg *canvas.Rectangle
	icon, text *canvas.Image

	objects []fyne.CanvasObject

	b *button
}

func (r *buttonRenderer) Layout(s fyne.Size) {
	size := s.Subtract(fyne.NewSize(buttonInset*2, buttonInset*2))
	offset := fyne.NewPos(buttonInset, buttonInset)

	for _, obj := range r.objects {
		obj.Move(offset)
		obj.Resize(size)
	}
}

func (r *buttonRenderer) MinSize() fyne.Size {
	iconSize := fyne.NewSize(r.b.editor.currentDevice.IconSize, r.b.editor.currentDevice.IconSize)
	return iconSize.Add(fyne.NewSize(buttonInset*2, buttonInset*2))
}

func (r *buttonRenderer) Refresh() {
	if r.b.editor.currentButton == r.b {
		r.border.StrokeColor = theme.FocusColor()
	} else {
		r.border.StrokeColor = &color.Gray{128}
	}

	r.text.Image = r.textToImage()
	r.text.Refresh()
	if r.b.key.Icon != r.icon.File {
		r.icon.File = r.b.key.Icon
		r.icon.Refresh()
	}

	r.border.Refresh()
}

func (r *buttonRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *buttonRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *buttonRenderer) Destroy() {
	// nothing
}

func (r *buttonRenderer) textToImage() image.Image {
	textImg := image.NewNRGBA(image.Rect(0, 0, r.b.editor.currentDevice.IconSize, r.b.editor.currentDevice.IconSize))
	img, err := api.DrawText(textImg, r.b.key.Text, r.b.key.TextSize, r.b.key.TextAlignment)
	if err != nil {
		fyne.LogError("Failed to draw text to imge", err)
	}
	return img
}
