package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/unix-streamdeck/api"
	"strings"
)

type editor struct {
	currentButton       *button
	config              *api.Config
	info                []*api.StreamDeckInfo
	currentDeviceConfig *api.Deck
	currentDevice       *api.StreamDeckInfo
	deviceButtons       map[string][]fyne.CanvasObject
	layouts             map[string]*fyne.Container
	deviceSelector		*widget.Select

	iconHandler, keyHandler               *widget.Select
	pageLabel                             *toolbarLabel
	buttons                               []fyne.CanvasObject
	keyDetailSelector, iconDetailSelector *fyne.Container

	win fyne.Window
}

func newEditor(info []*api.StreamDeckInfo, w fyne.Window) *editor {
	c, err := conn.GetConfig()
	if err != nil {
		dialog.ShowError(err, w)
		c = &api.Config{}
	}
	currentDevice := info[0]
	var config *api.Deck
	for i := range c.Decks {
		if c.Decks[i].Serial == currentDevice.Serial {
			config = &c.Decks[i]
		}
	}
	ed := &editor{config: c, info: info, win: w, currentDevice: currentDevice, currentDeviceConfig: config,
		deviceButtons: make(map[string][]fyne.CanvasObject), layouts: make(map[string]*fyne.Container)}
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
	iconForm := fyne.NewContainerWithLayout(layout.NewFormLayout(), iconHandler, e.iconDetailSelector)
	keyForm := fyne.NewContainerWithLayout(layout.NewFormLayout(), keyHandler, e.keyDetailSelector)
	return fyne.NewContainerWithLayout(layout.NewVBoxLayout(), iconForm, widget.NewSeparator(), keyForm)
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
			e.currentButton.key.KeyHandler = "Default"
		} else {
			ui = loadDefaultIconUI(e)
			e.currentButton.key.IconHandler = "Default"
		}
	} else {
		if handlerType == "Key" {
			e.currentButton.key.KeyHandler = name
		} else {
			e.currentButton.key.IconHandler = name
		}
	}
	e.currentButton.updateKey()
	e.currentButton.Refresh()

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
	for i := 0; i < e.currentDevice.Cols*e.currentDevice.Rows; i++ {
		keys = append(keys, api.Key{})
	}

	return keys
}

func (e *editor) pageListener(serial string, page int32) {
	if e.currentDevice.Serial != serial {
		for i := range e.info {
			if e.info[i].Serial == serial {
				e.info[i].Page = int(page)
			}
		}
		return
	}

	if int(page) == e.currentDevice.Page {
		return
	}

	e.currentDevice.Page = int(page)
	e.setPage(int(page), false)
}

func (e *editor) refreshEditor() {
	if e.currentButton != nil {
		handler := e.currentButton.key.KeyHandler
		if handler == "" {
			handler = "Default"
		}
		if e.keyHandler.Selected != handler {
			e.keyHandler.SetSelected(handler)
		}
		handler = e.currentButton.key.IconHandler
		if handler == "" {
			handler = "Default"
		}
		if e.iconHandler.Selected != handler {
			e.iconHandler.SetSelected(handler)
		}
	}
}

func (e *editor) refresh() {
	for _, b := range e.buttons {
		if e.currentButton == nil {
			e.currentButton = b.(*button)
		}
		if b.(*button).keyID >= len(e.currentDeviceConfig.Pages[e.currentDevice.Page]) {
			e.currentDeviceConfig.Pages[e.currentDevice.Page] = append(e.currentDeviceConfig.Pages[e.currentDevice.Page], api.Key{})
		}
		b.(*button).key = e.currentDeviceConfig.Pages[e.currentDevice.Page][b.(*button).keyID]
		go b.Refresh()
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
		e.currentDeviceConfig.Pages[e.currentDevice.Page][b.(*button).keyID] = newKey
		b.Refresh()
	}
	e.refreshEditor()

	err := conn.SetConfig(e.config)
	if err != nil {
		dialog.ShowError(err, e.win)
	}
}

func (e *editor) setPage(page int, pushToDbus bool) {
	if pushToDbus {
		err := conn.SetPage(e.currentDevice.Serial, page)
		if err != nil {
			dialog.ShowError(err, e.win)
			return
		}
	}

	text := fmt.Sprintf("%d/%d", page+1, len(e.currentDeviceConfig.Pages))
	e.pageLabel.label.SetText(text)
	e.currentDevice.Page = page
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
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			dialog.ShowConfirm("Reset config?", "Are you sure you want to reset?",
				func(ok bool) {
					if ok {
						e.reset()
					}
				}, e.win)
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MediaSkipPreviousIcon(), func() {
			if e.currentDevice.Page == 0 {
				return
			}

			e.setPage(e.currentDevice.Page - 1, true)
		}),
		e.pageLabel,
		widget.NewToolbarAction(theme.MediaSkipNextIcon(), func() {
			if e.currentDevice.Page == len(e.currentDeviceConfig.Pages)-1 {
				return
			}

			e.setPage(e.currentDevice.Page + 1, true)
		}),
		widget.NewToolbarSpacer(),

		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			e.currentDeviceConfig.Pages = append(e.currentDeviceConfig.Pages, e.emptyPage())
			err := conn.SetConfig(e.config)
			if err != nil {
				dialog.ShowError(err, e.win)
				return
			}
			e.setPage(len(e.currentDeviceConfig.Pages) - 1, true)
		}),
		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			if len(e.currentDeviceConfig.Pages) == 1 {
				e.reset()
				return
			}

			for i := len(e.currentDeviceConfig.Pages) - 1; i > e.currentDevice.Page; i-- {
				e.currentDeviceConfig.Pages[i-1] = e.currentDeviceConfig.Pages[i]
			}
			e.currentDeviceConfig.Pages = e.currentDeviceConfig.Pages[:len(e.currentDeviceConfig.Pages)-1]

			e.setPage(e.currentDevice.Page - 1, true)
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
	if len(e.currentDeviceConfig.Pages) >= 1 {
		page = e.currentDeviceConfig.Pages[e.currentDevice.Page]
	}

	var layouts []*fyne.Container
	for j := range e.info {
		var buttons []fyne.CanvasObject
		for i := 0; i < e.info[j].Cols*e.info[j].Rows; i++ {
			var key api.Key
			if i < len(page) {
				key = page[i]
			}
			btn := newButton(key, i, e)
			if i == 0 {
				e.currentButton = btn
			}
			buttons = append(buttons, btn)
		}
		e.deviceButtons[e.info[j].Serial] = buttons
		buttonGrid := fyne.NewContainerWithLayout(layout.NewGridLayout(e.info[j].Cols),
			buttons...)
		e.layouts[e.info[j].Serial] = buttonGrid
		layouts = append(layouts, buttonGrid)
	}

	editor := e.loadEditor()
	e.setPage(e.currentDevice.Page, false)

	var deviceIDs []string

	for i := range e.info {
		deviceString := ""
		if e.info[i].Cols == 5 {
			deviceString = "Elgato Streamdeck Original: "
		} else if e.info[i].Cols == 3 {
			deviceString = "Elgato Streamdeck Mini: "
		} else if e.info[i].Cols == 8 {
			deviceString = "Elgato Streamdeck XL: "
		}
		deviceString += e.info[i].Serial
		deviceIDs = append(deviceIDs, deviceString)
	}

	e.deviceSelector = widget.NewSelect(deviceIDs, func(selected string) {
		serial := strings.Split(selected, ": ")[1]
		for i := range e.info {
			if e.info[i].Serial == serial {
				e.currentDevice = e.info[i]
			}
			e.layouts[e.info[i].Serial].Hide()
		}
		for i := range e.config.Decks {
			if e.config.Decks[i].Serial == serial {
				e.currentDeviceConfig = &e.config.Decks[i]
			}
		}
		e.buttons = e.deviceButtons[serial]
		container := e.layouts[serial]
		container.Show()
		for i := range e.info {
			if e.info[i].Serial == serial {
				e.setPage(e.info[i].Page, false)
			}
		}
	})
	e.buttons = e.deviceButtons[deviceIDs[0]]
	e.deviceSelector.SetSelectedIndex(0)

	form := widget.NewForm(widget.NewFormItem("Device: ", e.deviceSelector))

	if len(deviceIDs) == 1 {
		form.Hide()
	}

	topGrid := fyne.NewContainerWithLayout(layout.NewBorderLayout(toolbar, form, nil, nil), toolbar, form)

	layoutsCont := fyne.NewContainerWithLayout(layout.NewCenterLayout())

	for i := range layouts {
		layoutsCont.Add(layouts[i])
	}

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(topGrid, editor, nil, nil),
		topGrid, editor, layoutsCont)
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
