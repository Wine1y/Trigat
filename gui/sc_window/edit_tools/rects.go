package editTools

import (
	_ "embed"

	"github.com/Wine1y/trigat/gui"
	"github.com/Wine1y/trigat/gui/sc_window/settings"
	"github.com/Wine1y/trigat/utils"
	"github.com/veandco/go-sdl2/sdl"
)

//go:embed icons/rect_tool.png
var rectIconData []byte
var rectIcon = utils.LoadPNGSurface(rectIconData)

type RectsTool struct {
	isDragging          bool
	isShiftPressed      bool
	rects               []rect
	lastCursorPos       *sdl.Point
	rectBorderThickness int32
	rectColor           sdl.Color
	settings            []settings.ToolSetting
}

func NewRectsTool() *RectsTool {
	tool := RectsTool{
		isDragging:     false,
		isShiftPressed: false,
		rects:          make([]rect, 0, 1),
	}

	widthSlider := settings.NewSliderSetting(1, 10, func(value uint) {
		tool.rectBorderThickness = int32(value)
	})

	colorPicker := settings.NewColorPickerSetting(func(color sdl.Color) {
		tool.rectColor = color
	})

	toolSettings := []settings.ToolSetting{widthSlider, colorPicker}

	tool.rectBorderThickness = int32(widthSlider.CurrentValue())
	tool.rectColor = colorPicker.CurrentColor()
	tool.settings = toolSettings
	return &tool
}

func (tool RectsTool) ToolIcon() *sdl.Surface {
	return rectIcon
}

func (tool *RectsTool) ToolCallbacks(queue *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		tool.lastCursorPos = &sdl.Point{X: x, Y: y}
		tool.rects = append(
			tool.rects,
			rect{
				sdlRect:         &sdl.Rect{X: x, Y: y, W: 1, H: 1},
				borderThickness: tool.rectBorderThickness,
				color:           tool.rectColor,
			},
		)
		tool.isDragging = true
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		if tool.isDragging {
			rect := tool.rects[len(tool.rects)-1].sdlRect
			rect.W = x - rect.X
			rect.H = y - rect.Y
			if tool.isShiftPressed {
				utils.RectIntoSquare(rect)
			}
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
		queue.Push(RectAction{tool: tool, lastRect: tool.rects[len(tool.rects)-1]})
		return false
	})

	callbacks.KeyDown = append(callbacks.KeyDown, func(keysym sdl.Keysym) bool {
		if keysym.Sym != sdl.K_LSHIFT && keysym.Sym != sdl.K_RSHIFT {
			return false
		}
		if tool.isDragging {
			utils.RectIntoSquare(tool.rects[len(tool.rects)-1].sdlRect)
		}
		tool.isShiftPressed = true
		return false
	})

	callbacks.KeyUp = append(callbacks.KeyUp, func(keysym sdl.Keysym) bool {
		if keysym.Sym != sdl.K_LSHIFT && keysym.Sym != sdl.K_RSHIFT {
			return false
		}
		if tool.isDragging {
			rect := tool.rects[len(tool.rects)-1].sdlRect
			rect.W = tool.lastCursorPos.X - rect.X
			rect.H = tool.lastCursorPos.Y - rect.Y
		}
		tool.isShiftPressed = false
		return false
	})

	return callbacks
}

func (tool RectsTool) RenderCurrentState(ren *sdl.Renderer) {
	for _, rect := range tool.rects {
		utils.DrawThickRectangle(ren, rect.sdlRect, rect.borderThickness, rect.color)
	}
}

func (tool RectsTool) RenderScreenshot(ren *sdl.Renderer) {
	tool.RenderCurrentState(ren)
}

func (tool RectsTool) ToolSettings() []settings.ToolSetting {
	return tool.settings
}

func (tool RectsTool) ToolColor() *sdl.Color {
	return &tool.rectColor
}

type rect struct {
	sdlRect         *sdl.Rect
	borderThickness int32
	color           sdl.Color
}

type RectAction struct {
	tool     *RectsTool
	lastRect rect
}

func (action RectAction) Undo() {
	action.tool.rects = action.tool.rects[:len(action.tool.rects)-1]
}

func (action RectAction) Redo() {
	action.tool.rects = append(action.tool.rects, action.lastRect)
}
