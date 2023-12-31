package gui

import (
	"runtime"
	"time"
	"unicode/utf8"

	"github.com/Wine1y/trigat/config"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const WINDOWSIZE_FULLSCREEN int32 = 0

var MILLISECONDS_PER_FRAME int64 = 1000 / int64(config.GetAppFPS())
var ArrowCursor *sdl.Cursor = nil
var HandCursor *sdl.Cursor = nil
var IBeamCursor *sdl.Cursor = nil
var SizeAllCursor *sdl.Cursor = nil

type SDLWindow struct {
	win         *sdl.Window
	ren         *sdl.Renderer
	shouldClose bool
	render      func(*sdl.Renderer)
	callbacks   func() *WindowCallbackSet
}

func NewSDLWindow(
	title string,
	width, height, x, y int32,
	flags uint32,
	renderCallback func(*sdl.Renderer),
	callbacks func() *WindowCallbackSet,
) *SDLWindow {
	runtime.LockOSThread()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	loadSystemCursors()
	if err := ttf.Init(); err != nil {
		panic(err)
	}

	if width == WINDOWSIZE_FULLSCREEN || height == WINDOWSIZE_FULLSCREEN {
		displayMode, err := sdl.GetCurrentDisplayMode(0)
		if err != nil {
			panic(err)
		}
		if width == WINDOWSIZE_FULLSCREEN {
			width = displayMode.W
		}
		if height == WINDOWSIZE_FULLSCREEN {
			height = displayMode.H
		}
	}

	win, err := sdl.CreateWindow(title, x, y, width, height, flags|sdl.WINDOW_HIDDEN)
	if err != nil {
		panic(err)
	}
	ren, err := sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	return &SDLWindow{
		win:       win,
		ren:       ren,
		render:    renderCallback,
		callbacks: callbacks,
	}
}

func (window *SDLWindow) StartMainLoop() {
	lastTick := time.Now()
	for {
		window.shouldClose = window.handleEvents()
		if window.shouldClose {
			break
		}
		window.render(window.ren)
		window.ren.Present()
		msPassed := time.Since(lastTick).Milliseconds()
		if msPassed < MILLISECONDS_PER_FRAME {
			time.Sleep(time.Millisecond * time.Duration(MILLISECONDS_PER_FRAME-msPassed))
		}
		lastTick = time.Now()
	}
	for _, cb := range window.callbacks().Quit {
		if cb() {
			break
		}
	}
	window.ren.Destroy()
	window.win.Destroy()
	sdl.Quit()
	ttf.Quit()
}

func (window *SDLWindow) handleEvents() bool {
	callbackSet := window.callbacks()
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.GetType() {
		case sdl.QUIT:
			return true
		case sdl.MOUSEBUTTONDOWN:
			event := event.(*sdl.MouseButtonEvent)
			for _, cb := range callbackSet.MouseDown {
				if cb(event.Button, event.X, event.Y) {
					break
				}
			}
		case sdl.MOUSEBUTTONUP:
			event := event.(*sdl.MouseButtonEvent)
			for _, cb := range callbackSet.MouseUp {
				if cb(event.Button, event.X, event.Y) {
					break
				}
			}
		case sdl.MOUSEMOTION:
			event := event.(*sdl.MouseMotionEvent)
			for _, cb := range callbackSet.MouseMove {
				if cb(event.X, event.Y) {
					break
				}
			}
		case sdl.MOUSEWHEEL:
			event := event.(*sdl.MouseWheelEvent)
			for _, cb := range callbackSet.MouseWheel {
				if cb(event.X, event.Y) {
					break
				}
			}
		case sdl.KEYDOWN:
			event := event.(*sdl.KeyboardEvent)
			for _, cb := range callbackSet.KeyDown {
				if cb(event.Keysym) {
					break
				}
			}
		case sdl.KEYUP:
			event := event.(*sdl.KeyboardEvent)
			for _, cb := range callbackSet.KeyUp {
				if cb(event.Keysym) {
					break
				}
			}
		case sdl.TEXTINPUT:
			event := event.(*sdl.TextInputEvent)
			for _, cb := range callbackSet.TextInput {
				rn, _ := utf8.DecodeRune(event.Text[:])
				if cb(rn) {
					break
				}
			}
		case sdl.WINDOWEVENT:
			event := event.(*sdl.WindowEvent)
			if event.Event == sdl.WINDOWEVENT_RESIZED {
				for _, cb := range callbackSet.SizeChange {
					if cb(event.Data1, event.Data2) {
						break
					}
				}
			}
		}
	}
	return window.shouldClose
}

func (window SDLWindow) Renderer() *sdl.Renderer {
	return window.ren
}

func (window SDLWindow) SDLWin() *sdl.Window {
	return window.win
}

func (window *SDLWindow) Close() {
	window.shouldClose = true
}

func loadSystemCursors() {
	ArrowCursor = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW)
	HandCursor = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_HAND)
	IBeamCursor = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_IBEAM)
	SizeAllCursor = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEALL)
}
