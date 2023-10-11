package hotkeys

import "golang.design/x/hotkey"

type AppHotKey struct {
	hk          *hotkey.Hotkey
	isListening bool
	stopChan    chan bool
	onKeyDown   *func()
	onKeyUp     *func()
}

func NewHotKey(
	key hotkey.Key,
	mods []hotkey.Modifier,
	onKeyDown *func(),
	onKeyUp *func(),
) *AppHotKey {
	hk := hotkey.New(mods, key)
	return &AppHotKey{
		hk:          hk,
		isListening: false,
		stopChan:    make(chan bool),
		onKeyDown:   onKeyDown,
		onKeyUp:     onKeyUp,
	}
}

func (appHotKey *AppHotKey) register() error {
	return appHotKey.hk.Register()
}
func (appHotKey *AppHotKey) unregister() error {
	return appHotKey.hk.Unregister()
}

func (appHotKey *AppHotKey) startListening() {
	if appHotKey.isListening {
		panic("Invalid AppHotKey state (already listening)")
	}
	appHotKey.isListening = true
	for {
		select {
		case <-appHotKey.stopChan:
			appHotKey.isListening = false
			return
		case <-appHotKey.hk.Keydown():
			if appHotKey.onKeyDown != nil {
				go (*appHotKey.onKeyDown)()
			}
		case <-appHotKey.hk.Keyup():
			if appHotKey.onKeyUp != nil {
				go (*appHotKey.onKeyUp)()
			}
		}
	}
}

func (appHotKey *AppHotKey) stopListening() {
	if !appHotKey.isListening {
		panic("Invalid AppHotKey state (already stopped)")
	}
	appHotKey.stopChan <- true
}
