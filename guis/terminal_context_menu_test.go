package guis

import (
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit"
)

func TestTerminalContextMenuReflectsSelectionAndRoutesCommands(t *testing.T) {
	terminal := uikit.NewUITerminalView(rect(0, 0, 320, 160))
	cleared := 0
	found := 0
	exported := 0
	items := terminalContextMenuItems(terminal, uikit.ContextMenuState{}, func() { found++ }, func() { exported++ }, func() { cleared++ })
	if len(items) != 6 {
		t.Fatalf("context menu item count = %d, want 6", len(items))
	}
	if items[0].Title != "Copy	Ctrl+Shift+C" || items[0].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatalf("copy item without selection = %#v, want disabled Copy", items[0])
	}
	if items[1].Title != "Copy All Output	Ctrl+Shift+A" {
		t.Fatalf("copy-all item title = %q", items[1].Title)
	}
	if items[2].Title != "Paste	Ctrl+Shift+V" || items[2].Flags&fltk_bridge.MENU_DIVIDER == 0 {
		t.Fatalf("paste item = %#v, want Paste followed by divider", items[2])
	}
	if items[3].Title != "Find in Terminal	Ctrl+Shift+F" {
		t.Fatalf("find item title = %q", items[3].Title)
	}
	items[3].Callback()
	if found != 1 {
		t.Fatalf("find callback count = %d, want 1", found)
	}
	if items[4].Title != "Save Terminal Output…	Ctrl+Shift+S" {
		t.Fatalf("export item title = %q", items[4].Title)
	}
	items[4].Callback()
	if exported != 1 {
		t.Fatalf("export callback count = %d, want 1", exported)
	}
	if items[5].Title != "Clear Terminal	Ctrl+Shift+L" {
		t.Fatalf("clear item title = %q", items[5].Title)
	}
	items[5].Callback()
	if cleared != 1 {
		t.Fatalf("clear callback count = %d, want 1", cleared)
	}

	items = terminalContextMenuItems(terminal, uikit.ContextMenuState{HasSelection: true}, func() {}, func() {}, func() {})
	if items[0].Flags&fltk_bridge.MENU_INACTIVE != 0 {
		t.Fatal("copy item stayed disabled for selected terminal text")
	}
}

func TestSessionTabContextMenuReflectsRuntimeStateAndRoutesCommands(t *testing.T) {
	invocations := map[string]int{}
	items := sessionTabContextMenuItems(
		sessionTabMenuState{reconnectable: true, reopenable: true, closable: true, closeOthers: true, closeLeft: true, closeRight: true},
		sessionTabMenuActions{
			activate:    func() { invocations["activate"]++ },
			rename:      func() { invocations["rename"]++ },
			pin:         func() { invocations["pin"]++ },
			duplicate:   func() { invocations["duplicate"]++ },
			reconnect:   func() { invocations["reconnect"]++ },
			reopen:      func() { invocations["reopen"]++ },
			close:       func() { invocations["close"]++ },
			closeOthers: func() { invocations["close-others"]++ },
			closeLeft:   func() { invocations["close-left"]++ },
			closeRight:  func() { invocations["close-right"]++ },
		})
	if len(items) != 10 {
		t.Fatalf("session context menu item count = %d, want 10", len(items))
	}
	for index, title := range []string{"Activate Session", "Rename Session…", "Pin Session", "Duplicate Session	Ctrl+Shift+D", "Reconnect Session	Ctrl+Shift+R", "Reopen Closed Session	Ctrl+Shift+T", "Close Session	Ctrl+W", "Close Other Sessions", "Close Sessions to the Left", "Close Sessions to the Right"} {
		if items[index].Title != title || items[index].Flags&fltk_bridge.MENU_INACTIVE != 0 {
			t.Fatalf("session item %d = %#v", index, items[index])
		}
		items[index].Callback()
	}
	for _, action := range []string{"activate", "rename", "pin", "duplicate", "reconnect", "reopen", "close", "close-others", "close-left", "close-right"} {
		if invocations[action] != 1 {
			t.Fatalf("%s callback count = %d, want 1", action, invocations[action])
		}
	}

	items = sessionTabContextMenuItems(sessionTabMenuState{selected: true}, sessionTabMenuActions{})
	if items[0].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatal("active session kept Activate enabled")
	}
	if items[4].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatal("non-interactive session kept Reconnect enabled")
	}
	if items[5].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatal("empty close history kept Reopen enabled")
	}
	if items[6].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatal("running session kept Close enabled")
	}
	if items[7].Flags&fltk_bridge.MENU_INACTIVE == 0 || items[8].Flags&fltk_bridge.MENU_INACTIVE == 0 || items[9].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatal("unavailable batch close actions remained enabled")
	}

	items = sessionTabContextMenuItems(sessionTabMenuState{pinned: true}, sessionTabMenuActions{})
	if items[2].Title != "Unpin Session" || items[6].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatalf("pinned menu state = %#v", items)
	}
}

func TestSessionTabListMenuMarksActiveAndPinnedAndRoutesStableIdentity(t *testing.T) {
	selected := ""
	items := sessionTabListMenuItems([]sessionTabListEntry{
		{ID: "runtime-1", Title: "Production", Selected: false, Pinned: true},
		{ID: "runtime-2", Title: "Logs", Selected: true},
		{ID: "runtime-3", Title: "Shell", Selected: false},
	}, func(id string) { selected = id })
	if len(items) != 3 {
		t.Fatalf("tab list menu item count = %d, want 3", len(items))
	}
	if items[0].Title != "[Pinned] Production" || items[0].Flags&fltk_bridge.MENU_RADIO == 0 || items[0].Flags&fltk_bridge.MENU_VALUE != 0 {
		t.Fatalf("pinned tab list item = %#v", items[0])
	}
	if items[1].Title != "Logs" || items[1].Flags&fltk_bridge.MENU_RADIO == 0 || items[1].Flags&fltk_bridge.MENU_VALUE == 0 {
		t.Fatalf("selected tab list item = %#v", items[1])
	}
	items[0].Callback()
	if selected != "runtime-1" {
		t.Fatalf("tab list selected %q, want stable runtime-1", selected)
	}
	if got := sessionTabListMenuTitle(" Build/Logs\\Path\nNow	 "); got != "Build-Logs-Path Now" {
		t.Fatalf("menu-safe shell title = %q", got)
	}
	if got := sessionTabListMenuTitle("\x00\n"); got != "Session" {
		t.Fatalf("empty menu-safe shell title = %q", got)
	}
}

func TestSessionTabBatchCloseResolvesStableRuntimeAndPreservesTarget(t *testing.T) {
	workspace := newSessionWorkspace()
	for _, id := range []string{"one", "two", "three", "four"} {
		workspace.Open(connectionProfile{ID: id, Name: id, Type: connectionTypeLocal})
	}
	tabs := uikit.NewUITabView(rect(0, 0, 480, 180))
	for _, state := range workspace.Tabs() {
		tabs.AddTabWithID(state.ID, state.Profile.Name, nil)
	}
	tabs.SelectTab(0)
	workspace.Select(0)
	app := &finalShellApp{sessions: workspace, sessionTabs: tabs}
	tabs.OnTabChanged(app.selectSessionTab)

	if !app.canCloseSessionsToRight("runtime-2") {
		t.Fatal("close-right should be available with idle sessions to the right")
	}
	app.closeSessionsToRight("runtime-2")
	states := workspace.Tabs()
	if len(states) != 2 || states[0].ID != "runtime-1" || states[1].ID != "runtime-2" {
		t.Fatalf("close-right retained wrong sessions: %#v", states)
	}
	if tabs.Count() != 2 || workspace.ActiveIndex() != 0 {
		t.Fatalf("close-right drifted native tabs or active session: tabs=%d active=%d", tabs.Count(), workspace.ActiveIndex())
	}

	if !app.canCloseOtherSessions("runtime-2") {
		t.Fatal("close-others should be available with one idle peer")
	}
	app.closeOtherSessions("runtime-2")
	states = workspace.Tabs()
	if len(states) != 1 || states[0].ID != "runtime-2" || workspace.ActiveIndex() != 0 || tabs.ActiveIndex() != 0 {
		t.Fatalf("close-others did not retain and activate target: %#v active=%d native=%d", states, workspace.ActiveIndex(), tabs.ActiveIndex())
	}
}

func TestSessionTabCloseLeftResolvesStableRuntimeAndPreservesTarget(t *testing.T) {
	workspace := newSessionWorkspace()
	for _, id := range []string{"one", "two", "three", "four"} {
		workspace.Open(connectionProfile{ID: id, Name: id, Type: connectionTypeLocal})
	}
	tabs := uikit.NewUITabView(rect(0, 0, 480, 180))
	for _, state := range workspace.Tabs() {
		tabs.AddTabWithID(state.ID, state.Profile.Name, nil)
	}
	tabs.SelectTab(3)
	workspace.Select(3)
	app := &finalShellApp{sessions: workspace, sessionTabs: tabs}
	tabs.OnTabChanged(app.selectSessionTab)

	if !app.canCloseSessionsToLeft("runtime-3") {
		t.Fatal("close-left should be available with idle sessions to the left")
	}
	app.closeSessionsToLeft("runtime-3")
	states := workspace.Tabs()
	if len(states) != 2 || states[0].ID != "runtime-3" || states[1].ID != "runtime-4" {
		t.Fatalf("close-left retained wrong sessions: %#v", states)
	}
	if tabs.Count() != 2 || workspace.ActiveIndex() != 1 || tabs.ActiveIndex() != 1 {
		t.Fatalf("close-left drifted native tabs or active session: tabs=%d workspace=%d native=%d", tabs.Count(), workspace.ActiveIndex(), tabs.ActiveIndex())
	}
}

func TestSessionTabBatchCloseFailsClosedWhenAffectedSessionIsRunning(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "one", Name: "one", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "two", Name: "two", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "three", Name: "three", Type: connectionTypeLocal})
	workspace.Select(1)
	if !workspace.BeginRun("run-two") {
		t.Fatal("failed to mark middle session running")
	}
	workspace.Select(0)
	app := &finalShellApp{sessions: workspace}

	if app.canCloseSessionsToRight("runtime-1") || app.canCloseSessionsToLeft("runtime-3") || app.canCloseOtherSessions("runtime-1") {
		t.Fatal("batch close stayed available across a running affected session")
	}
	app.closeSessionsToRight("runtime-1")
	app.closeSessionsToLeft("runtime-3")
	app.closeOtherSessions("runtime-1")
	if got := len(workspace.Tabs()); got != 3 {
		t.Fatalf("fail-closed batch close partially removed sessions: %d", got)
	}
}

func TestSessionTabBatchCloseSkipsPinnedPeers(t *testing.T) {
	workspace := newSessionWorkspace()
	for _, id := range []string{"one", "two", "three"} {
		workspace.Open(connectionProfile{ID: id, Name: id, Type: connectionTypeLocal})
	}
	if !workspace.SetPinned("runtime-2", true) {
		t.Fatal("failed to pin peer")
	}
	tabs := uikit.NewUITabView(rect(0, 0, 480, 180))
	for _, state := range workspace.Tabs() {
		tabs.AddTabWithID(state.ID, sessionTabTitle(state), nil)
	}
	app := &finalShellApp{sessions: workspace, sessionTabs: tabs}
	tabs.OnTabChanged(app.selectSessionTab)

	app.closeOtherSessions("runtime-1")
	states := workspace.Tabs()
	if len(states) != 2 || states[0].ID != "runtime-2" || !states[0].Pinned || states[1].ID != "runtime-1" {
		t.Fatalf("batch close did not preserve the leading pinned peer: %#v", states)
	}
}

func TestSessionTabContextMenuActionsResolveStableRuntimeIdentity(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "one", Name: "One", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "two", Name: "Two", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "three", Name: "Three", Type: connectionTypeLocal})

	tabs := uikit.NewUITabView(rect(0, 0, 480, 180))
	for _, state := range workspace.Tabs() {
		tabs.AddTabWithID(state.ID, state.Profile.Name, nil)
	}
	tabs.SelectTab(2)
	workspace.Select(2)
	app := &finalShellApp{sessions: workspace, sessionTabs: tabs}
	tabs.OnTabChanged(app.selectSessionTab)

	if !app.activateSessionByID("runtime-1") || workspace.ActiveIndex() != 0 || tabs.ActiveIndex() != 0 {
		t.Fatalf("stable activation drifted: workspace=%d tabs=%d", workspace.ActiveIndex(), tabs.ActiveIndex())
	}
	app.closeSessionByID("runtime-2")
	states := workspace.Tabs()
	if len(states) != 2 || states[0].ID != "runtime-1" || states[1].ID != "runtime-3" {
		t.Fatalf("stable close targeted the wrong runtime: %#v", states)
	}
	if app.activateSessionByID("runtime-2") {
		t.Fatal("removed runtime remained actionable")
	}
}

func TestPinnedSessionStaysVisibleWhileOverflowingSessionsNavigate(t *testing.T) {
	workspace := newSessionWorkspace()
	for _, id := range []string{"one", "two", "three", "four", "five"} {
		workspace.Open(connectionProfile{ID: id, Name: id, Type: connectionTypeLocal})
	}
	tabs := uikit.NewUITabView(rect(0, 0, 500, 180))
	tabs.SetAutomationID("sticky-product-sessions")
	for _, state := range workspace.Tabs() {
		tabs.AddTabWithID(state.ID, state.Profile.Name, nil)
	}
	tabs.SelectTab(4)
	workspace.Select(4)
	app := &finalShellApp{sessions: workspace, sessionTabs: tabs}
	tabs.OnTabChanged(app.selectSessionTab)

	app.togglePinnedSession("runtime-1")
	if !app.activateSessionByID("runtime-5") {
		t.Fatal("failed to activate final overflowing runtime")
	}
	var listed []uikit.TabListItem
	tabs.OnTabListRequested(func(items []uikit.TabListItem) { listed = items })
	if !tabs.RequestTabList() || len(listed) != 5 {
		t.Fatalf("overflow list = %#v", listed)
	}
	for _, item := range listed {
		wantVisible := item.ID == "runtime-1" || item.ID == "runtime-4" || item.ID == "runtime-5"
		if item.Visible != wantVisible {
			t.Fatalf("runtime %s visible=%t, want %t", item.ID, item.Visible, wantVisible)
		}
	}
	states := workspace.Tabs()
	if !states[0].Pinned || states[0].ID != "runtime-1" || tabs.TabID(0) != "runtime-1" || tabs.ActiveIndex() != 4 {
		t.Fatalf("sticky pin drifted product state: states=%#v active=%d", states, tabs.ActiveIndex())
	}
}

func TestTogglePinnedSessionKeepsNativeAndWorkspaceOrderAligned(t *testing.T) {
	workspace := newSessionWorkspace()
	for _, id := range []string{"one", "two", "three"} {
		workspace.Open(connectionProfile{ID: id, Name: id, Type: connectionTypeLocal})
	}
	workspace.Select(0)
	tabs := uikit.NewUITabView(rect(0, 0, 480, 180))
	for _, state := range workspace.Tabs() {
		tabs.AddTabWithID(state.ID, sessionTabTitle(state), nil)
	}
	tabs.SelectTab(0)
	app := &finalShellApp{sessions: workspace, sessionTabs: tabs}
	tabs.OnTabChanged(app.selectSessionTab)

	app.togglePinnedSession("runtime-3")
	states := workspace.Tabs()
	if states[0].ID != "runtime-3" || !states[0].Pinned || tabs.TabID(0) != "runtime-3" || !tabs.TabPinned(0) || tabs.PinnedCount() != 1 {
		t.Fatalf("pin did not align leading native/workspace tab: states=%#v native=%q", states, tabs.TabID(0))
	}
	active, _ := workspace.Active()
	if active.ID != "runtime-1" || tabs.ActiveIndex() != 1 {
		t.Fatalf("pinning background tab changed active identity: active=%#v native=%d", active, tabs.ActiveIndex())
	}
	if tabs.MoveTab(0, 1) || tabs.MoveTab(1, 0) {
		t.Fatal("native tab view allowed a pinned/ordinary partition crossing")
	}

	app.togglePinnedSession("runtime-3")
	states = workspace.Tabs()
	if states[0].ID != "runtime-3" || states[0].Pinned || tabs.TabID(0) != "runtime-3" || tabs.TabPinned(0) || tabs.PinnedCount() != 0 {
		t.Fatalf("unpin did not keep runtime at unpinned boundary: states=%#v native=%q", states, tabs.TabID(0))
	}
}
