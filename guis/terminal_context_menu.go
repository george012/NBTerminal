package guis

import (
	"fmt"
	"strings"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit"
)

func terminalContextMenuItems(terminal *uikit.UITerminalView, state uikit.ContextMenuState, find, export, clear func()) []uikit.MenuItem {
	copyFlags := 0
	if !state.HasSelection {
		copyFlags = fltk_bridge.MENU_INACTIVE
	}
	pasteFlags := fltk_bridge.MENU_DIVIDER
	if !state.InputEnabled {
		pasteFlags |= fltk_bridge.MENU_INACTIVE
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
			Flags: pasteFlags,
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
	selected, pinned, bellMuted, inputLocked, reconnectable, reopenable, closable, closeOthers, closeLeft, closeRight bool
}

type sessionTabMenuActions struct {
	activate, rename, pin, muteBell, lockInput, duplicate, reconnect, reopen, close, closeOthers, closeLeft, closeRight func()
}

func sessionTabContextMenuItems(state sessionTabMenuState, actions sessionTabMenuActions) []uikit.MenuItem {
	inactiveWhen := func(enabled bool) int {
		if enabled {
			return 0
		}
		return fltk_bridge.MENU_INACTIVE
	}
	pinTitle := "Pin Session"
	if state.pinned {
		pinTitle = "Unpin Session"
	}
	bellTitle := "Mute Bell"
	if state.bellMuted {
		bellTitle = "Unmute Bell"
	}
	inputTitle := "Lock Input	Ctrl+Shift+I"
	if state.inputLocked {
		inputTitle = "Unlock Input	Ctrl+Shift+I"
	}
	return []uikit.MenuItem{
		{Title: "Activate Session", Flags: inactiveWhen(!state.selected), Callback: actions.activate},
		{Title: "Rename Session…", Callback: actions.rename},
		{Title: pinTitle, Callback: actions.pin},
		{Title: bellTitle, Callback: actions.muteBell},
		{Title: inputTitle, Callback: actions.lockInput},
		{Title: "Duplicate Session	Ctrl+Shift+D", Callback: actions.duplicate},
		{Title: "Reconnect Session	Ctrl+Shift+R", Flags: inactiveWhen(state.reconnectable), Callback: actions.reconnect},
		{Title: "Reopen Closed Session	Ctrl+Shift+T", Flags: inactiveWhen(state.reopenable) | fltk_bridge.MENU_DIVIDER, Callback: actions.reopen},
		{Title: "Close Session	Ctrl+W", Flags: inactiveWhen(state.closable), Callback: actions.close},
		{Title: "Close Other Sessions", Flags: inactiveWhen(state.closeOthers), Callback: actions.closeOthers},
		{Title: "Close Sessions to the Left", Flags: inactiveWhen(state.closeLeft), Callback: actions.closeLeft},
		{Title: "Close Sessions to the Right", Flags: inactiveWhen(state.closeRight), Callback: actions.closeRight},
	}
}

type sessionTabListEntry struct {
	ID, Title        string
	Selected, Pinned bool
}

func sessionTabListMenuTitle(title string) string {
	title = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\':
			return '-'
		case '\n', '\r', '	':
			return ' '
		}
		if r < ' ' || r == 0x7f {
			return -1
		}
		return r
	}, title)
	title = strings.TrimSpace(title)
	if title == "" {
		return "Session"
	}
	return title
}

func sessionTabListMenuItems(entries []sessionTabListEntry, activate func(id string)) []uikit.MenuItem {
	items := make([]uikit.MenuItem, 0, len(entries))
	for _, entry := range entries {
		entry := entry
		title := sessionTabListMenuTitle(entry.Title)
		if entry.Pinned {
			title = "[Pinned] " + title
		}
		flags := fltk_bridge.MENU_RADIO
		if entry.Selected {
			flags |= fltk_bridge.MENU_VALUE
		}
		items = append(items, uikit.MenuItem{Title: title, Flags: flags, Callback: func() {
			if activate != nil {
				activate(entry.ID)
			}
		}})
	}
	return items
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
			pinned:        state.Pinned,
			bellMuted:     state.BellMuted,
			inputLocked:   state.InputLocked,
			reconnectable: reconnectable,
			reopenable:    a.sessions.CanReopenClosed(),
			closable:      state.Status != sessionRunning && !state.Pinned,
			closeOthers:   a.canCloseOtherSessions(request.ID),
			closeLeft:     a.canCloseSessionsToLeft(request.ID),
			closeRight:    a.canCloseSessionsToRight(request.ID),
		}, sessionTabMenuActions{
			activate:  func() { a.activateSessionByID(request.ID) },
			rename:    func() { a.openSessionRename(request.ID) },
			pin:       func() { a.togglePinnedSession(request.ID) },
			muteBell:  func() { a.toggleSessionBell(request.ID) },
			lockInput: func() { a.toggleSessionInputLock(request.ID) },
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
			reopen:      a.reopenLastClosedSession,
			close:       func() { a.closeSessionByID(request.ID) },
			closeOthers: func() { a.closeOtherSessions(request.ID) },
			closeLeft:   func() { a.closeSessionsToLeft(request.ID) },
			closeRight:  func() { a.closeSessionsToRight(request.ID) },
		}))
		menu.Popup()
	})
}

func (a *finalShellApp) installSessionTabListMenu(parent *uikit.UIGroup) {
	if a == nil || parent == nil || a.sessionTabs == nil {
		return
	}
	menu := uikit.NewUIContextMenu(rect(0, 0, 0, 0))
	parent.AddSubview(menu)
	a.sessionListMenu = menu
	a.sessionTabs.OnTabListRequested(func(items []uikit.TabListItem) {
		entries := make([]sessionTabListEntry, len(items))
		for index, item := range items {
			entries[index] = sessionTabListEntry{ID: item.ID, Title: item.Title, Selected: item.Selected, Pinned: item.Pinned}
		}
		menu.SetMenu(sessionTabListMenuItems(entries, func(id string) { a.activateSessionByID(id) }))
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

func (a *finalShellApp) togglePinnedSession(id string) {
	index := a.sessionIndexByID(id)
	if index < 0 {
		return
	}
	states := a.sessions.Tabs()
	state := states[index]
	pinned := !state.Pinned
	nativeIndex := index
	if a.sessionTabs != nil {
		var ok bool
		nativeIndex, ok = a.sessionTabs.SetTabPinned(index, pinned)
		if !ok {
			return
		}
	}
	if !a.sessions.SetPinned(id, pinned) {
		if a.sessionTabs != nil {
			a.sessionTabs.SetTabPinned(nativeIndex, !pinned)
		}
		return
	}
	a.refreshSessionTabs()
	if state.Pinned {
		a.setStatus(fmt.Sprintf("Unpinned %s", state.Profile.Name))
	} else {
		a.setStatus(fmt.Sprintf("Pinned %s", state.Profile.Name))
	}
}

func (a *finalShellApp) toggleSessionBell(id string) {
	index := a.sessionIndexByID(id)
	if index < 0 {
		return
	}
	state := a.sessions.Tabs()[index]
	muted := !state.BellMuted
	if !a.sessions.SetBellMuted(id, muted) {
		return
	}
	a.refreshSessionTabs()
	if muted {
		a.setStatus(fmt.Sprintf("Muted bell for %s", state.Profile.Name))
	} else {
		a.setStatus(fmt.Sprintf("Unmuted bell for %s", state.Profile.Name))
	}
}

func (a *finalShellApp) toggleSessionInputLock(id string) {
	index := a.sessionIndexByID(id)
	if index < 0 {
		return
	}
	state := a.sessions.Tabs()[index]
	locked := !state.InputLocked
	if !a.sessions.SetInputLocked(id, locked) {
		return
	}
	if index == a.sessions.ActiveIndex() && a.output != nil {
		a.output.SetInputEnabled(!locked)
		a.updateCommandControls()
	}
	a.refreshSessionTabs()
	if locked {
		a.setStatus(fmt.Sprintf("Locked input for %s", state.Profile.Name))
	} else {
		a.setStatus(fmt.Sprintf("Unlocked input for %s", state.Profile.Name))
	}
}

func (a *finalShellApp) toggleActiveSessionInputLock() {
	if id := a.activeSessionID(); id != "" {
		a.toggleSessionInputLock(id)
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
		if state.Pinned {
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
