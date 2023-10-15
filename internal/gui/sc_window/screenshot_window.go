package scWindow

import (
	"image"
	"reflect"
	"time"
	"unsafe"

	"github.com/Wine1y/trigat/config"
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/pkg"
	"github.com/Wine1y/trigat/pkg/hotkeys"
	"github.com/kbinani/screenshot"
	"github.com/veandco/go-sdl2/sdl"
	hk "golang.design/x/hotkey"
)

const windowFlags uint32 = sdl.WINDOW_FULLSCREEN_DESKTOP | sdl.WINDOW_ALWAYS_ON_TOP | sdl.WINDOW_SKIP_TASKBAR | sdl.WINDOW_BORDERLESS | sdl.WINDOW_HIDDEN

var dimColor = sdl.Color{R: 0, G: 0, B: 0}
var dimAlpha uint8 = 100

type ScreenshotWindow struct {
	screenshotTexture *sdl.Texture
	toolsPanel        *ToolsPanel
	initAnimation     *pkg.Animation
	dimAnimation      *pkg.Animation
	undimAnimation    *pkg.Animation
	dimmed            bool
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
	window := ScreenshotWindow{
		dimmed: true,
		initAnimation: pkg.NewLinearAnimation(
			0, 100,
			int(config.GetAppFPS()), time.Millisecond*750,
			1, false,
		),
		dimAnimation: pkg.NewLinearAnimation(
			0, int(dimAlpha),
			int(config.GetAppFPS()), time.Millisecond*750,
			1, false,
		),
		undimAnimation: pkg.NewLinearAnimation(
			int(dimAlpha), 0,
			int(config.GetAppFPS()), time.Millisecond*750,
			1, false,
		),
	}
	window.dimAnimation.End()
	window.undimAnimation.End()
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
	window.SDLWin().SetWindowOpacity(0)
	window.SDLWin().Show()
	window.render(window.Renderer())
	window.Renderer().Present()
	return &window
}

func (window *ScreenshotWindow) render(ren *sdl.Renderer) {
	if !window.initAnimation.IsEnded() {
		window.SDLWin().SetWindowOpacity(float32(window.initAnimation.CurrentValue()) / 100)
	}
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
			pkg.CopyTexture(ren, window.screenshotTexture, nil, nil)
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
	exitHk := hotkeys.NewHotKey(hk.KeyEscape, nil, &exitCb, nil)
	return hotkeys.NewHotKeySet(exitHk)
}

func (window *ScreenshotWindow) DimBackground() {
	if !window.dimmed {
		window.dimmed = true
		window.dimAnimation.ReStart()
	}
}

func (window *ScreenshotWindow) UndimBackground() {
	if window.dimmed {
		window.dimmed = false
		window.undimAnimation.ReStart()
	}
}

func (window *ScreenshotWindow) drawScreenshotBackground(ren *sdl.Renderer) {
	pkg.CopyTexture(ren, window.screenshotTexture, nil, nil)

	rect := ren.GetViewport()

	var currentDimAlpha uint8
	if window.dimmed {
		currentDimAlpha = uint8(window.dimAnimation.CurrentValue())
	} else {
		currentDimAlpha = uint8(window.undimAnimation.CurrentValue())
	}
	pkg.DrawFilledRectangle(
		ren,
		&rect,
		sdl.Color{R: dimColor.R, G: dimColor.G, B: dimColor.B, A: currentDimAlpha},
	)
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
	pixels := pkg.ReadRGBA32(ren, nil)
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&pixels))
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
