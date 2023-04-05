package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
	"github.com/unix-streamdeck/api"
	"strconv"
	"strings"
)

var (
	handlers = []*api.Module{
		{Name: "Default", IsIcon: true, IsKey: true},
	}
)

func initHandlers(conn *api.Connection) {
	modules, err := conn.GetModules()
	if err != nil {
		fyne.LogError("Unable to get handlers", err)
	}
	handlers = append(handlers, modules...)
}

func loadDefaultIconUI(e *editor) fyne.CanvasObject {

	entry := widget.NewMultiLineEntry()
	entry.OnChanged = func(text string) {
		e.currentButton.currentConfig.Text = text
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}

	icon := widget.NewButton("Select Icon", func() {
		file, err := zenity.SelectFile(zenity.FileFilters{{"Files", []string{"*.png", "*.jpg", "*.jpeg"}}})
		if err != nil && err.Error() != "dialog canceled" {
			dialog.ShowError(err, e.win)
			return
		}
		if file != "" {
			e.currentButton.currentConfig.Icon = file
			e.currentButton.Refresh()
			e.currentButton.updateKey()
		}
	})

	clearIcon := widget.NewButton("Clear Icon", func() {
		e.currentButton.currentConfig.Icon = ""
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	})
	iconGroup := fyne.NewContainerWithLayout(layout.NewGridLayout(2), icon, clearIcon)
	//iconGroup := widget.NewForm(widget.NewFormItem("", icon), widget.NewFormItem("", clearIcon))

	textAlignment := widget.NewSelect([]string{"TOP", "MIDDLE", "BOTTOM"}, func(alignment string) {
		e.currentButton.currentConfig.TextAlignment = alignment
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	})

	textSize := widget.NewEntry()
	textSize.OnChanged = func(size string) {
		if size == "" {
			e.currentButton.currentConfig.TextSize = 0
			e.currentButton.Refresh()
			e.currentButton.updateKey()
			return
		}
		sizeInt, err := strconv.Atoi(size)
		if err != nil {
			dialog.ShowError(err, e.win)
			return
		}
		e.currentButton.currentConfig.TextSize = sizeInt
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}

	entry.SetText(e.currentButton.currentConfig.Text)
	if e.currentButton.currentConfig.TextSize != 0 {
		textSize.SetText(strconv.Itoa(e.currentButton.currentConfig.TextSize))
	} else {
		textSize.SetText("")
	}
	textAlignment.SetSelected(strings.ToUpper(e.currentButton.currentConfig.TextAlignment))

	return widget.NewForm(
		widget.NewFormItem("Text", entry),
		widget.NewFormItem("Text Alignment", textAlignment),
		widget.NewFormItem("Font Size", textSize),
		widget.NewFormItem("Icon", iconGroup),
	)
}

func loadDefaultKeyUI(e *editor) fyne.CanvasObject {

	url := widget.NewEntry()
	url.Text = e.currentButton.currentConfig.Url
	page := widget.NewEntry()
	page.Text = strconv.FormatInt(int64(e.currentButton.currentConfig.SwitchPage), 10)
	keyBind := widget.NewEntry()
	keyBind.Text = e.currentButton.currentConfig.Keybind
	command := widget.NewEntry()
	command.Text = e.currentButton.currentConfig.Command
	brightness := widget.NewEntry()
	brightness.Text = strconv.FormatInt(int64(e.currentButton.currentConfig.Brightness), 10)

	url.OnChanged = func(text string) {
		e.currentButton.currentConfig.Url = text
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}

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
		e.currentButton.currentConfig.SwitchPage = pageNum
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}

	keyBind.OnChanged = func(text string) {
		e.currentButton.currentConfig.Keybind = text
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}

	command.OnChanged = func(text string) {
		e.currentButton.currentConfig.Command = text
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}

	brightness.OnChanged = func(text string) {
		brightness := 0
		if text != "" {
			num, err := strconv.ParseInt(text, 10, 0)
			if err != nil {
				dialog.ShowError(err, e.win)
				return
			}
			if int(num) > 100 || int(num) < 0 {
				dialog.ShowError(errors.New("Brightness out of range"), e.win)
				return
			}
			brightness = int(num)
		}
		e.currentButton.currentConfig.Brightness = brightness
		e.currentButton.Refresh()
		e.currentButton.updateKey()
	}
	return widget.NewForm(
		widget.NewFormItem("URL", url),
		widget.NewFormItem("Switch Page", page),
		widget.NewFormItem("Keybind", keyBind),
		widget.NewFormItem("Command", command),
		widget.NewFormItem("Brightness", brightness),
	)
}

func loadUI(fields []api.Field, itemMap map[string]string, e *editor) fyne.CanvasObject {
	var items []*widget.FormItem
	for _, field := range fields {
		item := generateField(field, itemMap, e)
		if item != nil {
			items = append(items, item)
		}
	}
	return widget.NewForm(items...)
}

func generateField(field api.Field, itemMap map[string]string, e *editor) *widget.FormItem {
	if field.Type == "Text" {
		item := widget.NewEntry()
		item.Text = itemMap[field.Name]
		item.OnChanged = func(text string) {
			itemMap[field.Name] = text
			e.currentButton.Refresh()
			e.currentButton.updateKey()
		}
		return widget.NewFormItem(field.Title, item)
	} else if field.Type == "File" {
		file := widget.NewButton("Select File", func() {
			var fileTypes []string
			for _, fileType := range field.FileTypes {
				fileTypes = append(fileTypes, "*"+fileType)
			}
			file, err := zenity.SelectFile(zenity.FileFilters{{"Files", fileTypes}})
			if err != nil && err.Error() != "dialog canceled" {
				dialog.ShowError(err, e.win)
				return
			}
			if file != "" {
				itemMap[field.Name] = file
				e.currentButton.Refresh()
				e.currentButton.updateKey()
			}
		})
		clearFile := widget.NewButton("Clear File", func() {
			itemMap[field.Name] = ""
			e.currentButton.Refresh()
			e.currentButton.updateKey()
		})
		item := fyne.NewContainerWithLayout(layout.NewGridLayout(2), file, clearFile)
		return widget.NewFormItem(field.Title, item)
	} else if field.Type == "TextAlignment" {
		item := widget.NewSelect([]string{"TOP", "MIDDLE", "BOTTOM"}, func(alignment string) {
			itemMap[field.Name] = alignment
			e.currentButton.Refresh()
			e.currentButton.updateKey()
		})
		alignment, ok := itemMap[field.Name]
		if ok {
			item.SetSelected(strings.ToUpper(alignment))
		}
		return widget.NewFormItem(field.Title, item)
	} else if field.Type == "Number" {
		item := widget.NewEntry()
		item.Text = itemMap[field.Name]
		item.OnChanged = func(text string) {
			value := 0
			if text != "" {
				num, err := strconv.ParseInt(text, 10, 0)
				if err != nil {
					dialog.ShowError(err, e.win)
					return
				}
				value = int(num)
			}
			itemMap[field.Name] = strconv.Itoa(value)
			e.currentButton.Refresh()
			e.currentButton.updateKey()
		}
		return widget.NewFormItem(field.Title, item)
	//} else if field.Type == "Select" {
	//	item := widget.NewSelect(field.Values, func(value string) {
	//		itemMap[field.Name] = value
	//		e.currentButton.Refresh()
	//		e.currentButton.updateKey()
	//	})
	//	action, ok := itemMap[field.Name]
	//	if ok {
	//		item.SetSelected(action)
	//	}
	//	return widget.NewFormItem(field.Title, item)
	}
	return nil
}
