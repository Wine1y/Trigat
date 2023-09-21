package editTools

import (
	_ "embed"

	"github.com/Wine1y/trigat/gui"
	"github.com/Wine1y/trigat/utils"
	"github.com/veandco/go-sdl2/sdl"
)

var paintColor = sdl.Color{R: 255, G: 0, B: 0, A: 255}
var paintThickness int32 = 2

//go:embed icons/paint_tool.png
var paintIconData []byte
var paintRgbIcon = utils.LoadPNGSurface(paintIconData)

type PaintTool struct {
	isDragging bool
	strokes    [][]sdl.Point
}

func NewPaintTool() *PaintTool {
	return &PaintTool{
		isDragging: false,
		strokes:    make([][]sdl.Point, 0, 1),
	}
}

func (tool PaintTool) ToolIcon() *sdl.Surface {
	return paintRgbIcon
}

func (tool *PaintTool) ToolCallbacks(queue *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		tool.strokes = append(tool.strokes, []sdl.Point{{X: x, Y: y}})
		tool.isDragging = true
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		if !tool.isDragging {
			return false
		}
		tool.strokes[len(tool.strokes)-1] = append(
			tool.strokes[len(tool.strokes)-1],
			sdl.Point{X: x, Y: y},
		)
		return false
	})

	callbacks.MouseUp = append(callbacks.MouseUp, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT || !tool.isDragging {
			return false
		}
		tool.isDragging = false
		queue.Push(PaintAction{tool: tool, lastStroke: tool.strokes[len(tool.strokes)-1]})
		return false
	})

	return callbacks
}

func (tool PaintTool) RenderCurrentState(ren *sdl.Renderer) {
	for _, stroke := range tool.strokes {
		if len(stroke) == 1 {
			utils.DrawFilledRectangle(
				ren,
				&sdl.Rect{X: stroke[0].X, Y: stroke[0].Y, W: paintThickness, H: paintThickness},
				paintColor,
			)
			continue
		}
		for i := 0; i < len(stroke)-1; i++ {
			utils.DrawThickLine(ren, &stroke[i], &stroke[i+1], paintThickness, paintColor)
		}
	}
}

func (tool PaintTool) RenderScreenshot(ren *sdl.Renderer) {
	tool.RenderCurrentState(ren)
}

type PaintAction struct {
	tool       *PaintTool
	lastStroke []sdl.Point
}

func (action PaintAction) Undo() {
	action.tool.strokes = action.tool.strokes[:len(action.tool.strokes)-1]
}

func (action PaintAction) Redo() {
	action.tool.strokes = append(action.tool.strokes, action.lastStroke)
}
