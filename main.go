package main

import (
	"log"

	"fyne.io/fyne/v2/app"
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
	w.SetContent(e.loadUI())
	w.ShowAndRun()
}
