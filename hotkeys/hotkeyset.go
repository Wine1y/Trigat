package hotkeys

type HotKeySet struct {
	hotkeys []*AppHotKey
}

func NewHotKeySet(hotkeys ...*AppHotKey) *HotKeySet {
	return &HotKeySet{
		hotkeys: hotkeys,
	}
}

func (set *HotKeySet) StartListeningAll() error {
	for _, appHotKey := range set.hotkeys {
		if err := appHotKey.register(); err != nil {
			return err
		}
		go appHotKey.startListening()
	}
	return nil
}

func (set *HotKeySet) StopListeningAll() error {
	for _, appHotKey := range set.hotkeys {
		appHotKey.stopListening()
		if err := appHotKey.unregister(); err != nil {
			return err
		}
	}
	return nil
}
