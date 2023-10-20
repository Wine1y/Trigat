package main

import (
	"runtime"

	"github.com/Wine1y/trigat/internal"
	scWindow "github.com/Wine1y/trigat/internal/gui/sc_window"
	"github.com/Wine1y/trigat/pkg/hotkeys"
	hk "golang.design/x/hotkey"
)

func main() {
	app := internal.NewApp()

	screenshotCb := func() {
		screenshotWindow := scWindow.NewScreenshotWindow()
		app.OpenWindow(screenshotWindow)
		screenshotWindow = nil
		runtime.GC()
	}

	screenshotHk := hotkeys.NewHotKey(hk.KeyS, nil, &screenshotCb, nil)
	defaultHotKeys := hotkeys.NewHotKeySet(screenshotHk)

	app.Start(defaultHotKeys)
}
