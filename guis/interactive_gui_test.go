package guis

import (
	"os"
	"testing"

	"github.com/0xdevelop/fltk2go/uikit"
)

func TestWorkingDirectoryHostMatchesTransport(t *testing.T) {
	localHost, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	local := connectionProfile{Type: connectionTypeLocal}
	remote := connectionProfile{Type: connectionTypeSSH, Host: "server.example.com"}

	for _, test := range []struct {
		name    string
		profile connectionProfile
		host    string
		want    bool
	}{
		{name: "empty local host", profile: local, host: "", want: true},
		{name: "local hostname", profile: local, host: localHost, want: true},
		{name: "localhost", profile: local, host: "localhost", want: true},
		{name: "foreign local host", profile: local, host: "remote.invalid", want: false},
		{name: "remote fqdn", profile: remote, host: "server.example.com", want: true},
		{name: "remote short hostname", profile: remote, host: "server", want: true},
		{name: "remote runtime hostname", profile: remote, host: "worker-42.internal", want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := workingDirectoryHostMatches(test.profile, test.host); got != test.want {
				t.Fatalf("workingDirectoryHostMatches(%#v, %q) = %t, want %t", test.profile, test.host, got, test.want)
			}
		})
	}
	if hostNamesEquivalent("192.0.2.1", "192.example.com") {
		t.Fatal("IP address was incorrectly matched by its first label")
	}
}

func TestActiveTerminalWorkingDirectoryUpdatesRuntimeAndVisibleSubtitle(t *testing.T) {
	directory := t.TempDir()
	workspace := newSessionWorkspace()
	workspace.Open(connectionProfile{ID: "local", Name: "Local", Type: connectionTypeLocal, WorkingDir: "/saved"})
	app := &finalShellApp{
		sessions:         workspace,
		terminalSubtitle: mutedLabel(0, 0, 500, 24, "initial"),
	}

	app.activeTerminalWorkingDirectoryChanged(uikit.TerminalWorkingDirectory{Path: directory})
	active, _ := workspace.Active()
	if active.CurrentDirectory != directory {
		t.Fatalf("runtime current directory = %q, want %q", active.CurrentDirectory, directory)
	}
	if got := terminalSubtitleText(active); got != "Current directory: "+directory {
		t.Fatalf("visible subtitle = %q", got)
	}

	app.activeTerminalWorkingDirectoryChanged(uikit.TerminalWorkingDirectory{Path: directory + "/missing"})
	active, _ = workspace.Active()
	if active.CurrentDirectory != directory {
		t.Fatal("missing local directory replaced valid runtime metadata")
	}
}
