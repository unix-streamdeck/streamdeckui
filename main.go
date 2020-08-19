package main

import (
	"log"

	"fyne.io/fyne/app"
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

	log.Println("Device connected, size is", info.Cols, info.Rows)

	a := app.New()
	w := a.NewWindow("StreamDeck Unix")

	w.SetContent(loadUI(info, w))
	w.ShowAndRun()
}
