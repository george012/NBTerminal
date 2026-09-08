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

type sessionTabMenuState struct {
	selected, reconnectable, closable, closeOthers, closeLeft, closeRight bool
}

type sessionTabMenuActions struct {
	activate, duplicate, reconnect, close, closeOthers, closeLeft, closeRight func()
}

func sessionTabContextMenuItems(state sessionTabMenuState, actions sessionTabMenuActions) []uikit.MenuItem {
	inactiveWhen := func(enabled bool) int {
		if enabled {
			return 0
		}
		return fltk_bridge.MENU_INACTIVE
	}
	return []uikit.MenuItem{
		{Title: "Activate Session", Flags: inactiveWhen(!state.selected), Callback: actions.activate},
		{Title: "Duplicate Session	Ctrl+Shift+D", Callback: actions.duplicate},
		{Title: "Reconnect Session	Ctrl+Shift+R", Flags: inactiveWhen(state.reconnectable) | fltk_bridge.MENU_DIVIDER, Callback: actions.reconnect},
		{Title: "Close Session	Ctrl+W", Flags: inactiveWhen(state.closable), Callback: actions.close},
		{Title: "Close Other Sessions", Flags: inactiveWhen(state.closeOthers), Callback: actions.closeOthers},
		{Title: "Close Sessions to the Left", Flags: inactiveWhen(state.closeLeft), Callback: actions.closeLeft},
		{Title: "Close Sessions to the Right", Flags: inactiveWhen(state.closeRight), Callback: actions.closeRight},
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
		menu.SetMenu(sessionTabContextMenuItems(sessionTabMenuState{
			selected:      index == a.sessions.ActiveIndex(),
			reconnectable: reconnectable,
			closable:      state.Status != sessionRunning,
			closeOthers:   a.canCloseOtherSessions(request.ID),
			closeLeft:     a.canCloseSessionsToLeft(request.ID),
			closeRight:    a.canCloseSessionsToRight(request.ID),
		}, sessionTabMenuActions{
			activate: func() { a.activateSessionByID(request.ID) },
			duplicate: func() {
				if a.activateSessionByID(request.ID) {
					a.duplicateActiveSession()
				}
			},
			reconnect: func() {
				if a.activateSessionByID(request.ID) {
					a.reconnectActiveSession()
				}
			},
			close:       func() { a.closeSessionByID(request.ID) },
			closeOthers: func() { a.closeOtherSessions(request.ID) },
			closeLeft:   func() { a.closeSessionsToLeft(request.ID) },
			closeRight:  func() { a.closeSessionsToRight(request.ID) },
		}))
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

type sessionBatchCloseMode uint8

const (
	closeOtherSessions sessionBatchCloseMode = iota
	closeSessionsToLeft
	closeSessionsToRight
)

// batchCloseSessionIDs resolves the affected runtimes from the current stable
// identity and fails closed when any one of them is running. This prevents a
// context-menu command from leaving a surprising partially closed workspace.
func (a *finalShellApp) batchCloseSessionIDs(targetID string, mode sessionBatchCloseMode) ([]string, bool) {
	if a == nil || a.sessions == nil {
		return nil, false
	}
	target := a.sessionIndexByID(targetID)
	if target < 0 {
		return nil, false
	}
	states := a.sessions.Tabs()
	ids := make([]string, 0, len(states)-1)
	for index, state := range states {
		if index == target || (mode == closeSessionsToLeft && index > target) || (mode == closeSessionsToRight && index < target) {
			continue
		}
		if state.Status == sessionRunning {
			return nil, false
		}
		ids = append(ids, state.ID)
	}
	return ids, len(ids) > 0
}

func (a *finalShellApp) canCloseOtherSessions(targetID string) bool {
	_, ok := a.batchCloseSessionIDs(targetID, closeOtherSessions)
	return ok
}

func (a *finalShellApp) canCloseSessionsToLeft(targetID string) bool {
	_, ok := a.batchCloseSessionIDs(targetID, closeSessionsToLeft)
	return ok
}

func (a *finalShellApp) canCloseSessionsToRight(targetID string) bool {
	_, ok := a.batchCloseSessionIDs(targetID, closeSessionsToRight)
	return ok
}

func (a *finalShellApp) closeOtherSessions(targetID string) {
	a.closeSessionBatch(targetID, closeOtherSessions)
}

func (a *finalShellApp) closeSessionsToLeft(targetID string) {
	a.closeSessionBatch(targetID, closeSessionsToLeft)
}

func (a *finalShellApp) closeSessionsToRight(targetID string) {
	a.closeSessionBatch(targetID, closeSessionsToRight)
}

func (a *finalShellApp) closeSessionBatch(targetID string, mode sessionBatchCloseMode) {
	ids, ok := a.batchCloseSessionIDs(targetID, mode)
	if !ok {
		return
	}
	// Resolve every stable ID immediately before closing. Reverse order avoids
	// needless index churn and keeps the right-click target authoritative.
	for index := len(ids) - 1; index >= 0; index-- {
		a.closeSessionByID(ids[index])
	}
	switch mode {
	case closeSessionsToLeft:
		a.setStatus("Closed sessions to the left")
	case closeSessionsToRight:
		a.setStatus("Closed sessions to the right")
	default:
		// Closing every peer necessarily makes the right-click target active.
		// Select it explicitly so native and workspace state converge even when
		// callbacks are unavailable in a lightweight embedding.
		a.activateSessionByID(targetID)
		a.setStatus("Closed other sessions")
	}
}
