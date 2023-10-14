package gui

import (
	"runtime"
	"unicode/utf8"

	"github.com/Wine1y/trigat/config"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var TICKS_PER_FRAME uint64 = 1000 / config.GetAppFPS()

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
	if err := ttf.Init(); err != nil {
		panic(err)
	}
	win, err := sdl.CreateWindow(title, x, y, width, height, flags|sdl.WINDOW_HIDDEN)
	if err != nil {
		panic(err)
	}
	ren, err := sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	//Made to avoid unrendered empty frame when starting the app
	win.SetWindowOpacity(0)
	win.Show()

	return &SDLWindow{
		win:       win,
		ren:       ren,
		render:    renderCallback,
		callbacks: callbacks,
	}
}

func (window *SDLWindow) StartMainLoop() {
	window.render(window.ren)
	window.ren.Present()
	window.win.SetWindowOpacity(1)
	lastTick := sdl.GetTicks64()
	for {
		window.shouldClose = window.handleEvents()
		if window.shouldClose {
			break
		}
		window.render(window.ren)
		window.ren.Present()
		ticksPassed := sdl.GetTicks64() - lastTick
		if ticksPassed < TICKS_PER_FRAME {
			sdl.Delay(uint32(TICKS_PER_FRAME - ticksPassed))
		}
		lastTick = sdl.GetTicks64()
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
