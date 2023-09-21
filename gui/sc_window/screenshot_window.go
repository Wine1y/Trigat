package scWindow

import (
	"image"
	"reflect"
	"unsafe"

	"github.com/Wine1y/trigat/gui"
	"github.com/Wine1y/trigat/hotkeys"
	"github.com/Wine1y/trigat/utils"
	"github.com/kbinani/screenshot"
	"github.com/veandco/go-sdl2/sdl"
	hk "golang.design/x/hotkey"
)

const windowFlags uint32 = sdl.WINDOW_FULLSCREEN_DESKTOP | sdl.WINDOW_ALWAYS_ON_TOP | sdl.WINDOW_SKIP_TASKBAR | sdl.WINDOW_BORDERLESS

var dimColor = sdl.Color{R: 0, G: 0, B: 0, A: 100}

type ScreenshotWindow struct {
	screenshotTexture *sdl.Texture
	toolsPanel        *ToolsPanel
	*gui.SDLWindow
}

func NewScreenshotWindow() *ScreenshotWindow {
	screenImage, err := takeScreenshot()
	if err != nil {
		panic("Can't take a screenshot")
	}
	screenshotSurface, err := getScreenshotSurface(screenImage)
	if err != nil {
		panic("Can't create a screenshot surface")
	}
	defer screenshotSurface.Free()
	window := ScreenshotWindow{}
	sdlWindow := gui.NewSDLWindow(
		"",
		640, 480, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		windowFlags,
		window.render,
		window.callbackSet,
	)
	screenshotTexture, err := sdlWindow.Renderer().CreateTextureFromSurface(screenshotSurface)
	if err != nil {
		panic("Can't create a screenshot texure")
	}
	window.toolsPanel = NewToolsPanel(sdlWindow.Renderer())
	window.screenshotTexture = screenshotTexture
	window.SDLWindow = sdlWindow
	return &window
}

func (window *ScreenshotWindow) render(ren *sdl.Renderer) {
	window.drawScreenshotBackground(ren)
	window.toolsPanel.DrawToolsState(ren)
	window.toolsPanel.DrawPanel(ren)
}

func (window *ScreenshotWindow) callbackSet() *gui.WindowCallbackSet {
	set := gui.NewWindowCallbackSet()
	window.toolsPanel.SetToolsCallbacks(set)
	set.Quit = append(set.Quit, func() bool {
		window.screenshotTexture.Destroy()
		return false
	})
	set.KeyDown = append(set.KeyDown, func(keysym sdl.Keysym) bool {
		if keysym.Sym == sdl.K_s && (keysym.Mod&sdl.KMOD_CTRL) != 0 {
			ren := window.Renderer()
			utils.CopyTexture(ren, window.screenshotTexture, nil, nil)
			window.toolsPanel.RenderScreenshot(ren)
			surface := readRenderIntoSurface(ren)
			croppedSurface := window.toolsPanel.CropScreenshot(surface)
			croppedSurface.SaveBMP("C:\\Users\\Q\\Desktop\\trigat_screen.bmp")
			if croppedSurface != surface {
				croppedSurface.Free()
			}
			surface.Free()
			window.Close()
		}
		return false
	})
	return set
}

func (window *ScreenshotWindow) HotKeys() *hotkeys.HotKeySet {
	exitCb := func() { window.Close() }
	undoCb := func() {
		window.toolsPanel.UndoLastAction()
	}
	redoCb := func() {
		window.toolsPanel.RedoLastAction()
	}
	exitHk := hotkeys.NewHotKey(hk.KeyEscape, nil, &exitCb, nil)
	undoHk := hotkeys.NewHotKey(hk.KeyZ, []hk.Modifier{hk.ModCtrl}, &undoCb, nil)
	redoHk := hotkeys.NewHotKey(hk.KeyZ, []hk.Modifier{hk.ModCtrl, hk.ModAlt}, &redoCb, nil)
	return hotkeys.NewHotKeySet(exitHk, undoHk, redoHk)
}

func (window *ScreenshotWindow) drawScreenshotBackground(ren *sdl.Renderer) {
	utils.CopyTexture(ren, window.screenshotTexture, nil, nil)

	rect := ren.GetViewport()
	utils.DrawFilledRectangle(ren, &rect, dimColor)
}

func takeScreenshot() (*image.RGBA, error) {
	return screenshot.CaptureDisplay(0)
}

func getScreenshotSurface(screenshot *image.RGBA) (*sdl.Surface, error) {
	w, h := screenshot.Rect.Max.X, screenshot.Rect.Max.Y
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&screenshot.Pix))
	//This SliceHeader hack is used to avoid "cgo argument has Go pointer to Go pointer" exception
	//Although, by doing this we should be sure that image won't be deallocated until surface is freed
	screenshotSurface, err := sdl.CreateRGBSurfaceFrom(
		unsafe.Pointer(sh.Data),
		int32(w),
		int32(h),
		32,
		w*4,
		0x000000FF, 0x0000FF00, 0x00FF0000, 0xFF000000,
	)
	if err != nil {
		return nil, err
	}
	return screenshotSurface, nil
}

func readRenderIntoSurface(ren *sdl.Renderer) *sdl.Surface {
	vp := ren.GetViewport()
	pitch := int(vp.W) * 4
	pixels := make([]uint8, pitch*int(vp.H))

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&pixels))
	err := ren.ReadPixels(
		nil,
		uint32(sdl.PIXELFORMAT_RGBA32),
		unsafe.Pointer(sh.Data),
		pitch,
	)
	if err != nil {
		panic(err)
	}
	surface, err := sdl.CreateRGBSurfaceFrom(
		unsafe.Pointer(sh.Data),
		vp.W,
		vp.H,
		32,
		pitch,
		0x000000FF, 0x0000FF00, 0x00FF0000, 0xFF000000,
	)
	if err != nil {
		panic(err)
	}
	return surface
}
