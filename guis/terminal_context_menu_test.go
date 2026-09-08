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
	items := sessionTabContextMenuItems(false, true, true,
		func() { invocations["activate"]++ },
		func() { invocations["duplicate"]++ },
		func() { invocations["reconnect"]++ },
		func() { invocations["close"]++ },
	)
	if len(items) != 4 {
		t.Fatalf("session context menu item count = %d, want 4", len(items))
	}
	for index, title := range []string{"Activate Session", "Duplicate Session	Ctrl+Shift+D", "Reconnect Session	Ctrl+Shift+R", "Close Session	Ctrl+W"} {
		if items[index].Title != title || items[index].Flags&fltk_bridge.MENU_INACTIVE != 0 {
			t.Fatalf("session item %d = %#v", index, items[index])
		}
		items[index].Callback()
	}
	for _, action := range []string{"activate", "duplicate", "reconnect", "close"} {
		if invocations[action] != 1 {
			t.Fatalf("%s callback count = %d, want 1", action, invocations[action])
		}
	}

	items = sessionTabContextMenuItems(true, false, false, func() {}, func() {}, func() {}, func() {})
	if items[0].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatal("active session kept Activate enabled")
	}
	if items[2].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatal("non-interactive session kept Reconnect enabled")
	}
	if items[3].Flags&fltk_bridge.MENU_INACTIVE == 0 {
		t.Fatal("running session kept Close enabled")
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
