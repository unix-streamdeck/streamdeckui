package main

import (
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
)

type handler struct {
	hasIcon bool

	loadUI func(*editor) fyne.CanvasObject
}

var (
	handlers = map[string]*handler{
		"Url":     &handler{false, loadURLUI},
		"Page":    &handler{false, loadPageUI},
		"Counter": &handler{true, func(*editor) fyne.CanvasObject { return nil }},
		"Time":    &handler{true, func(*editor) fyne.CanvasObject { return nil }},
	}
)

func loadPageUI(e *editor) fyne.CanvasObject {
	page := widget.NewEntry()
	page.Text = strconv.FormatInt(int64(e.currentButton.key.SwitchPage), 10)

	page.OnChanged = func(text string) {
		pageNum := 0
		if text != "" {
			num, err := strconv.ParseInt(text, 10, 0)
			if err != nil {
				dialog.ShowError(err, e.win)
				return
			}
			pageNum = int(num)
		}
		e.currentButton.key.SwitchPage = pageNum
		e.currentButton.updateKey()
	}

	return widget.NewForm(
		widget.NewFormItem("Page", page),
	)
}

func loadURLUI(e *editor) fyne.CanvasObject {
	url := widget.NewEntry()
	url.Text = e.currentButton.key.Url

	url.OnChanged = func(text string) {
		e.currentButton.key.Url = text
		e.currentButton.updateKey()
	}

	return widget.NewForm(
		widget.NewFormItem("Url", url),
	)
}
