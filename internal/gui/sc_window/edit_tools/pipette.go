package editTools

import (
	_ "embed"
	"reflect"
	"unsafe"

	"github.com/Wine1y/trigat/assets"
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/pkg"
	"github.com/veandco/go-sdl2/sdl"
)

type PipetteTool struct {
	ren        *sdl.Renderer
	isDragging bool
	lastColor  *sdl.Color
	DefaultScreenshotEditTool
}

func NewPipetteTool(renderer *sdl.Renderer) *PipetteTool {
	return &PipetteTool{
		isDragging: false,
		ren:        renderer,
	}
}

func (tool PipetteTool) ToolIcon() *sdl.Surface {
	return assets.PipetteIcon
}

func (tool *PipetteTool) OnToolDeactivated() {
	tool.isDragging = false
}

func (tool *PipetteTool) ToolCallbacks(_ *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		color := getPixelColor(tool.ren, x, y)
		tool.lastColor = &color
		tool.isDragging = true
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		if tool.isDragging {
			color := getPixelColor(tool.ren, x, y)
			tool.lastColor = &color
		}
		return false
	})

	callbacks.MouseUp = append(callbacks.MouseUp, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT || !tool.isDragging {
			return false
		}
		tool.isDragging = false
		return false
	})
	return callbacks
}

func (tool PipetteTool) RenderCurrentState(ren *sdl.Renderer) {
	if tool.lastColor != nil {
		vp := ren.GetViewport()
		pkg.DrawFilledCircle(ren, &sdl.Point{X: 55, Y: vp.H - 55}, 50, *tool.lastColor)
	}
}

func (tool PipetteTool) RenderScreenshot(_ *sdl.Renderer) {}

func getPixelColor(ren *sdl.Renderer, x, y int32) sdl.Color {
	pixel := make([]uint8, 4)
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&pixel))
	err := ren.ReadPixels(
		&sdl.Rect{X: x, Y: y, W: 1, H: 1},
		uint32(sdl.PIXELFORMAT_RGBA32),
		unsafe.Pointer(sh.Data),
		4,
	)
	if err != nil {
		panic(err)
	}
	return sdl.Color{R: pixel[0], G: pixel[1], B: pixel[2], A: pixel[3]}
}
