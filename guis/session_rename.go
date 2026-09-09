package guis

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit"
)

const (
	sessionRenameWidth    = 520
	sessionRenameHeight   = 238
	sessionCustomTitleMax = 64
)

type sessionRenameWindow struct {
	owner     *finalShellApp
	window    *uikit.UIWindow
	sessionID string
	name      *uikit.Input
	status    *uikit.UILabel
}

func normalizeSessionCustomTitle(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > sessionCustomTitleMax {
		return "", fmt.Errorf("session name must be %d characters or fewer", sessionCustomTitleMax)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("session name must be a single line")
		}
	}
	return value, nil
}

func (a *finalShellApp) openSessionRename(sessionID string) {
	if a == nil || a.sessionIndexByID(sessionID) < 0 {
		return
	}
	if current := a.sessionRename; current != nil && current.window != nil && !current.window.IsClosed() {
		if current.sessionID == sessionID {
			current.window.Show()
			current.focusName()
			return
		}
		current.window.Close()
		fltk_bridge.AddTimeout(0, func() { a.openSessionRename(sessionID) })
		return
	}
	dialog := &sessionRenameWindow{owner: a, sessionID: sessionID}
	a.sessionRename = dialog
	dialog.build()
}

func (d *sessionRenameWindow) build() {
	if d == nil || d.owner == nil {
		return
	}
	index := d.owner.sessionIndexByID(d.sessionID)
	if index < 0 {
		return
	}
	state := d.owner.sessions.Tabs()[index]
	windowRect := centeredScreenRect(sessionRenameWidth, sessionRenameHeight)
	if d.owner.window != nil && d.owner.window.Raw() != nil {
		raw := d.owner.window.Raw()
		windowRect = rect(raw.XRoot()+(raw.W()-sessionRenameWidth)/2, raw.YRoot()+(raw.H()-sessionRenameHeight)/2, sessionRenameWidth, sessionRenameHeight)
	}
	d.window = uikit.NewWindowWithRect(windowRect, "Rename Session")
	d.window.SetResizable(false)
	if raw := d.window.Raw(); raw != nil {
		raw.SetXClass(nativeWindowClass())
		raw.SetNonModal()
		raw.SetColor(tokenColor(modernTheme.background))
	}
	root := d.window.RootView()
	root.SetAutomationID("session_rename.window").SetAutomationRole("window").SetAutomationName("Rename Session")
	d.window.OnClose(func() {
		root.SetAutomationID("")
		if d.owner != nil && d.owner.sessionRename == d {
			d.owner.sessionRename = nil
		}
	})

	root.AddSubview(titleLabel(28, 20, 464, nativeControls.WindowTitleHeight, "Rename Session"))
	subtitle := mutedLabel(28, 54, 464, nativeControls.SupportingLineHeight*2, "Set a temporary label for this tab. Leave it blank to follow the shell title.")
	subtitle.SetAlignment(fltk_bridge.ALIGN_LEFT | fltk_bridge.ALIGN_INSIDE | fltk_bridge.ALIGN_WRAP)
	root.AddSubview(subtitle)
	d.name = inputNoLabel(28, 104, 464, nativeControls.InputHeight, "session_rename.name", "Session name")
	d.name.SetText(state.CustomTitle)
	root.AddSubview(d.name)
	d.status = mutedLabel(28, 146, 464, nativeControls.SupportingLineHeight, "Runtime only · saved connection unchanged")
	styleDynamicLabel(d.status)
	d.status.View().SetAutomationID("session_rename.status")
	root.AddSubview(d.status)
	root.AddSubview(button(286, 180, 96, nativeControls.PrimaryButtonHeight, "Cancel", "session_rename.cancel", d.close))
	root.AddSubview(primaryButton(392, 180, 100, nativeControls.PrimaryButtonHeight, "Rename", "session_rename.save", d.save))
	d.window.Show()
	d.focusName()
}

func (d *sessionRenameWindow) focusName() {
	if d == nil || d.name == nil || d.name.View() == nil {
		return
	}
	if raw := d.name.View().Raw(); raw != nil {
		if focusable, ok := raw.(interface{ TakeFocus() int }); ok {
			focusable.TakeFocus()
		}
	}
}

func (d *sessionRenameWindow) save() {
	if d == nil || d.owner == nil || d.name == nil {
		return
	}
	title, err := normalizeSessionCustomTitle(d.name.Text())
	if err != nil {
		d.status.SetText(err.Error())
		return
	}
	index := d.owner.sessionIndexByID(d.sessionID)
	if index < 0 || !d.owner.sessions.SetCustomTitle(d.sessionID, title) {
		d.status.SetText("This session is no longer available.")
		return
	}
	state := d.owner.sessions.Tabs()[index]
	d.owner.refreshSessionTabs()
	if title == "" {
		d.owner.setStatus(fmt.Sprintf("Following shell title for %s", state.Profile.Name))
	} else {
		d.owner.setStatus(fmt.Sprintf("Renamed session to %s", title))
	}
	d.close()
}

func (d *sessionRenameWindow) close() {
	if d != nil && d.window != nil {
		d.window.Close()
	}
}
