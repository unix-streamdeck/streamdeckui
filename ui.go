package main

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/unix-streamdeck/api"
)

type editor struct {
	currentButton      *button
	config             *api.Config
	iconSize           int
	pageCols, pageRows int
	currentPage        int

	iconHandler, keyHandler               *widget.Select
	pageLabel                             *toolbarLabel
	buttons                               []fyne.CanvasObject
	keyDetailSelector, iconDetailSelector *fyne.Container

	win fyne.Window
}

func newEditor(info *api.StreamDeckInfo, w fyne.Window) *editor {
	c, err := conn.GetConfig()
	if err != nil {
		dialog.ShowError(err, w)
		c = &api.Config{}
	}

	size := int(float32(info.IconSize) / w.Canvas().Scale())
	ed := &editor{config: c, iconSize: size, currentPage: info.Page,
		pageCols: info.Cols, pageRows: info.Rows, win: w}
	go ed.registerPageListener() // TODO remove "go" once daemon fixed
	return ed
}

func (e *editor) loadEditor() fyne.CanvasObject {

	var keyIds []string
	var iconIds []string

	initHandlers(conn)
	for _, module := range handlers {
		if module.IsKey {
			keyIds = append(keyIds, module.Name)
		}
		if module.IsIcon {
			iconIds = append(iconIds, module.Name)
		}
	}
	e.keyDetailSelector = fyne.NewContainerWithLayout(layout.NewMaxLayout())
	e.iconDetailSelector = fyne.NewContainerWithLayout(layout.NewMaxLayout())
	e.iconHandler = widget.NewSelect(iconIds, e.chooseIconHandler)
	e.keyHandler = widget.NewSelect(keyIds, e.chooseKeyHandler)
	e.refreshEditor()

	iconHandler := widget.NewForm(
		widget.NewFormItem("Icon Handler", e.iconHandler),
	)
	keyHandler := widget.NewForm(
		widget.NewFormItem("Key Handler", e.keyHandler),
	)
	return fyne.NewContainerWithLayout(layout.NewVBoxLayout(), iconHandler, e.iconDetailSelector, keyHandler, e.keyDetailSelector)
}

func (e *editor) chooseKeyHandler(name string) {
	e.chooseHandler(name, "Key")
}

func (e *editor) chooseIconHandler(name string) {
	e.chooseHandler(name, "Icon")
}

func (e *editor) chooseHandler(name string, handlerType string) {
	var module *api.Module
	for _, mod := range handlers {
		if mod.Name == name {
			module = mod
			break
		}
	}
	if module == nil {
		fyne.LogError("Handler not found "+name, nil)
		return
	}
	if (handlerType == "Key" && !module.IsKey) || (handlerType == "Icon" && !module.IsIcon) {
		fyne.LogError("Handler not found "+name, nil)
		return
	}
	var ui fyne.CanvasObject


	var fields []api.Field
	var itemMap map[string]string
	if handlerType == "Key" {
		fields = module.KeyFields
		if e.currentButton.key.KeyHandlerFields == nil {
			e.currentButton.key.KeyHandlerFields = make(map[string]string)
		}
		itemMap = e.currentButton.key.KeyHandlerFields

	} else {
		fields = module.IconFields
		if e.currentButton.key.IconHandlerFields == nil {
			e.currentButton.key.IconHandlerFields = make(map[string]string)
		}
		itemMap = e.currentButton.key.IconHandlerFields
	}

	if fields != nil {
		ui = loadUI(fields, itemMap, e)
	} else {
		ui = widget.NewForm()
	}

	if name == "Default" {
		if handlerType == "Key" {
			ui = loadDefaultKeyUI(e)
		} else {
			ui = loadDefaultIconUI(e)
		}
	} else {
		if handlerType == "Key" {
			e.currentButton.key.KeyHandler = name
		} else {
			e.currentButton.key.IconHandler = name
		}
		e.currentButton.updateKey()
		e.currentButton.Refresh()
	}

	if ui != nil {
		if handlerType == "Key" {
			e.keyDetailSelector.Objects = []fyne.CanvasObject{ui}
		} else {
			e.iconDetailSelector.Objects = []fyne.CanvasObject{ui}
		}
	}
	if handlerType == "Key" {
		e.keyDetailSelector.Refresh()
	} else {
		e.iconDetailSelector.Refresh()
	}
}

func (e *editor) editButton(b *button) {
	old := e.currentButton
	e.currentButton = b

	old.Refresh()
	b.Refresh()

	e.refreshEditor()
}

func (e *editor) emptyPage() api.Page {
	var keys api.Page
	for i := 0; i < e.pageCols*e.pageRows; i++ {
		keys = append(keys, api.Key{})
	}

	return keys
}

func (e *editor) pageListener(page int32) {
	if int(page) == e.currentPage {
		return
	}

	e.setPage(int(page))
}

func (e *editor) refreshEditor() {

	handler := e.currentButton.key.KeyHandler
	if handler == "" {
		handler = "Default"
	}
	e.keyHandler.SetSelected(handler)
	handler = e.currentButton.key.IconHandler
	if handler == "" {
		handler = "Default"
	}
	e.iconHandler.SetSelected(handler)
}

func (e *editor) refresh() {
	for _, b := range e.buttons {
		if e.currentButton == nil {
			e.currentButton = b.(*button)
		}
		if b.(*button).keyID >= len(e.config.Pages[e.currentPage]) {
			e.config.Pages[e.currentPage] = append(e.config.Pages[e.currentPage], api.Key{})
		}
		b.(*button).key = e.config.Pages[e.currentPage][b.(*button).keyID]
		b.Refresh()
	}

	e.refreshEditor()
}

func (e *editor) registerPageListener() {
	err := conn.RegisterPageListener(e.pageListener)
	if err != nil {
		dialog.ShowError(err, e.win)
	}
}

func (e *editor) reset() {
	for _, b := range e.buttons {
		newKey := api.Key{}
		b.(*button).key = newKey
		e.config.Pages[e.currentPage][b.(*button).keyID] = newKey
		b.Refresh()
	}
	e.refreshEditor()

	err := conn.SetConfig(e.config)
	if err != nil {
		dialog.ShowError(err, e.win)
	}
}

func (e *editor) setPage(page int) {
	err := conn.SetPage(page)
	if err != nil {
		dialog.ShowError(err, e.win)
		return
	}

	text := fmt.Sprintf("%d/%d", page+1, len(e.config.Pages))
	e.pageLabel.label.SetText(text)
	e.currentPage = page
	e.currentButton = nil
	e.refresh()
}

func (e *editor) loadToolbar() *widget.Toolbar {
	e.pageLabel = newToolbarLabel()
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
			err := conn.SetConfig(e.config)
			if err != nil {
				dialog.ShowError(err, e.win)
			}
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			err := conn.CommitConfig()
			if err != nil {
				dialog.ShowError(err, e.win)
			}
		}),
		widget.NewToolbarAction(theme.ContentUndoIcon(), func() {
			err := conn.ReloadConfig()
			if err != nil {
				dialog.ShowError(err, e.win)
			}
			c, err := conn.GetConfig()
			if err != nil {
				dialog.ShowError(err, e.win)
			}
			e.config = c
			e.refresh()
		}),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			dialog.ShowConfirm("Reset config?", "Are you sure you want to reset?",
				func(ok bool) {
					if ok {
						e.reset()
					}
				}, e.win)
		}),

		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MediaSkipPreviousIcon(), func() {
			if e.currentPage == 0 {
				return
			}

			e.setPage(e.currentPage - 1)
		}),
		e.pageLabel,
		widget.NewToolbarAction(theme.MediaSkipNextIcon(), func() {
			if e.currentPage == len(e.config.Pages)-1 {
				return
			}

			e.setPage(e.currentPage + 1)
		}),
		widget.NewToolbarSpacer(),

		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			if e.currentPage == len(e.config.Pages)-1 {
				e.config.Pages = append(e.config.Pages, e.emptyPage())
			} else {
				e.config.Pages = append(e.config.Pages, api.Page{}) // dummy value
				for i := len(e.config.Pages) - 1; i > e.currentPage; i-- {
					e.config.Pages[i] = e.config.Pages[i-1]
				}
				e.config.Pages[e.currentPage+1] = e.emptyPage()
			}
			err := conn.SetConfig(e.config)
			if err != nil {
				dialog.ShowError(err, e.win)
				return
			}
			e.setPage(e.currentPage + 1)
		}),
		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			if len(e.config.Pages) == 1 {
				e.reset()
				return
			}

			for i := len(e.config.Pages) - 1; i > e.currentPage; i-- {
				e.config.Pages[i-1] = e.config.Pages[i]
			}
			e.config.Pages = e.config.Pages[:len(e.config.Pages)-1]

			e.setPage(e.currentPage - 1)
			err := conn.SetConfig(e.config)
			if err != nil {
				dialog.ShowError(err, e.win)
				return
			}
		}),
	)
}

func (e *editor) loadUI() fyne.CanvasObject {
	toolbar := e.loadToolbar()
	var page api.Page
	if len(e.config.Pages) >= 1 {
		page = e.config.Pages[e.currentPage]
	}

	for i := 0; i < e.pageCols*e.pageRows; i++ {
		var key api.Key
		if i < len(page) {
			key = page[i]
		}
		btn := newButton(key, i, e)
		if i == 0 {
			e.currentButton = btn
		}
		e.buttons = append(e.buttons, btn)
	}

	editor := e.loadEditor()
	e.setPage(e.currentPage)

	grid := fyne.NewContainerWithLayout(layout.NewGridLayout(e.pageCols),
		e.buttons...)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(toolbar, editor, nil, nil),
		toolbar, editor, grid)
}

type toolbarLabel struct {
	label *widget.Label
}

func (t *toolbarLabel) ToolbarObject() fyne.CanvasObject {
	return t.label
}

func newToolbarLabel() *toolbarLabel {
	return &toolbarLabel{label: widget.NewLabel("0")}
}
