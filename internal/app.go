package internal

import (
	"github.com/Wine1y/trigat/assets"
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/pkg/hotkeys"
	"github.com/getlantern/systray"
)

type App struct {
	currentWindow  gui.Window
	defaultHotKeys *hotkeys.HotKeySet
	currentHotKeys *hotkeys.HotKeySet
	exitCh         chan struct{}
}

func NewApp() *App {
	return &App{
		exitCh: make(chan struct{}),
	}
}

func (app *App) Start(defaultHotKeys *hotkeys.HotKeySet) {
	go app.startSystemTray()
	app.defaultHotKeys = defaultHotKeys
	app.setHotkeys(app.defaultHotKeys)
	println("App started")
	<-app.exitCh
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

func (app *App) Close() {
	systray.Quit()
	app.exitCh <- struct{}{}
}

func (app *App) startSystemTray() {
	systray.Run(app.onTrayStart, func() {})
}

func (app *App) onTrayStart() {
	systray.SetIcon(assets.TrigatIconData)
	systray.SetTooltip("Trigat")
	exitItem := systray.AddMenuItem("Exit", "Close the app")
	go func() {
		<-exitItem.ClickedCh
		app.Close()
	}()
}
