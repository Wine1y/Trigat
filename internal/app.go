package internal

import (
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/pkg/hotkeys"
)

type App struct {
	currentWindow  gui.Window
	defaultHotKeys *hotkeys.HotKeySet
	currentHotKeys *hotkeys.HotKeySet
}

func (app *App) Start(defaultHotKeys *hotkeys.HotKeySet) {
	app.defaultHotKeys = defaultHotKeys
	app.setHotkeys(app.defaultHotKeys)
}

func (app *App) OpenWindow(window gui.Window) {
	if app.currentWindow != nil {
		app.currentWindow.Close()
	}
	app.currentWindow = window
	app.setHotkeys(window.HotKeys())
	app.currentWindow.StartMainLoop()
	app.setHotkeys(app.defaultHotKeys)
	app.currentWindow = nil
}

func (app *App) setHotkeys(hotkeys *hotkeys.HotKeySet) {
	if app.currentHotKeys != nil {
		if err := app.currentHotKeys.StopListeningAll(); err != nil {
			panic("Can't deactivate hotkeys")
		}
	}
	if err := hotkeys.StartListeningAll(); err != nil {
		panic("Can't activate hotkeys")
	}
	app.currentHotKeys = hotkeys
}
