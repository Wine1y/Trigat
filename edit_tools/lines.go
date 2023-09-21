package editTools

import (
	_ "embed"
	"math"

	"github.com/Wine1y/trigat/gui"
	"github.com/Wine1y/trigat/utils"
	"github.com/veandco/go-sdl2/sdl"
)

var lineColor = sdl.Color{R: 255, G: 0, B: 0, A: 255}
var lineThickness int32 = 4

//go:embed icons/line_tool.png
var lineIconData []byte
var lineIcon = utils.LoadPNGSurface(lineIconData)

type LinesTool struct {
	isDragging     bool
	isShiftPressed bool
	lines          [][2]sdl.Point
	lastCursorPos  *sdl.Point
}

func NewLinesTool() *LinesTool {
	return &LinesTool{
		isDragging:     false,
		isShiftPressed: false,
		lines:          make([][2]sdl.Point, 0, 1),
	}
}

func (tool LinesTool) ToolIcon() *sdl.Surface {
	return lineIcon
}

func (tool *LinesTool) ToolCallbacks(queue *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()
	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		tool.lastCursorPos = &sdl.Point{X: x, Y: y}
		tool.isDragging = true
		newLine := [2]sdl.Point{{X: x, Y: y}, {X: x, Y: y}}
		tool.lines = append(tool.lines, newLine)
		return false
	})

	callbacks.MouseUp = append(callbacks.MouseUp, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT || !tool.isDragging {
			return false
		}
		tool.lastCursorPos = &sdl.Point{X: x, Y: y}
		tool.isDragging = false
		line := &tool.lines[len(tool.lines)-1]
		if tool.isShiftPressed {
			line[1] = closestStraightLinePoint(line[0], sdl.Point{X: x, Y: y})
		} else {
			line[1] = sdl.Point{X: x, Y: y}
		}
		queue.Push(LineAction{tool: tool, lastLine: tool.lines[len(tool.lines)-1]})
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		if tool.isDragging {
			tool.lastCursorPos = &sdl.Point{X: x, Y: y}
			line := &tool.lines[len(tool.lines)-1]
			if tool.isShiftPressed {
				line[1] = closestStraightLinePoint(line[0], sdl.Point{X: x, Y: y})
			} else {
				line[1] = sdl.Point{X: x, Y: y}
			}
		}
		return false
	})

	callbacks.KeyDown = append(callbacks.KeyDown, func(keysym sdl.Keysym) bool {
		if keysym.Sym != sdl.K_LSHIFT && keysym.Sym != sdl.K_RSHIFT {
			return false
		}
		tool.isShiftPressed = true
		if tool.isDragging {
			line := &tool.lines[len(tool.lines)-1]
			line[1] = closestStraightLinePoint(line[0], line[1])
		}
		return false
	})

	callbacks.KeyUp = append(callbacks.KeyUp, func(keysym sdl.Keysym) bool {
		if keysym.Sym != sdl.K_LSHIFT && keysym.Sym != sdl.K_RSHIFT {
			return false
		}
		tool.isShiftPressed = false
		if tool.isDragging && tool.lastCursorPos != nil {
			line := &tool.lines[len(tool.lines)-1]
			line[1] = *tool.lastCursorPos
		}
		return false
	})
	return callbacks
}

func (tool LinesTool) RenderCurrentState(ren *sdl.Renderer) {
	for _, line := range tool.lines {
		utils.DrawThickLine(ren, &line[0], &line[1], lineThickness, lineColor)
	}
}

func (tool LinesTool) RenderScreenshot(ren *sdl.Renderer) {
	tool.RenderCurrentState(ren)
}

func closestStraightLinePoint(start sdl.Point, current sdl.Point) sdl.Point {
	vertical := sdl.Point{X: start.X, Y: current.Y}
	horizontal := sdl.Point{X: current.X, Y: start.Y}
	diagonalLength := int32((utils.Abs(int(start.Y-vertical.Y)) + utils.Abs(int(start.X-horizontal.X))) / 2)
	var diagonalX, diagonalY int32
	if current.X > start.X {
		diagonalX = start.X + diagonalLength
	} else {
		diagonalX = start.X - diagonalLength
	}
	if current.Y > start.Y {
		diagonalY = start.Y + diagonalLength
	} else {
		diagonalY = start.Y - diagonalLength
	}
	diagonal := sdl.Point{X: diagonalX, Y: diagonalY}

	var currentPoint *sdl.Point = nil
	var currentDistance *float64 = nil

	for _, point := range []*sdl.Point{&vertical, &horizontal, &diagonal} {
		distance := math.Sqrt(
			math.Pow(float64(current.X-point.X), 2) +
				math.Pow(float64(current.Y-point.Y), 2),
		)
		if currentDistance == nil || distance < *currentDistance {
			currentPoint = point
			currentDistance = &distance
		}
	}
	return *currentPoint
}

type LineAction struct {
	tool     *LinesTool
	lastLine [2]sdl.Point
}

func (action LineAction) Undo() {
	action.tool.lines = action.tool.lines[:len(action.tool.lines)-1]
}

func (action LineAction) Redo() {
	action.tool.lines = append(action.tool.lines, action.lastLine)
}
