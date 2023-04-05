module github.com/unix-streamdeck/streamdeckui

go 1.16

require (
	fyne.io/fyne/v2 v2.3.0
	github.com/ncruces/zenity v0.7.12
	github.com/unix-streamdeck/api v1.0.1
)

replace github.com/unix-streamdeck/api v1.0.1 => ../api
