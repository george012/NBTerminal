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
