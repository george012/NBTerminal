package guis

import (
	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit"
)

func terminalContextMenuItems(terminal *uikit.UITerminalView, state uikit.ContextMenuState, find, export, clear func()) []uikit.MenuItem {
	copyFlags := 0
	if !state.HasSelection {
		copyFlags = fltk_bridge.MENU_INACTIVE
	}
	return []uikit.MenuItem{
		{
			Title: "Copy	Ctrl+Shift+C",
			Flags: copyFlags,
			Callback: func() {
				if terminal != nil {
					terminal.CopySelection()
				}
			},
		},
		{
			Title: "Copy All Output	Ctrl+Shift+A",
			Callback: func() {
				if terminal != nil {
					terminal.CopyAllText()
				}
			},
		},
		{
			Title: "Paste	Ctrl+Shift+V",
			Flags: fltk_bridge.MENU_DIVIDER,
			Callback: func() {
				if terminal != nil {
					terminal.PasteClipboard()
				}
			},
		},
		{Title: "Find in Terminal	Ctrl+Shift+F", Callback: find},
		{Title: "Save Terminal Output…	Ctrl+Shift+S", Callback: export},
		{Title: "Clear Terminal	Ctrl+Shift+L", Callback: clear},
	}
}

func sessionTabContextMenuItems(selected, reconnectable, closable bool, activate, duplicate, reconnect, close func()) []uikit.MenuItem {
	inactiveWhen := func(enabled bool) int {
		if enabled {
			return 0
		}
		return fltk_bridge.MENU_INACTIVE
	}
	return []uikit.MenuItem{
		{Title: "Activate Session", Flags: inactiveWhen(!selected), Callback: activate},
		{Title: "Duplicate Session	Ctrl+Shift+D", Callback: duplicate},
		{Title: "Reconnect Session	Ctrl+Shift+R", Flags: inactiveWhen(reconnectable) | fltk_bridge.MENU_DIVIDER, Callback: reconnect},
		{Title: "Close Session	Ctrl+W", Flags: inactiveWhen(closable), Callback: close},
	}
}

func (a *finalShellApp) installTerminalContextMenu(parent *uikit.UIGroup) {
	if a == nil || parent == nil || a.output == nil {
		return
	}
	menu := uikit.NewUIContextMenu(rect(0, 0, 0, 0))
	parent.AddSubview(menu)
	a.terminalContextMenu = menu
	a.output.OnContextMenu(func(state uikit.ContextMenuState) {
		menu.SetMenu(terminalContextMenuItems(a.output, state, a.openTerminalFind, a.exportTerminalOutput, a.clearTerminalOutput))
		menu.Popup()
	})
}

func (a *finalShellApp) installSessionTabContextMenu(parent *uikit.UIGroup) {
	if a == nil || parent == nil || a.sessionTabs == nil {
		return
	}
	menu := uikit.NewUIContextMenu(rect(0, 0, 0, 0))
	parent.AddSubview(menu)
	a.sessionContextMenu = menu
	a.sessionTabs.OnTabContextMenuRequested(func(request uikit.TabContextMenuState) {
		index := a.sessionIndexByID(request.ID)
		if index < 0 {
			return
		}
		state := a.sessions.Tabs()[index]
		reconnectable := state.Profile.Type == connectionTypeLocal || state.Profile.Type == connectionTypeSSH
		menu.SetMenu(sessionTabContextMenuItems(
			index == a.sessions.ActiveIndex(),
			reconnectable,
			state.Status != sessionRunning,
			func() { a.activateSessionByID(request.ID) },
			func() {
				if a.activateSessionByID(request.ID) {
					a.duplicateActiveSession()
				}
			},
			func() {
				if a.activateSessionByID(request.ID) {
					a.reconnectActiveSession()
				}
			},
			func() { a.closeSessionByID(request.ID) },
		))
		menu.Popup()
	})
}

func (a *finalShellApp) sessionIndexByID(id string) int {
	if a == nil || a.sessions == nil {
		return -1
	}
	for index, state := range a.sessions.Tabs() {
		if state.ID == id {
			return index
		}
	}
	return -1
}

func (a *finalShellApp) activateSessionByID(id string) bool {
	index := a.sessionIndexByID(id)
	if index < 0 || a.sessionTabs == nil {
		return false
	}
	if index == a.sessions.ActiveIndex() {
		return true
	}
	a.sessionTabs.SelectTab(index)
	return a.sessions.ActiveIndex() == index
}

func (a *finalShellApp) closeSessionByID(id string) {
	if index := a.sessionIndexByID(id); index >= 0 {
		a.closeSessionAt(index)
	}
}
