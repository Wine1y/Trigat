package scWindow

import (
	"bytes"
	"image"
	"os"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	"github.com/Wine1y/trigat/config"
	"github.com/Wine1y/trigat/internal/gui"
	editTools "github.com/Wine1y/trigat/internal/gui/sc_window/edit_tools"
	"github.com/Wine1y/trigat/pkg"
	"github.com/Wine1y/trigat/pkg/hotkeys"
	"github.com/kbinani/screenshot"
	"github.com/veandco/go-sdl2/sdl"
	"golang.design/x/clipboard"
	hk "golang.design/x/hotkey"
)

const windowFlags uint32 = sdl.WINDOW_FULLSCREEN_DESKTOP | sdl.WINDOW_SKIP_TASKBAR | sdl.WINDOW_BORDERLESS | sdl.WINDOW_HIDDEN

var dimColor = sdl.Color{R: 0, G: 0, B: 0}
var dimAlpha uint8 = 100
var initAnimationDuration time.Duration = time.Millisecond * 750
var dimAnimationDuration time.Duration = time.Millisecond * 650

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
			int(config.GetAppFPS()), initAnimationDuration,
			1, false,
		),
		dimAnimation: pkg.NewLinearAnimation(
			0, int(dimAlpha),
			int(config.GetAppFPS()), dimAnimationDuration,
			1, false,
		),
		undimAnimation: pkg.NewLinearAnimation(
			int(dimAlpha), 0,
			int(config.GetAppFPS()), dimAnimationDuration,
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
	window.SDLWindow = sdlWindow
	window.screenshotTexture = pkg.CreateTextureFromSurface(window.Renderer(), screenshotSurface)
	window.SDLWin().SetWindowOpacity(0)
	window.SDLWin().Show()
	window.toolsPanel = NewToolsPanel(
		window.Renderer(),
		window.onNewToolSelected,
		window.saveImage, window.copyImage, window.searchImage,
	)
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

func (window ScreenshotWindow) renderScreenshot() (*[]byte, *sdl.Surface) {
	ren := window.Renderer()
	pkg.CopyTexture(ren, window.screenshotTexture, nil, nil)
	window.toolsPanel.RenderScreenshot(ren)
	pixels, surface := readRenderIntoSurface(ren)
	croppedSurface := window.toolsPanel.CropScreenshot(surface)
	if croppedSurface != surface {
		surface.Free()
	}
	return pixels, croppedSurface
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
			window.saveImage()
			return true
		}
		if keysym.Sym == sdl.K_c && (keysym.Mod&sdl.KMOD_CTRL) != 0 {
			window.copyImage()
			return true
		}
		if keysym.Sym == sdl.K_g && (keysym.Mod&sdl.KMOD_CTRL) != 0 {
			window.searchImage()
			return true
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

func (window *ScreenshotWindow) onNewToolSelected(tool editTools.ScreenshotEditTool) {
	if tool.RequiresScreenDim() {
		window.dimBackground()
	} else {
		window.undimBackground()
	}
}

func (window *ScreenshotWindow) dimBackground() {
	if !window.dimmed {
		window.dimmed = true
		window.dimAnimation.ReStart()
	}
}

func (window *ScreenshotWindow) undimBackground() {
	if window.dimmed {
		window.dimmed = false
		window.undimAnimation.ReStart()
	}
}

func (window *ScreenshotWindow) saveImage() {
	savingOptions, success := pkg.RequestSavingOptions("Saving screenshot", "screenshot")
	if !success {
		return
	}
	pixels, surface := window.renderScreenshot()
	window.Close()
	go func() {
		file, err := os.Create(savingOptions.Filepath)
		if err != nil {
			panic(err)
		}
		savingOptions.Method.WritingFunction(surface, file)
		surface.Free()
		file.Close()
		runtime.KeepAlive(pixels)
	}()
}

func (window *ScreenshotWindow) copyImage() {
	pixels, surface := window.renderScreenshot()
	window.Close()
	go func() {
		buf := bytes.NewBuffer(make([]byte, 0, surface.W*surface.H))
		pkg.WriteSurfaceToPNG(surface, buf)
		clipboard.Write(clipboard.FmtImage, buf.Bytes())
		surface.Free()
		runtime.KeepAlive(pixels)
	}()
}

func (window *ScreenshotWindow) searchImage() {
	pixels, surface := window.renderScreenshot()
	window.Close()
	go func() {
		buf := bytes.NewBuffer(make([]byte, 0, surface.W*surface.H))
		pkg.WriteSurfaceToPNG(surface, buf)
		imageUrl, err := pkg.UploadImage(buf)
		if err != nil {
			panic(err)
		}
		pkg.OpenUrlInBrowser(pkg.GetSearchURL(imageUrl))
		surface.Free()
		runtime.KeepAlive(pixels)
	}()
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

func readRenderIntoSurface(ren *sdl.Renderer) (surfaceData *[]byte, surface *sdl.Surface) {
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
	return &pixels, surface
}
