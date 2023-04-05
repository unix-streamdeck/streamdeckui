package main

import (
    "fmt"
    "log"
    "os/exec"
    "strings"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"
    "github.com/unix-streamdeck/api"
)

type editor struct {
    currentButton       *button
    copiedButton        *button
    config              *api.ConfigV3
    info                []*api.StreamDeckInfoV1
    currentDeviceConfig *api.DeckV3
    currentDevice       *api.StreamDeckInfoV1
    deviceButtons       map[string][]fyne.CanvasObject
    layouts             map[string]*fyne.Container
    deviceSelector      *widget.Select

    iconHandler, keyHandler, applicationSelectorSelect *widget.Select
    pageLabel                                          *toolbarLabel
    buttons                                            []fyne.CanvasObject
    keyDetailSelector, iconDetailSelector              *fyne.Container

    selectedApplication string

    copiedKey api.KeyV3

    win fyne.Window
}

func newEditor(info []*api.StreamDeckInfoV1, w fyne.Window) *editor {
    c, err := conn.GetConfig()
    if err != nil {
        dialog.ShowError(err, w)
        c = &api.ConfigV3{}
    }
    currentDevice := info[0]
    var config *api.DeckV3
    for i := range c.Decks {
        if c.Decks[i].Serial == currentDevice.Serial {
            config = &c.Decks[i]
        }
    }
    ed := &editor{config: c, info: info, win: w, currentDevice: currentDevice, currentDeviceConfig: config,
        deviceButtons: make(map[string][]fyne.CanvasObject), layouts: make(map[string]*fyne.Container)}
    go ed.registerPageListener()
    return ed
}

func (e *editor) loadEditor() fyne.CanvasObject {

    e.applicationSelectorSelect = widget.NewSelect(getWindowList(), e.chooseActiveApplication)

    applicationSelector := widget.NewForm(
        widget.NewFormItem("Application", e.applicationSelectorSelect),
    )

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
    iconForm := fyne.NewContainerWithLayout(layout.NewFormLayout(), fyne.NewContainerWithLayout(layout.NewCenterLayout(), iconHandler), e.iconDetailSelector)
    keyForm := fyne.NewContainerWithLayout(layout.NewFormLayout(), fyne.NewContainerWithLayout(layout.NewCenterLayout(), keyHandler), e.keyDetailSelector)
    tabs := container.NewAppTabs(
        container.NewTabItem("Icon Config", iconForm),
        container.NewTabItem("Keypress Config", keyForm),
    )
    tabs.SetTabLocation(container.TabLocationTop)
    return fyne.NewContainerWithLayout(layout.NewBorderLayout(applicationSelector, nil, nil, nil), applicationSelector, tabs)
}

func (e *editor) chooseKeyHandler(name string) {
    e.chooseHandler(name, "Key")
}

func (e *editor) chooseIconHandler(name string) {
    e.chooseHandler(name, "Icon")
}

func (e *editor) chooseActiveApplication(name string) {
    if name == "Default" {
        name = ""
    }
    e.selectedApplication = name
    e.currentButton.updateApplication(name)
    e.refresh()
    e.refreshEditor()
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
        if e.currentButton.currentConfig.KeyHandlerFields == nil {
            e.currentButton.currentConfig.KeyHandlerFields = make(map[string]string)
        }
        itemMap = e.currentButton.currentConfig.KeyHandlerFields

    } else {
        fields = module.IconFields
        if e.currentButton.currentConfig.IconHandlerFields == nil {
            e.currentButton.currentConfig.IconHandlerFields = make(map[string]string)
        }
        itemMap = e.currentButton.currentConfig.IconHandlerFields
    }

    if fields != nil {
        ui = loadUI(fields, itemMap, e)
    } else {
        ui = widget.NewForm()
    }

    if name == "Default" {
        if handlerType == "Key" {
            ui = loadDefaultKeyUI(e)
            e.currentButton.currentConfig.KeyHandler = "Default"
        } else {
            ui = loadDefaultIconUI(e)
            e.currentButton.currentConfig.IconHandler = "Default"
        }
    } else {
        if handlerType == "Key" {
            e.currentButton.currentConfig.KeyHandler = name
        } else {
            e.currentButton.currentConfig.IconHandler = name
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

    e.applicationSelectorSelect.SetSelectedIndex(0)
    e.selectedApplication = ""

    e.refreshEditor()
}

func (e *editor) emptyPage() api.PageV3 {
    var page api.PageV3
    for i := 0; i < e.currentDevice.Cols*e.currentDevice.Rows; i++ {
        page.Keys = append(page.Keys, api.KeyV3{
            Application: map[string]*api.KeyConfigV3{"": {}},
        })
    }

    return page
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
        handler := e.currentButton.currentConfig.KeyHandler
        if handler == "" {
            handler = "Default"
        }
        e.keyHandler.SetSelected(handler)
        handler = e.currentButton.currentConfig.IconHandler
        if handler == "" {
            handler = "Default"
        }
        e.iconHandler.SetSelected(handler)
    }
}

func (e *editor) refresh() {
    for _, b := range e.buttons {
        go e.refreshButton(b)
    }

    e.refreshEditor()
}

func (e *editor) refreshButton(b fyne.CanvasObject) {
    if e.currentButton == nil {
        e.currentButton = b.(*button)
    }
    if b.(*button).keyID >= len(e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys) {
        e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys = append(e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys, api.KeyV3{})
    }
    var currentConfig *api.KeyConfigV3
    if api.CompareKeys(b.(*button).key, e.currentButton.key) {
        var ok bool
        currentConfig, ok = e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys[b.(*button).keyID].Application[e.selectedApplication]
        if !ok {
            currentConfig = &api.KeyConfigV3{}
            if e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys[b.(*button).keyID].Application != nil {
                e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys[b.(*button).keyID].Application[e.selectedApplication] = currentConfig
            }
        }
    } else {
        currentConfig, _ = e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys[b.(*button).keyID].Application[""]
    }

    b.(*button).currentConfig = currentConfig
    b.(*button).key = e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys[b.(*button).keyID]
    b.Refresh()
}

func (e *editor) registerPageListener() {
    err := conn.RegisterPageListener(e.pageListener)
    if err != nil {
        dialog.ShowError(err, e.win)
    }
}

func (e *editor) reset() {
    for _, b := range e.buttons {
        newKey := api.KeyV3{}
        b.(*button).key = newKey
        e.currentDeviceConfig.Pages[e.currentDevice.Page].Keys[b.(*button).keyID] = newKey
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

// Save config. Used by both the toolbar action and the keyboard shortcut
func (e *editor) saveConfig() {
    err := conn.SetConfig(e.config)
    if err != nil {
        dialog.ShowError(err, e.win)
        return
    }
    err = conn.CommitConfig()
    if err != nil {
        dialog.ShowError(err, e.win)
    }
}

// Copy current button. Used by both the toolbar action and the keyboard shortcut
func (e *editor) copyButton() {
    e.copiedButton = e.currentButton
}

// Paste copied button, if any. Used by both the toolbar action and the keyboard shortcut
func (e *editor) pasteButton() {
    if e.copiedButton != nil {
        e.currentButton.key = e.copiedButton.key
        e.refreshEditor()
    }
}

func (e *editor) loadToolbar() *widget.Toolbar {
    e.pageLabel = newToolbarLabel("0")

    fileMenu := fyne.NewMenu("File",
        fyne.NewMenuItem("Preview Config on Device", func() {
            config := e.cleanupConfig(e.config)
            err := conn.SetConfig(config)
            if err != nil {
                dialog.ShowError(err, e.win)
            }
        }),
        fyne.NewMenuItem("Save", e.saveConfig),
        fyne.NewMenuItem("Reset Config To Empty Config", func() {
            dialog.ShowConfirm("Reset config?", "Are you sure you want to reset?",
                func(ok bool) {
                    if ok {
                        e.reset()
                    }
                }, e.win)
        }),
        fyne.NewMenuItemSeparator(),
        fyne.NewMenuItem("Quit", func() {
            e.win.Close()
        }),
    )

    editMenu := fyne.NewMenu(
        "Edit",
        fyne.NewMenuItem("Copy", e.copyButton),
        fyne.NewMenuItem("Paste", e.pasteButton),
    )

    actionsMenu := fyne.NewMenu(
        "Actions",
        fyne.NewMenuItem("Run Buttons Actions", func() {
            err := conn.PressButton(e.currentDevice.Serial, e.currentButton.keyID)
            if err != nil {
                fyne.LogError("Failed to run button press", err)
            }
        }),
        fyne.NewMenuItem("Reload Config From Disk", func() {
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
        fyne.NewMenuItemSeparator(),
        fyne.NewMenuItem("Add Page", func() {
            e.currentDeviceConfig.Pages = append(e.currentDeviceConfig.Pages, e.emptyPage())
            err := conn.SetConfig(e.config)
            if err != nil {
                dialog.ShowError(err, e.win)
                return
            }
            e.setPage(len(e.currentDeviceConfig.Pages)-1, true)
        }),
        fyne.NewMenuItem("Remove Page", func() {
            if len(e.currentDeviceConfig.Pages) == 1 {
                e.reset()
                return
            }

            for i := len(e.currentDeviceConfig.Pages) - 1; i > e.currentDevice.Page; i-- {
                e.currentDeviceConfig.Pages[i-1] = e.currentDeviceConfig.Pages[i]
            }
            e.currentDeviceConfig.Pages = e.currentDeviceConfig.Pages[:len(e.currentDeviceConfig.Pages)-1]

            e.setPage(e.currentDevice.Page-1, true)
            err := conn.SetConfig(e.config)
            if err != nil {
                dialog.ShowError(err, e.win)
                return
            }
        }),
    )

    e.win.SetMainMenu(fyne.NewMainMenu(fileMenu, editMenu, actionsMenu))

    return widget.NewToolbar(
        widget.NewToolbarSpacer(),
        widget.NewToolbarAction(theme.MediaSkipPreviousIcon(), func() {
            if e.currentDevice.Page == 0 {
                return
            }

            e.setPage(e.currentDevice.Page-1, true)
        }),
        e.pageLabel,
        widget.NewToolbarAction(theme.MediaSkipNextIcon(), func() {
            if e.currentDevice.Page == len(e.currentDeviceConfig.Pages)-1 {
                return
            }

            e.setPage(e.currentDevice.Page+1, true)
        }),
        widget.NewToolbarSpacer(),
    )
}

func (e *editor) loadUI() fyne.CanvasObject {
    toolbar := e.loadToolbar()
    var page api.PageV3
    if len(e.currentDeviceConfig.Pages) >= 1 {
        page = e.currentDeviceConfig.Pages[e.currentDevice.Page]
    }

    var layouts []*fyne.Container
    for j := range e.info {
        var buttons []fyne.CanvasObject
        for i := 0; i < e.info[j].Cols*e.info[j].Rows; i++ {
            var key api.KeyV3
            if i < len(page.Keys) {
                key = page.Keys[i]
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
        deviceIDs = append(deviceIDs, fmt.Sprintf("%s: %s", e.info[i].Name, e.info[i].Serial))
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

func (e *editor) cleanupConfig(config *api.ConfigV3) *api.ConfigV3 {
    for deckIndex := range config.Decks {
        deck := &config.Decks[deckIndex]
        for pageIndex := range deck.Pages {
            page := &deck.Pages[pageIndex]
            for keyIndex := range page.Keys {
                key := &deck.Pages[pageIndex].Keys[keyIndex]
                for app, keyConfig := range key.Application {
                    if api.CompareKeyConfigs(*keyConfig, api.KeyConfigV3{}) {
                        delete(key.Application, app)
                    }
                }
            }
        }
    }
    return config
}

type ToolbarActionWithLabel struct {
    Icon        fyne.Resource
    label       string
    OnActivated func()
    editor      editor
}

// ToolbarObject gets a button to render this ToolbarAction
func (t *ToolbarActionWithLabel) ToolbarObject() fyne.CanvasObject {
    button := widget.NewButtonWithIcon(t.label, t.Icon, t.OnActivated)
    button.Importance = widget.LowImportance

    return button
}

func newToolBarActionWithLabel(text string, icon fyne.Resource, onActivated func()) *ToolbarActionWithLabel {
    return &ToolbarActionWithLabel{label: text, Icon: icon, OnActivated: onActivated}
}

type toolbarLabel struct {
    label *widget.Label
}

func (t *toolbarLabel) ToolbarObject() fyne.CanvasObject {
    return t.label
}

func newToolbarLabel(text string) *toolbarLabel {
    return &toolbarLabel{label: widget.NewLabel(text)}
}

func getWindowList() []string {
    command := "xdotool search --onlyvisible --desktop 0 --name \".*\" | xargs -n 1 xdotool getwindowclassname"
    processes, err := exec.Command("/bin/sh", "-c", command).Output()
    if err != nil {
        log.Fatalln(err)
    }
    processList := strings.Split(strings.Trim(string(processes), "\n"), "\n")
    processList = deduplicate(processList)
    return processList
}

func deduplicate(s []string) []string {
    inResult := make(map[string]bool)
    var result []string
    result = append(result, "Default")
    for _, str := range s {
        if str == "" {
            continue
        }
        if _, ok := inResult[str]; !ok {
            inResult[str] = true
            result = append(result, str)
        }
    }
    return result
}
