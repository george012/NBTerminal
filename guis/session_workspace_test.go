package guis

import (
	"strings"
	"testing"
)

func TestSessionWorkspaceKeepsProfileSelectionSeparateFromActiveTab(t *testing.T) {
	workspace := newSessionWorkspace()
	local := connectionProfile{ID: "local", Name: "Local", Type: connectionTypeLocal}
	server := connectionProfile{ID: "server", Name: "Server", Type: connectionTypeSSH}

	localIndex, created := workspace.Open(local)
	if !created || localIndex != 0 {
		t.Fatalf("unexpected local open: index=%d created=%t", localIndex, created)
	}
	serverIndex, created := workspace.Open(server)
	if !created || serverIndex != 1 {
		t.Fatalf("unexpected server open: index=%d created=%t", serverIndex, created)
	}
	workspace.Select(localIndex)
	workspace.SetProfileSelection(server.ID)

	active, ok := workspace.Active()
	if !ok || active.Profile.ID != local.ID {
		t.Fatalf("profile selection changed active terminal tab: %#v", active)
	}
	if workspace.ProfileSelection() != server.ID {
		t.Fatalf("profile selection not retained separately: %q", workspace.ProfileSelection())
	}
}

func TestSessionWorkspaceCreatesIndependentRuntimeSessionsAndPreservesPerTabState(t *testing.T) {
	workspace := newSessionWorkspace()
	local := connectionProfile{ID: "local", Name: "本地", Type: connectionTypeLocal}
	server := connectionProfile{ID: "server", Name: "Сервер", Type: connectionTypeSSH}

	firstLocalIndex, created := workspace.Open(local)
	if !created || firstLocalIndex != 0 {
		t.Fatalf("first local session was not created: index=%d created=%t", firstLocalIndex, created)
	}
	firstLocal, _ := workspace.Active()
	workspace.SetActiveDraft("printf '简体 · 繁體 · Русский'")
	workspace.AppendActiveOutput("本地输出\n")
	workspace.Open(server)
	workspace.SetActiveDraft("uname -a")
	workspace.AppendActiveOutput("сервер\n")

	secondLocalIndex, created := workspace.Open(local)
	if !created || secondLocalIndex != 2 {
		t.Fatalf("same profile must create an independent runtime session: index=%d created=%t", secondLocalIndex, created)
	}
	secondLocal, _ := workspace.Active()
	if firstLocal.ID == secondLocal.ID || firstLocal.ID == local.ID || secondLocal.ID == local.ID {
		t.Fatalf("runtime IDs must be unique and independent from profile IDs: first=%q second=%q", firstLocal.ID, secondLocal.ID)
	}
	if firstLocal.ProfileID != local.ID || secondLocal.ProfileID != local.ID {
		t.Fatalf("runtime sessions lost source profile identity: first=%#v second=%#v", firstLocal, secondLocal)
	}
	if firstLocal.InstanceNumber != 1 || secondLocal.InstanceNumber != 2 {
		t.Fatalf("unexpected runtime instance numbers: first=%d second=%d", firstLocal.InstanceNumber, secondLocal.InstanceNumber)
	}
	if secondLocal.CommandDraft != "" || strings.Contains(secondLocal.Output, "本地输出\n") {
		t.Fatalf("new runtime session inherited mutable state: %#v", secondLocal)
	}

	workspace.Select(firstLocalIndex)
	active, _ := workspace.Active()
	if active.CommandDraft != "printf '简体 · 繁體 · Русский'" || !strings.Contains(active.Output, "本地输出\n") {
		t.Fatalf("first local runtime state was not preserved: %#v", active)
	}
	if len(workspace.Tabs()) != 3 {
		t.Fatalf("expected three independent runtime sessions: %#v", workspace.Tabs())
	}
}

func TestSessionWorkspaceRunningLifecycleAndClosePolicy(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "one", Name: "One", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "two", Name: "Two", Type: connectionTypeLocal})
	workspace.Select(0)

	if !workspace.BeginRun("run-1") {
		t.Fatal("BeginRun failed")
	}
	if workspace.Close(0) {
		t.Fatal("running session tab must not close")
	}
	if !workspace.FinishRun("run-1", sessionSucceeded) {
		t.Fatal("FinishRun failed")
	}
	if !workspace.Close(0) {
		t.Fatal("idle completed session should close")
	}
	active, ok := workspace.Active()
	if !ok || active.Profile.ID != "two" || workspace.ActiveIndex() != 0 {
		t.Fatalf("nearest session was not activated after close: active=%#v index=%d", active, workspace.ActiveIndex())
	}
}

func TestSessionWorkspaceRejectsStaleCompletion(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "local", Name: "Local", Type: connectionTypeLocal})
	if !workspace.BeginRun("new-run") {
		t.Fatal("BeginRun failed")
	}
	if workspace.FinishRun("stale-run", sessionFailed) {
		t.Fatal("stale completion mutated active session")
	}
	active, _ := workspace.Active()
	if active.Status != sessionRunning || active.RunID != "new-run" {
		t.Fatalf("running state was corrupted: %#v", active)
	}
}

func TestSessionWorkspaceRuntimeIdentitySurvivesEarlierTabClose(t *testing.T) {
	workspace := newSessionWorkspace()
	profile := connectionProfile{ID: "same-profile", Name: "Server", Type: connectionTypeSSH}
	workspace.Open(profile)
	workspace.Open(profile)
	if !workspace.Close(0) {
		t.Fatal("failed to close first runtime session")
	}
	workspace.Open(profile)
	tabs := workspace.Tabs()
	if len(tabs) != 2 || tabs[0].InstanceNumber != 2 || tabs[1].InstanceNumber != 3 {
		t.Fatalf("runtime instance numbers were reused after close: %#v", tabs)
	}
	if tabs[0].ID == tabs[1].ID {
		t.Fatalf("runtime ID was reused after close: %#v", tabs)
	}
	if got := sessionTabTitle(tabs[1]); got != "Server · 3" {
		t.Fatalf("duplicate profile tab title = %q, want %q", got, "Server · 3")
	}
}

func TestSessionWorkspaceReopensMostRecentlyClosedProfileAsFreshRuntime(t *testing.T) {
	workspace := newSessionWorkspace()
	first := connectionProfile{ID: "first", Name: "First", Type: connectionTypeLocal}
	second := connectionProfile{ID: "second", Name: "Second", Type: connectionTypeSSH}
	workspace.Open(first)
	firstRuntime, _ := workspace.Active()
	workspace.Open(second)
	secondRuntime, _ := workspace.Active()

	if !workspace.Close(workspace.ActiveIndex()) || !workspace.Close(workspace.ActiveIndex()) {
		t.Fatal("failed to close runtime sessions")
	}
	if _, ok := workspace.Active(); ok {
		t.Fatal("workspace should be empty before reopen")
	}

	index, reopened := workspace.ReopenLastClosed()
	if !reopened || index != 0 {
		t.Fatalf("reopen latest profile: index=%d reopened=%t", index, reopened)
	}
	active, _ := workspace.Active()
	if active.ProfileID != first.ID || active.ID == firstRuntime.ID || active.CommandDraft != "" {
		t.Fatalf("reopened session is not a fresh runtime for the latest profile: %#v", active)
	}
	if !workspace.Close(index) {
		t.Fatal("failed to close reopened profile")
	}
	if _, reopened = workspace.ReopenLastClosed(); !reopened {
		t.Fatal("failed to reopen the profile a second time")
	}
	active, _ = workspace.Active()
	if active.ProfileID != first.ID || active.ID == secondRuntime.ID {
		t.Fatalf("closing a reopened tab did not update the LIFO stack: %#v", active)
	}
}

func TestSessionWorkspaceReopenRejectsEmptyHistory(t *testing.T) {
	workspace := newSessionWorkspace()
	if workspace.CanReopenClosed() {
		t.Fatal("empty workspace reported reopen history")
	}
	if index, ok := workspace.ReopenLastClosed(); ok || index != -1 {
		t.Fatalf("empty reopen = index %d, ok %t", index, ok)
	}
}

func TestSessionWorkspacePinnedRuntimeRejectsCloseUntilUnpinned(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "production", Name: "Production", Type: connectionTypeLocal})
	state, _ := workspace.Active()

	if !workspace.SetPinned(state.ID, true) {
		t.Fatal("failed to pin runtime")
	}
	if workspace.Close(0) {
		t.Fatal("pinned runtime was closed")
	}
	if got := workspace.Tabs(); len(got) != 1 || !got[0].Pinned {
		t.Fatalf("pinned state was not preserved: %#v", got)
	}
	if got := sessionTabTitle(workspace.Tabs()[0]); got != "◆ Production" {
		t.Fatalf("pinned tab title = %q", got)
	}
	if !workspace.SetPinned(state.ID, false) || !workspace.Close(0) {
		t.Fatal("unpin did not restore normal close behavior")
	}
}

func TestSessionWorkspaceGroupsPinnedRuntimesAtLeadingEdge(t *testing.T) {
	workspace := newSessionWorkspace()
	for _, id := range []string{"one", "two", "three", "four"} {
		workspace.Open(connectionProfile{ID: id, Name: id, Type: connectionTypeLocal})
	}
	workspace.Select(0)

	if !workspace.SetPinned("runtime-3", true) || !workspace.SetPinned("runtime-2", true) {
		t.Fatal("failed to pin runtimes")
	}
	states := workspace.Tabs()
	if got := []string{states[0].ID, states[1].ID, states[2].ID, states[3].ID}; strings.Join(got, ",") != "runtime-3,runtime-2,runtime-1,runtime-4" {
		t.Fatalf("pinned order = %v, want stable leading pin group", got)
	}
	active, _ := workspace.Active()
	if active.ID != "runtime-1" {
		t.Fatalf("pinning background runtimes changed active identity: %#v", active)
	}

	if !workspace.SetPinned("runtime-3", false) {
		t.Fatal("failed to unpin runtime")
	}
	states = workspace.Tabs()
	if got := []string{states[0].ID, states[1].ID, states[2].ID, states[3].ID}; strings.Join(got, ",") != "runtime-2,runtime-3,runtime-1,runtime-4" {
		t.Fatalf("unpin order = %v, want runtime at unpinned boundary", got)
	}
	active, _ = workspace.Active()
	if active.ID != "runtime-1" {
		t.Fatalf("unpinning background runtime changed active identity: %#v", active)
	}
}

func TestSessionWorkspaceMoveActiveCannotCrossPinnedBoundary(t *testing.T) {
	workspace := newSessionWorkspace()
	for _, id := range []string{"one", "two", "three"} {
		workspace.Open(connectionProfile{ID: id, Name: id, Type: connectionTypeLocal})
	}
	if !workspace.SetPinned("runtime-2", true) {
		t.Fatal("failed to pin runtime")
	}

	workspace.Select(0)
	if workspace.MoveActive(1) {
		t.Fatal("pinned runtime crossed into the unpinned group")
	}
	workspace.Select(1)
	if workspace.MoveActive(-1) {
		t.Fatal("unpinned runtime crossed into the pinned group")
	}
	if got := workspace.Tabs(); got[0].ID != "runtime-2" || !got[0].Pinned || got[1].ID != "runtime-1" {
		t.Fatalf("rejected boundary moves changed runtime order: %#v", got)
	}
}

func TestSessionWorkspaceReportsReopenAvailabilityAcrossCloseAndReopen(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "local", Name: "Local", Type: connectionTypeLocal})
	if workspace.CanReopenClosed() {
		t.Fatal("open workspace reported reopen history before any close")
	}
	if !workspace.Close(0) || !workspace.CanReopenClosed() {
		t.Fatal("closing an idle runtime did not expose reopen history")
	}
	if _, ok := workspace.ReopenLastClosed(); !ok || workspace.CanReopenClosed() {
		t.Fatal("reopening the only closed runtime did not consume reopen history")
	}
}

func TestSessionWorkspaceDuplicatesActiveProfileAsFreshRuntime(t *testing.T) {
	workspace := newSessionWorkspace()
	profile := connectionProfile{ID: "local", Name: "Local Shell", Type: connectionTypeLocal}
	workspace.Open(profile)
	workspace.SetActiveDraft("do not copy")
	workspace.AppendActiveOutput("private terminal state")
	original, _ := workspace.Active()

	index, duplicated := workspace.DuplicateActive()
	if !duplicated || index != 1 {
		t.Fatalf("duplicate active = index %d, ok %t", index, duplicated)
	}
	duplicate, ok := workspace.Active()
	if !ok || duplicate.ID == original.ID || duplicate.ProfileID != profile.ID || duplicate.InstanceNumber != 2 {
		t.Fatalf("duplicate runtime identity = %#v, original %#v", duplicate, original)
	}
	if duplicate.CommandDraft != "" || strings.Contains(duplicate.Output, "private terminal state") {
		t.Fatalf("duplicate inherited mutable terminal state: %#v", duplicate)
	}
}

func TestSessionWorkspaceDuplicatesLocalRuntimeAtReportedWorkingDirectory(t *testing.T) {
	workspace := newSessionWorkspace()
	profile := connectionProfile{ID: "local", Name: "Local Shell", Type: connectionTypeLocal, WorkingDir: "/saved"}
	workspace.Open(profile)
	original, _ := workspace.Active()
	if !workspace.SetWorkingDirectory(original.ID, "/home/user/current") {
		t.Fatal("failed to retain runtime working directory")
	}

	if _, duplicated := workspace.DuplicateActive(); !duplicated {
		t.Fatal("failed to duplicate active runtime")
	}
	duplicate, _ := workspace.Active()
	if duplicate.Profile.WorkingDir != "/home/user/current" {
		t.Fatalf("duplicate working directory = %q, want reported current directory", duplicate.Profile.WorkingDir)
	}
	if duplicate.CurrentDirectory != "/home/user/current" {
		t.Fatalf("duplicate did not expose its intentional starting directory: %#v", duplicate)
	}
	if workspace.Tabs()[0].Profile.WorkingDir != "/saved" {
		t.Fatal("runtime metadata mutated the saved profile snapshot")
	}
}

func TestSessionWorkspaceRejectsWorkingDirectoryForUnknownRuntime(t *testing.T) {
	workspace := newSessionWorkspace()
	if workspace.SetWorkingDirectory("missing", "/tmp") {
		t.Fatal("unknown runtime accepted working-directory metadata")
	}
}

func TestSessionWorkspaceDuplicateRejectsEmptyWorkspace(t *testing.T) {
	workspace := newSessionWorkspace()
	if index, ok := workspace.DuplicateActive(); ok || index != -1 {
		t.Fatalf("empty duplicate = index %d, ok %t", index, ok)
	}
}

func TestSessionWorkspaceMovesActiveTabWithoutLosingRuntimeState(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "one", Name: "One", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "two", Name: "Two", Type: connectionTypeLocal})
	workspace.SetActiveDraft("keep this draft")
	workspace.AppendActiveOutput("keep this output")
	active, _ := workspace.Active()

	if !workspace.MoveActive(-1) {
		t.Fatal("MoveActive(-1) failed")
	}
	states := workspace.Tabs()
	if workspace.ActiveIndex() != 0 || states[0].ID != active.ID || states[1].ProfileID != "one" {
		t.Fatalf("runtime order or active identity drifted: active=%d states=%#v", workspace.ActiveIndex(), states)
	}
	if states[0].CommandDraft != "keep this draft" || !strings.Contains(states[0].Output, "keep this output") {
		t.Fatalf("moving active runtime lost tab-local state: %#v", states[0])
	}
	if workspace.MoveActive(-1) {
		t.Fatal("moving beyond the leading boundary should be a no-op")
	}
	if !workspace.MoveActive(1) || workspace.ActiveIndex() != 1 || workspace.Tabs()[1].ID != active.ID {
		t.Fatal("moving active runtime right did not restore its order")
	}
}

func TestSessionWorkspaceMoveActiveRejectsEmptyAndUnsupportedDelta(t *testing.T) {
	workspace := newSessionWorkspace()
	if workspace.MoveActive(1) {
		t.Fatal("empty workspace accepted a move")
	}
	workspace.Open(connectionProfile{ID: "only", Name: "Only", Type: connectionTypeLocal})
	if workspace.MoveActive(0) || workspace.MoveActive(2) || workspace.ActiveIndex() != 0 {
		t.Fatal("unsupported move changed a single-tab workspace")
	}
}

func TestSessionWorkspaceMovesRuntimeByStableIdentityWithoutChangingSelection(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "one", Name: "One", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "two", Name: "Two", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "three", Name: "Three", Type: connectionTypeLocal})
	workspace.Select(2)
	active, _ := workspace.Active()

	if !workspace.Move("runtime-1", 1) {
		t.Fatal("stable runtime move failed")
	}
	states := workspace.Tabs()
	if states[0].ID != "runtime-2" || states[1].ID != "runtime-1" || states[2].ID != "runtime-3" {
		t.Fatalf("runtime order = %#v", states)
	}
	if selected, _ := workspace.Active(); selected.ID != active.ID || workspace.ActiveIndex() != 2 {
		t.Fatalf("active identity drifted: index=%d state=%#v", workspace.ActiveIndex(), selected)
	}
}

func TestSessionWorkspaceMoveRejectsPinnedBoundaryAndUnknownRuntime(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "one", Name: "One", Type: connectionTypeLocal})
	workspace.Open(connectionProfile{ID: "two", Name: "Two", Type: connectionTypeLocal})
	if !workspace.SetPinned("runtime-1", true) {
		t.Fatal("failed to pin runtime")
	}
	if workspace.Move("runtime-2", 0) || workspace.Move("missing", 1) || workspace.Move("runtime-2", 1) {
		t.Fatal("invalid runtime move was accepted")
	}
}

func TestSessionWorkspaceKeepsTerminalTitlesBoundToRuntimeIdentity(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "local", Name: "Local Shell", Type: connectionTypeLocal})
	first, _ := workspace.Active()
	workspace.Open(connectionProfile{ID: "remote", Name: "Production", Type: connectionTypeSSH})
	second, _ := workspace.Active()

	if !workspace.SetTerminalTitle(first.ID, "~/NBTerminal") {
		t.Fatal("failed to set background runtime title")
	}
	if !workspace.SetTerminalTitle(second.ID, "deploy@prod") {
		t.Fatal("failed to set active runtime title")
	}
	tabs := workspace.Tabs()
	if got := sessionTabTitle(tabs[0]); got != "~/NBTerminal" {
		t.Fatalf("first runtime title = %q, want OSC title", got)
	}
	if got := sessionTabTitle(tabs[1]); got != "deploy@prod" {
		t.Fatalf("second runtime title = %q, want OSC title", got)
	}
	if workspace.SetTerminalTitle("missing", "spoof") {
		t.Fatal("unknown runtime accepted a title update")
	}
}

func TestSessionWorkspaceCustomTitleOverridesShellTitleWithoutMutatingProfile(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "prod", Name: "Production", Type: connectionTypeSSH})
	state, _ := workspace.Active()
	if !workspace.SetTerminalTitle(state.ID, "deploy@prod") || !workspace.SetCustomTitle(state.ID, "Incident bridge") {
		t.Fatal("failed to set runtime titles")
	}
	state, _ = workspace.Active()
	if got := sessionTabTitle(state); got != "Incident bridge" {
		t.Fatalf("custom tab title = %q", got)
	}
	if state.Profile.Name != "Production" || state.TerminalTitle != "deploy@prod" {
		t.Fatalf("custom title mutated profile or shell metadata: %#v", state)
	}
	if !workspace.SetCustomTitle(state.ID, "") {
		t.Fatal("failed to reset custom title")
	}
	state, _ = workspace.Active()
	if got := sessionTabTitle(state); got != "deploy@prod" {
		t.Fatalf("reset title = %q, want shell title", got)
	}
	if workspace.SetCustomTitle("missing", "spoof") {
		t.Fatal("unknown runtime accepted a custom title")
	}
}

func TestSessionWorkspaceKeepsBellAttentionUntilSessionIsViewed(t *testing.T) {
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "local", Name: "Local Shell", Type: connectionTypeLocal})
	first, _ := workspace.Active()
	workspace.Open(connectionProfile{ID: "remote", Name: "Production", Type: connectionTypeSSH})

	if !workspace.SetAttention(first.ID) {
		t.Fatal("failed to mark background runtime for attention")
	}
	if got := sessionTabTitle(workspace.Tabs()[0]); got != "● Local Shell" {
		t.Fatalf("attention tab title = %q, want visible marker", got)
	}
	if !workspace.Select(0) {
		t.Fatal("failed to select marked runtime")
	}
	if workspace.Tabs()[0].NeedsAttention {
		t.Fatal("viewing runtime did not clear attention")
	}
	if got := sessionTabTitle(workspace.Tabs()[0]); got != "Local Shell" {
		t.Fatalf("viewed tab title = %q, want cleared marker", got)
	}
	if workspace.SetAttention("missing") {
		t.Fatal("unknown runtime accepted attention")
	}
}

func TestQuickLocalSessionProfileStartsAtHomeWithoutPersistedSecrets(t *testing.T) {
	profile := quickLocalSessionProfile(" /home/tester ")
	if profile.ID != quickLocalSessionProfileID || profile.Name != "Local Shell" || profile.Type != connectionTypeLocal {
		t.Fatalf("unexpected quick local profile: %#v", profile)
	}
	if profile.WorkingDir != "/home/tester" {
		t.Fatalf("working directory = %q, want home", profile.WorkingDir)
	}
	if profile.PasswordEnc != "" || profile.PrivateKey != "" || profile.Host != "" || profile.Username != "" {
		t.Fatalf("ephemeral local profile retained remote credentials: %#v", profile)
	}

	workspace := newSessionWorkspace()
	workspace.Open(profile)
	workspace.Open(profile)
	tabs := workspace.Tabs()
	if len(tabs) != 2 || tabs[0].ProfileID != quickLocalSessionProfileID || tabs[1].InstanceNumber != 2 {
		t.Fatalf("quick local sessions are not independent runtimes: %#v", tabs)
	}
}
