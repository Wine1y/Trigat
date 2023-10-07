package main

import (
	"runtime"

	"github.com/Wine1y/trigat/gui"
	scWindow "github.com/Wine1y/trigat/gui/sc_window"
	"github.com/Wine1y/trigat/hotkeys"
	hk "golang.design/x/hotkey"
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

func main() {
	app := App{}
	stopChan := make(chan bool)

	screenshotCb := func() {
		screenshotWindow := scWindow.NewScreenshotWindow()
		app.OpenWindow(screenshotWindow)
		screenshotWindow = nil
		runtime.GC()
	}
	exitCb := func() {
		stopChan <- true
	}

	screenshotHk := hotkeys.NewHotKey(hk.KeyS, nil, &screenshotCb, nil)
	exitHk := hotkeys.NewHotKey(hk.KeyQ, nil, &exitCb, nil)
	defaultHotKeys := hotkeys.NewHotKeySet(screenshotHk, exitHk)

	app.Start(defaultHotKeys)
	println("App started")
	<-stopChan
}
