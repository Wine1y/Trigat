package editTools

import (
	_ "embed"

	"github.com/Wine1y/trigat/assets"
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/internal/gui/sc_window/settings"
	"github.com/Wine1y/trigat/pkg"
	"github.com/veandco/go-sdl2/sdl"
)

type PaintTool struct {
	isDragging     bool
	strokes        []paintStroke
	settings       []settings.ToolSetting
	paintThickness int32
	paintColor     sdl.Color
	DefaultScreenshotEditTool
}

func NewPaintTool() *PaintTool {
	tool := PaintTool{
		isDragging: false,
		strokes:    make([]paintStroke, 0, 1),
	}

	widthSlider := settings.NewSliderSetting(1, 5, func(value uint) {
		tool.paintThickness = int32(value)
	})

	colorPicker := settings.NewColorPickerSetting(func(color sdl.Color) {
		tool.paintColor = color
	})

	toolSettings := []settings.ToolSetting{widthSlider, colorPicker}

	tool.paintThickness = int32(widthSlider.CurrentValue())
	tool.paintColor = colorPicker.CurrentColor()
	tool.settings = toolSettings
	return &tool
}

func (tool PaintTool) ToolIcon() *sdl.Surface {
	return assets.PaintIcon
}

func (tool *PaintTool) ToolCallbacks(queue *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		tool.strokes = append(
			tool.strokes,
			paintStroke{
				points:    []sdl.Point{{X: x, Y: y}},
				thickness: tool.paintThickness,
				color:     tool.paintColor,
			},
		)
		tool.isDragging = true
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		if !tool.isDragging {
			return false
		}
		tool.strokes[len(tool.strokes)-1].points = append(
			tool.strokes[len(tool.strokes)-1].points,
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
		if len(stroke.points) == 1 {
			pkg.DrawFilledRectangle(
				ren,
				&sdl.Rect{
					X: stroke.points[0].X, Y: stroke.points[0].Y,
					W: stroke.thickness, H: stroke.thickness,
				},
				stroke.color,
			)
			continue
		}
		for i := 0; i < len(stroke.points)-1; i++ {
			pkg.DrawThickLine(ren, &stroke.points[i], &stroke.points[i+1], stroke.thickness, stroke.color)
		}
	}
}

func (tool PaintTool) RenderScreenshot(ren *sdl.Renderer) {
	tool.RenderCurrentState(ren)
}

func (tool PaintTool) ToolSettings() []settings.ToolSetting {
	return tool.settings
}

func (tool PaintTool) ToolColor() *sdl.Color {
	return &tool.paintColor
}

func (tool *PaintTool) OnToolDeactivated() {
	tool.isDragging = false
}

type paintStroke struct {
	points    []sdl.Point
	thickness int32
	color     sdl.Color
}

type PaintAction struct {
	tool       *PaintTool
	lastStroke paintStroke
}

func (action PaintAction) Undo() {
	action.tool.strokes = action.tool.strokes[:len(action.tool.strokes)-1]
}

func (action PaintAction) Redo() {
	action.tool.strokes = append(action.tool.strokes, action.lastStroke)
}
