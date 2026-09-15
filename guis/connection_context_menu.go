package guis

import (
	"time"

	"github.com/0xdevelop/fltk2go/uikit"
)

type connectionContextMenuActions struct {
	connect, edit, duplicate, test, favorite, delete func()
}

func connectionContextMenuItems(favorite bool, actions connectionContextMenuActions) []uikit.MenuItem {
	favoriteTitle := "Add to Favorites"
	if favorite {
		favoriteTitle = "Remove from Favorites"
	}
	return []uikit.MenuItem{
		{Title: "Connect", Callback: actions.connect},
		{Title: "Edit…", Callback: actions.edit},
		{Title: "Duplicate…", Callback: actions.duplicate},
		{Title: "Test Connection", Callback: actions.test},
		{Title: favoriteTitle, Callback: actions.favorite},
		{Title: "Delete…", Callback: actions.delete},
	}
}

// selectContextProfile resolves a menu's stable profile identity at action
// time. Search, group filters, persistence refreshes, or other windows may have
// changed row indexes while the native menu was open.
func (m *connectionManagerWindow) selectContextProfile(id string) bool {
	if m == nil {
		return false
	}
	index := indexProfileByID(m.rows, id)
	if index < 0 {
		return false
	}
	if m.table != nil {
		return m.table.SelectRow(index)
	}
	m.idx = index
	return true
}

func (m *connectionManagerWindow) installContextMenu(parent interface{ AddSubview(viewable uikit.Viewable) }) {
	if m == nil || m.table == nil || parent == nil {
		return
	}
	menu := uikit.NewUIContextMenu(rect(0, 0, 0, 0))
	parent.AddSubview(menu)
	m.contextMenu = menu
	m.table.OnContextMenu(func(request uikit.TableContextMenuState) {
		if request.Row < 0 || request.Row >= len(m.rows) {
			return
		}
		profile := m.rows[request.Row]
		run := func(action func()) func() {
			return func() {
				if m.selectContextProfile(profile.ID) && action != nil {
					action()
				}
			}
		}
		menu.SetMenu(connectionContextMenuItems(profile.Favorite, connectionContextMenuActions{
			connect:   run(m.connectSelected),
			edit:      run(m.editSelected),
			duplicate: run(m.duplicateSelected),
			test:      run(m.testSelected),
			favorite:  run(m.toggleFavorite),
			delete:    run(m.deleteSelected),
		}))
		menu.Popup()
	})
}

// selectQuickContextProfile resolves the profile again when the menu command
// runs. Search changes can reorder the compact launcher while its native popup
// is open, so the original row index is not authoritative.
func (a *finalShellApp) selectQuickContextProfile(id string) bool {
	if a == nil {
		return false
	}
	index := indexProfileByID(a.rows, id)
	if index < 0 {
		return false
	}
	if a.table != nil {
		return a.table.SelectRow(index)
	}
	a.idx = index
	return true
}

func (a *finalShellApp) duplicateSelectedProfile() {
	profile, ok := a.selectedProfile()
	if !ok {
		return
	}
	a.openConnectionEditor(duplicateConnectionProfile(profile, a.allRows, time.Now().UnixNano()))
}

func (a *finalShellApp) testSelectedProfile() {
	if profile, ok := a.selectedProfile(); ok {
		a.testProfile(profile)
	}
}

func (a *finalShellApp) toggleSelectedProfileFavorite() {
	profile, ok := a.selectedProfile()
	if !ok {
		return
	}
	profile = toggledFavorite(profile)
	if err := a.persistProfile(profile); err != nil {
		a.setStatus(tr("status.save_failed"))
		a.showTopNotice(tr("status.save_failed"), err.Error(), true)
		return
	}
	a.refreshTable()
	a.updateSelectedSummary()
	a.setStatus(trf("manager.favorite_updated", profile.Name))
	if a.manager != nil {
		a.manager.reload(profile.ID)
	}
}

func (a *finalShellApp) installQuickConnectionContextMenu(parent interface{ AddSubview(viewable uikit.Viewable) }) {
	if a == nil || a.table == nil || parent == nil {
		return
	}
	menu := uikit.NewUIContextMenu(rect(0, 0, 0, 0))
	parent.AddSubview(menu)
	a.connectionContextMenu = menu
	a.table.OnContextMenu(func(request uikit.TableContextMenuState) {
		if request.Row < 0 || request.Row >= len(a.rows) {
			return
		}
		profile := a.rows[request.Row]
		run := func(action func()) func() {
			return func() {
				if a.selectQuickContextProfile(profile.ID) && action != nil {
					action()
				}
			}
		}
		menu.SetMenu(connectionContextMenuItems(profile.Favorite, connectionContextMenuActions{
			connect:   run(a.connectSelected),
			edit:      run(a.editSelectedProfile),
			duplicate: run(a.duplicateSelectedProfile),
			test:      run(a.testSelectedProfile),
			favorite:  run(a.toggleSelectedProfileFavorite),
			delete:    run(a.deleteProfile),
		}))
		menu.Popup()
	})
}
