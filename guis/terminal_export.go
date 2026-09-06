package guis

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xdevelop/NBTerminal/internal/persistence"
	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

func writeTerminalOutput(path, output string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("export path is required")
	}
	if output == "" {
		return errors.New("terminal output is empty")
	}
	return persistence.AtomicWriteFile(path, []byte(output), 0o600)
}

func terminalOutputExportName(now time.Time) string {
	return fmt.Sprintf("nbterminal-%s.txt", now.Local().Format("20060102-150405"))
}

func (a *finalShellApp) exportTerminalOutput() {
	if a == nil || a.output == nil {
		return
	}
	output := a.output.Text()
	if output == "" {
		a.setStatus("Terminal output is empty")
		return
	}

	chooser := fltk_bridge.NewNativeFileChooser()
	defer chooser.Destroy()
	chooser.SetTitle("Save Terminal Output")
	chooser.SetType(fltk_bridge.NativeFileChooser_BROWSE_SAVE_FILE)
	chooser.SetFilter("Text files\t*.txt\nAll files\t*")
	chooser.SetOptions(fltk_bridge.NativeFileChooser_SAVEAS_CONFIRM | fltk_bridge.NativeFileChooser_USE_FILTER_EXT)
	chooser.SetPresetFile(terminalOutputExportName(time.Now()))
	if chooser.Show() != 0 {
		return
	}
	filenames := chooser.Filenames()
	if len(filenames) == 0 || strings.TrimSpace(filenames[0]) == "" {
		return
	}
	path := filenames[0]
	if err := writeTerminalOutput(path, output); err != nil {
		a.setStatus("Could not save terminal output")
		a.showTopNotice("Save Terminal Output", err.Error(), true)
		return
	}
	// Keep the full destination out of the compact status strip while still
	// confirming which user-visible file was written.
	a.setStatus(fmt.Sprintf("Saved terminal output to %s", filepath.Base(path)))
}
