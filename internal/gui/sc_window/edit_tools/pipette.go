package editTools

import (
	_ "embed"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/Wine1y/trigat/assets"
	"github.com/Wine1y/trigat/config"
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/pkg"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const pipetteWidgetMargin int32 = 10
const pipetteWidgetCornerRadius int32 = 8
const pipetteWidgetCopiedFontSize int = 16
const pipetteWidgetCopiedTextAnimationDuration time.Duration = time.Millisecond * 1200

const colorTipleteCornerRadius int32 = 4
const colorTipleteSquareSide int32 = 40
const colorTipleteSquareMargin int32 = 5
const colorTripletShadingFactor float64 = 0.5
const colorTripletLightningFactor float64 = 1.5

var pipetteWidgetBackground sdl.Color = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var pipetteWidgetCurrentSquareColor sdl.Color = sdl.Color{R: 0, G: 0, B: 0, A: 255}
var pipetteWidgetCopiedTextColor sdl.Color = sdl.Color{R: 255, G: 255, B: 255, A: 255}

type PipetteTool struct {
	ren         *sdl.Renderer
	isDragging  bool
	widget      pipetteWidget
	deactivated bool
	DefaultScreenshotEditTool
}

func NewPipetteTool(renderer *sdl.Renderer) *PipetteTool {
	vp := renderer.GetViewport()
	widget := newPipetteWidget()
	widget.resize(vp.W, vp.H)
	return &PipetteTool{
		isDragging:  false,
		ren:         renderer,
		widget:      *widget,
		deactivated: true,
	}
}

func (tool PipetteTool) ToolIcon() *sdl.Surface {
	return assets.PipetteIcon
}

func (tool *PipetteTool) OnToolActivated() {
	tool.deactivated = false
}

func (tool *PipetteTool) OnToolDeactivated() {
	tool.isDragging = false
	tool.deactivated = true
}

func (tool *PipetteTool) ToolCallbacks(_ *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		click := sdl.Point{X: x, Y: y}
		inWidget := click.InRect(&tool.widget.bbox)
		switch {
		case button == sdl.BUTTON_RIGHT && !inWidget:
			color := tool.NewProbe(x, y)
			tool.copyColorToClipboard(color)
		case button == sdl.BUTTON_LEFT && !inWidget:
			tool.NewProbe(x, y)
			tool.isDragging = true
		case button == sdl.BUTTON_LEFT && inWidget:
			if color, clickedAtColorBox := tool.widget.getColorBoxAt(x, y); clickedAtColorBox {
				tool.copyColorToClipboard(*color)
			}
		}
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		if tool.isDragging {
			tool.NewProbe(x, y)
		}
		if _, colorHovered := tool.widget.getColorBoxAt(x, y); colorHovered {
			sdl.SetCursor(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_HAND))
		} else {
			sdl.SetCursor(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW))
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

	callbacks.SizeChange = append(callbacks.SizeChange, func(w, h int32) bool {
		tool.widget.resize(w, h)
		return false
	})

	callbacks.Quit = append(callbacks.Quit, func() bool {
		if tool.widget.copiedTexture != nil {
			tool.widget.copiedTexture.Destroy()
		}
		return false
	})

	return callbacks
}

func (tool *PipetteTool) NewProbe(x, y int32) sdl.Color {
	color := tool.getPixelColor(x, y)
	tool.widget.newColor(color)
	return color
}

func (tool PipetteTool) RenderScreenshot(_ *sdl.Renderer) {}

func (tool PipetteTool) RenderCurrentState(ren *sdl.Renderer) {
	if !tool.deactivated {
		tool.widget.draw(ren)
	}
}

func (tool PipetteTool) getPixelColor(x, y int32) sdl.Color {
	pixel := make([]uint8, 4)
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&pixel))
	err := tool.ren.ReadPixels(
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

func (tool *PipetteTool) copyColorToClipboard(color sdl.Color) error {
	copiedStr := fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B)
	if err := sdl.SetClipboardText(copiedStr); err != nil {
		return err
	}
	tool.widget.newCopiedColor(tool.ren, copiedStr)
	return nil
}

type pipetteWidget struct {
	bbox             sdl.Rect
	colorSquaresBBox [3]sdl.Rect
	colors           [3]sdl.Color
	initialized      bool
	lastCopiedString string
	copiedAnimation  *pkg.Animation
	copiedTexture    *pkg.StringTexture
	copiedFont       *ttf.Font
}

func newPipetteWidget() *pipetteWidget {
	copiedAnimation := pkg.NewLinearAnimation(
		0, 255,
		int(config.GetAppFPS()), pipetteWidgetCopiedTextAnimationDuration,
		1, true,
	)
	copiedAnimation.End()
	return &pipetteWidget{
		copiedAnimation: copiedAnimation,
		copiedFont:      assets.GetAppFont(pipetteWidgetCopiedFontSize),
	}
}

func (widget *pipetteWidget) resize(w, h int32) {
	widgetW := colorTipleteSquareSide*int32(cap(widget.colors)) + colorTipleteSquareMargin*int32(cap(widget.colors)+1)
	widgetH := colorTipleteSquareSide + colorTipleteSquareMargin*2
	widget.bbox = sdl.Rect{
		X: pipetteWidgetMargin, Y: h - pipetteWidgetMargin - widgetH,
		W: widgetW, H: widgetH,
	}
	for i := 0; i < len(widget.colors); i++ {
		widget.colorSquaresBBox[i] = sdl.Rect{
			X: widget.bbox.X + colorTipleteSquareMargin + int32(i)*(colorTipleteSquareSide+colorTipleteSquareMargin),
			Y: widget.bbox.Y + colorTipleteSquareMargin,
			W: colorTipleteSquareSide, H: colorTipleteSquareSide,
		}
	}
}

func (widget *pipetteWidget) newColor(color sdl.Color) {
	shaded := sdl.Color{
		R: uint8(pkg.Clamp(0, float64(color.R)*colorTripletShadingFactor, 255)),
		G: uint8(pkg.Clamp(0, float64(color.G)*colorTripletShadingFactor, 255)),
		B: uint8(pkg.Clamp(0, float64(color.B)*colorTripletShadingFactor, 255)),
		A: color.A,
	}

	lighted := sdl.Color{
		R: uint8(pkg.Clamp(0, float64(color.R)*colorTripletLightningFactor, 255)),
		G: uint8(pkg.Clamp(0, float64(color.G)*colorTripletLightningFactor, 255)),
		B: uint8(pkg.Clamp(0, float64(color.B)*colorTripletLightningFactor, 255)),
		A: color.A,
	}
	widget.colors[0] = shaded
	widget.colors[1] = color
	widget.colors[2] = lighted
	widget.initialized = true
}

func (widget *pipetteWidget) newCopiedColor(ren *sdl.Renderer, copiedFormatStr string) {
	if widget.lastCopiedString != copiedFormatStr {
		widget.lastCopiedString = copiedFormatStr
		if widget.copiedTexture != nil {
			widget.copiedTexture.Destroy()
		}
		widget.copiedTexture = pkg.NewStringTexture(
			ren,
			widget.copiedFont,
			fmt.Sprintf("%v copied!", copiedFormatStr),
			pipetteWidgetCopiedTextColor,
		)
	}
	widget.copiedAnimation.ReStart()
}

func (widget pipetteWidget) getColorBoxAt(x, y int32) (*sdl.Color, bool) {
	if !widget.initialized {
		return nil, false
	}
	point := sdl.Point{X: x, Y: y}
	for i := 0; i < len(widget.colors); i++ {
		if point.InRect(&widget.colorSquaresBBox[i]) {
			return &widget.colors[i], true
		}
	}
	return nil, false
}

func (widget pipetteWidget) draw(ren *sdl.Renderer) {
	if !widget.initialized {
		return
	}
	pkg.DrawRoundedFilledRectangle(ren, &widget.bbox, pipetteWidgetCornerRadius, pipetteWidgetBackground)
	for i := 0; i < len(widget.colors); i++ {
		colorSquareBBox := widget.colorSquaresBBox[i]
		pkg.DrawRoundedFilledRectangle(
			ren,
			&colorSquareBBox,
			colorTipleteCornerRadius,
			widget.colors[i],
		)
	}
	if !widget.copiedAnimation.IsEnded() {
		widget.copiedTexture.Texture.SetAlphaMod(uint8(widget.copiedAnimation.CurrentValue()))
		textureLT := sdl.Point{
			X: widget.bbox.X + (widget.bbox.W-widget.copiedTexture.TextWidth)/2,
			Y: widget.bbox.Y - pipetteWidgetMargin - widget.copiedTexture.TextHeight,
		}
		widget.copiedTexture.Draw(ren, &textureLT)
	}
}
