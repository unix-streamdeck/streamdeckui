package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/unix-streamdeck/api"
)

var conn *api.Connection

func main() {
	dev, err := api.Connect()
	if err != nil {
		log.Fatal("Could not connect to device: " + err.Error())
	}
	conn = dev

	defer dev.Close()
	info, err := dev.GetInfo()
	if err != nil {
		log.Fatal("Cound not read device info: " + err.Error())
	}

	a := app.New()
	w := a.NewWindow("StreamDeck Unix")

	e := newEditor(info, w)

	// CTRL-S : save config
	ctrlS := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	e.win.Canvas().AddShortcut(&ctrlS, func(shortcut fyne.Shortcut) {
		e.saveConfig()
	})

	// CTRL-C : copy current button
	e.win.Canvas().AddShortcut(&fyne.ShortcutCopy{}, func(shortcut fyne.Shortcut) {
		e.copyButton()
	})

	// CTRL-V : paste current button
	e.win.Canvas().AddShortcut(&fyne.ShortcutPaste{}, func(shortcut fyne.Shortcut) {
		e.pasteButton()
	})

	w.SetContent(e.loadUI())
	w.ShowAndRun()
}
