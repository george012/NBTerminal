package config

import (
	"github.com/george012/gtbox"
	"github.com/george012/gtbox/gtbox_app"
)

var (
	CurrentApp *ExtendApp
)

type ExtendApp struct {
	*gtbox_app.App
	NetListenAPIPortDefault int
	DataDir                 string
}

func NewApp(appName, bundleID, description string, runMode gtbox.RunMode, apiPort int) *ExtendApp {
	app := &ExtendApp{
		App:                     gtbox_app.NewApp(appName, ProjectVersion, bundleID, description, runMode),
		NetListenAPIPortDefault: apiPort,
		DataDir:                 "./dts",
	}

	return app
}
