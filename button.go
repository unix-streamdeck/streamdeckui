package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/unix-streamdeck/api"
	"image"
	"image/color"
)

type button struct {
	widget.BaseWidget
	editor *editor

	keyID int
	key   api.KeyV3
	currentConfig *api.KeyConfigV3
}

func newButton(key api.KeyV3, id int, e *editor) *button {
	currentConfig, _ := key.Application[""]
	b := &button{key: key, keyID: id, editor: e, currentConfig: currentConfig}
	b.ExtendBaseWidget(b)
	return b
}

func (b *button) CreateRenderer() fyne.WidgetRenderer {
	var icon *canvas.Image
	if b.currentConfig == nil {
		b.updateApplication(b.editor.selectedApplication)
	}
	if b.currentConfig.IconHandler == "" {
		icon = canvas.NewImageFromFile(b.currentConfig.Icon)
	} else {
		img, err := conn.GetHandlerExample(b.editor.currentDevice.Serial, *b.currentConfig)
		if err != nil {
			dialog.ShowError(err, b.editor.win)
			//js, _ := json.Marshal(b.currentConfig)
		}
		icon = canvas.NewImageFromImage(img)
	}
	text := &canvas.Image{}

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeWidth = 2
	border.SetMinSize(fyne.NewSize(float32(b.editor.currentDevice.IconSize), float32(b.editor.currentDevice.IconSize)))

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
	if b.keyID >= len(b.editor.currentDeviceConfig.Pages[b.editor.currentDevice.Page].Keys) {
		return
	}
	if b.currentConfig.IconHandler == "Default" {
		b.currentConfig.IconHandler = ""
	}
	if b.currentConfig.KeyHandler == "Default" {
		b.currentConfig.KeyHandler = ""
	}
	b.editor.currentDeviceConfig.Pages[b.editor.currentDevice.Page].Keys[b.keyID] = b.key
}

func (b *button) updateApplication(app string) {
	currentConfig, ok := b.key.Application[app]
	if !ok {
		currentConfig = &api.KeyConfigV3{}
		if b.key.Application == nil {
			b.key.Application = map[string]*api.KeyConfigV3{}
		}
		b.key.Application[app] = currentConfig
	}
	b.currentConfig = currentConfig
	b.Refresh()
}

const (
	buttonInset = 2
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
	iconSize := fyne.NewSize(float32(r.b.editor.currentDevice.IconSize), float32(r.b.editor.currentDevice.IconSize))
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
	if r.b.currentConfig == nil {
		r.b.currentConfig = &api.KeyConfigV3{}
		if r.b.key.Application == nil {
			r.b.key.Application = map[string]*api.KeyConfigV3{}
		}
		r.b.key.Application[r.b.editor.selectedApplication] = r.b.currentConfig
	}
	if r.b.currentConfig.IconHandler == "" {
		r.icon.Image = nil
		r.icon.File = r.b.currentConfig.Icon
		go r.icon.Refresh()
	} else {
		img, err := conn.GetHandlerExample(r.b.editor.currentDevice.Serial, *r.b.currentConfig)
		if err != nil {
			dialog.ShowError(err, r.b.editor.win)
			//js, _ := json.Marshal(r.b.currentConfig)
		}
		r.icon.File = ""
		r.icon.Image = img
		go r.icon.Refresh()
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
	if r.b.currentConfig == nil {
		return nil
	}
	img, err := api.DrawText(textImg, r.b.currentConfig.Text, r.b.currentConfig.TextSize, r.b.currentConfig.TextAlignment)
	if err != nil {
		fyne.LogError("Failed to draw text to imge", err)
	}
	return img
}
