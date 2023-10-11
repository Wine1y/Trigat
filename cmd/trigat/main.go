package main

import (
	"runtime"

	"github.com/Wine1y/trigat/internal"
	scWindow "github.com/Wine1y/trigat/internal/gui/sc_window"
	"github.com/Wine1y/trigat/pkg/hotkeys"
	hk "golang.design/x/hotkey"
)

func main() {
	app := internal.App{}
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
