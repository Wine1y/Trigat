package editTools

import (
	_ "embed"
	"fmt"

	"github.com/Wine1y/trigat/gui"
	"github.com/Wine1y/trigat/utils"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const selectionThickness int32 = 2
const selectionTooltipMargin int32 = 4

var selectionOuterTooltipColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var selectionInnerTooltipColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
var selectionBorderColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var selectionFillColor = sdl.Color{R: 255, G: 255, B: 255, A: 50}

//go:embed icons/selection_tool.png
var selectionIconData []byte
var selectionIcon = utils.LoadPNGSurface(selectionIconData)

type SelectionTool struct {
	isDragging     bool
	isShiftPressed bool
	selection      *sdl.Rect
	lastCursorPos  *sdl.Point
	tooltip        *selectionTooltip
	ren            *sdl.Renderer
}

func NewSelectionTool(renderer *sdl.Renderer) *SelectionTool {
	return &SelectionTool{
		isDragging:     false,
		isShiftPressed: false,
		tooltip:        &selectionTooltip{font: utils.LoadFont(defaultFontData, 14)},
		ren:            renderer,
	}
}

func (tool SelectionTool) ToolIcon() *sdl.Surface {
	return selectionIcon
}

func (tool *SelectionTool) ToolCallbacks(_ *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		tool.lastCursorPos = &sdl.Point{X: x, Y: y}
		tool.selection = &sdl.Rect{X: x, Y: y, W: 1, H: 1}
		tool.tooltip.updateTooltip(tool.ren, tool.selection)
		tool.isDragging = true
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		if tool.isDragging {
			sel := tool.selection
			sel.W = x - sel.X
			sel.H = y - sel.Y
			if tool.isShiftPressed {
				utils.RectIntoSquare(sel)
			}
			tool.tooltip.updateTooltip(tool.ren, tool.selection)
		}
		tool.lastCursorPos = &sdl.Point{X: x, Y: y}
		return false
	})

	callbacks.MouseUp = append(callbacks.MouseUp, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT || !tool.isDragging {
			return false
		}
		tool.lastCursorPos = &sdl.Point{X: x, Y: y}
		tool.isDragging = false
		return false
	})

	callbacks.KeyDown = append(callbacks.KeyDown, func(keysym sdl.Keysym) bool {
		if keysym.Sym != sdl.K_LSHIFT && keysym.Sym != sdl.K_RSHIFT {
			return false
		}
		if tool.isDragging {
			utils.RectIntoSquare(tool.selection)
			tool.tooltip.updateTooltip(tool.ren, tool.selection)
		}
		tool.isShiftPressed = true
		return false
	})

	callbacks.KeyUp = append(callbacks.KeyUp, func(keysym sdl.Keysym) bool {
		if keysym.Sym != sdl.K_LSHIFT && keysym.Sym != sdl.K_RSHIFT {
			return false
		}
		if tool.isDragging {
			sel := tool.selection
			sel.W = tool.lastCursorPos.X - sel.X
			sel.H = tool.lastCursorPos.Y - sel.Y
			tool.tooltip.updateTooltip(tool.ren, tool.selection)
		}
		tool.isShiftPressed = false
		return false
	})

	callbacks.Quit = append(callbacks.Quit, func() bool {
		if tool.tooltip.texture != nil {
			tool.tooltip.texture.Destroy()
		}
		return false
	})

	return callbacks
}

func (tool SelectionTool) RenderCurrentState(ren *sdl.Renderer) {
	if tool.selection != nil {
		sel := tool.selection
		utils.DrawFilledRectangle(ren, sel, selectionFillColor)
		utils.DrawThickRectangle(ren, sel, selectionThickness, selectionBorderColor)
		tool.tooltip.texture.Draw(ren, tool.tooltip.startingPosition)
	}
}

func (tool SelectionTool) RenderScreenshot(_ *sdl.Renderer) {}

func (tool SelectionTool) CropScreenshot(surface *sdl.Surface) *sdl.Surface {
	if tool.selection != nil {
		sel := tool.selection
		croppedSurface, err := sdl.CreateRGBSurface(
			0,
			sel.W, sel.H,
			int32(surface.Format.BitsPerPixel),
			surface.Format.Rmask, surface.Format.Gmask, surface.Format.Bmask, surface.Format.Amask,
		)
		if err != nil {
			panic(err)
		}
		if err := surface.Blit(sel, croppedSurface, nil); err != nil {
			panic(err)
		}
		return croppedSurface
	}
	return surface
}

type selectionTooltip struct {
	texture          *utils.StringTexture
	startingPosition *sdl.Point
	font             *ttf.Font
	color            *sdl.Color
}

func (tooltip *selectionTooltip) updateTooltip(ren *sdl.Renderer, selection *sdl.Rect) {
	text := fmt.Sprintf("%v x %v", utils.Abs(selection.W), utils.Abs(selection.H))
	textW, textH, _ := tooltip.font.SizeUTF8(text)

	startingPoint := sdl.Point{
		X: selection.X,
		Y: selection.Y + selectionTooltipMargin,
	}
	if selection.W < 0 {
		startingPoint.X -= (int32(textW) + selectionThickness)
	}
	if selection.H > 0 {
		startingPoint.Y += selection.H
	}
	vp := ren.GetViewport()
	tooltip.color = &selectionOuterTooltipColor
	if startingPoint.Y+int32(textH) > vp.H {
		startingPoint.Y -= (int32(textH) + selectionTooltipMargin + selectionThickness)
		tooltip.color = &selectionInnerTooltipColor
	}
	tooltip.startingPosition = &startingPoint
	tooltip.texture = utils.NewStringTexture(ren, tooltip.font, text, *tooltip.color)
}
