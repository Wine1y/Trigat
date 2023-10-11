package gui

import (
	"github.com/Wine1y/trigat/pkg/hotkeys"
	"github.com/veandco/go-sdl2/sdl"
)

type Window interface {
	HotKeys() *hotkeys.HotKeySet
	StartMainLoop()
	Close()
}

type WindowCallbackSet struct {
	MouseDown  []func(button uint8, x, y int32) bool
	MouseMove  []func(x, y int32) bool
	MouseUp    []func(button uint8, x, y int32) bool
	MouseWheel []func(x, y int32) bool
	KeyDown    []func(keysym sdl.Keysym) bool
	KeyUp      []func(keysym sdl.Keysym) bool
	TextInput  []func(rn rune) bool
	SizeChange []func(w, h int32) bool
	Quit       []func() bool
}

func NewWindowCallbackSet() *WindowCallbackSet {
	return &WindowCallbackSet{
		MouseDown:  make([]func(button uint8, x, y int32) bool, 0),
		MouseMove:  make([]func(x, y int32) bool, 0),
		MouseUp:    make([]func(button uint8, x, y int32) bool, 0),
		MouseWheel: make([]func(x, y int32) bool, 0),
		KeyDown:    make([]func(keysym sdl.Keysym) bool, 0),
		KeyUp:      make([]func(keysym sdl.Keysym) bool, 0),
		TextInput:  make([]func(rn rune) bool, 0),
		SizeChange: make([]func(w, h int32) bool, 0),
		Quit:       make([]func() bool, 0),
	}
}

func (set *WindowCallbackSet) Append(another *WindowCallbackSet) {
	set.MouseDown = append(set.MouseDown, another.MouseDown...)
	set.MouseMove = append(set.MouseMove, another.MouseMove...)
	set.MouseUp = append(set.MouseUp, another.MouseUp...)
	set.MouseWheel = append(set.MouseWheel, another.MouseWheel...)
	set.KeyDown = append(set.KeyDown, another.KeyDown...)
	set.KeyUp = append(set.KeyUp, another.KeyUp...)
	set.TextInput = append(set.TextInput, another.TextInput...)
	set.SizeChange = append(set.SizeChange, another.SizeChange...)
	set.Quit = append(set.Quit, another.Quit...)
}

func (set *WindowCallbackSet) Reset() {
	set.MouseDown = make([]func(button uint8, x, y int32) bool, 0)
	set.MouseMove = make([]func(x, y int32) bool, 0)
	set.MouseUp = make([]func(button uint8, x, y int32) bool, 0)
	set.MouseWheel = make([]func(x, y int32) bool, 0)
	set.KeyDown = make([]func(keysym sdl.Keysym) bool, 0)
	set.KeyUp = make([]func(keysym sdl.Keysym) bool, 0)
	set.TextInput = make([]func(rn rune) bool, 0)
	set.SizeChange = make([]func(w, h int32) bool, 0)
	set.Quit = make([]func() bool, 0)
}
