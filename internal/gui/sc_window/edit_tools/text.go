package editTools

import (
	_ "embed"
	"time"
	"unicode"

	"github.com/Wine1y/trigat/assets"
	"github.com/Wine1y/trigat/config"
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/internal/gui/sc_window/settings"
	"github.com/Wine1y/trigat/pkg"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var cursorColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var selectionColor = sdl.Color{R: 0, G: 0, B: 0, A: 100}
var paragraphBoundariesColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}

const paragraphPadding int32 = 5
const paragraphDraggingPadding int32 = 5
const cursorAnimationDuration time.Duration = time.Millisecond * 1250

type TextTool struct {
	paragraphs       []*pkg.TextParagraph
	activeParagraph  *pkg.TextParagraph
	ren              *sdl.Renderer
	textFont         *ttf.Font
	textColor        sdl.Color
	settings         []settings.ToolSetting
	cursorPos        int
	cursorAnimation  *pkg.Animation
	isShiftSelecting bool
	isMouseSelecting bool
	selection        textSelection
	draggingHandle   textDraggingHandle
	iBeamCursorSet   bool
	sizeAllCursorSet bool
	DefaultScreenshotEditTool
}

func NewTextTool(renderer *sdl.Renderer) *TextTool {
	tool := TextTool{
		paragraphs:     make([]*pkg.TextParagraph, 0),
		ren:            renderer,
		textFont:       assets.GetAppFont(14),
		selection:      textSelection{start: 0, length: 0, selected: false},
		draggingHandle: textDraggingHandle{draggingParagraph: nil, xHandleOffset: 0, yHandleOffset: 0},
	}

	colorPicker := settings.NewColorPickerSetting(func(color sdl.Color) {
		tool.textColor = color
		if tool.activeParagraph != nil {
			tool.activeParagraph.SetColor(color)
		}
	})
	toolSettings := []settings.ToolSetting{colorPicker}
	tool.textColor = colorPicker.CurrentColor()
	tool.settings = toolSettings
	return &tool
}

func (tool TextTool) ToolIcon() *sdl.Surface {
	return assets.TextIcon
}

func (tool *TextTool) ToolCallbacks(queue *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		click := sdl.Point{X: x, Y: y}
		for _, par := range tool.paragraphs {
			bbox := par.GetBBox()
			if click.InRect(bbox) {
				tool.deselectText()
				tool.activeParagraph = par
				tool.moveCursor(par.GetPositionByOffset(x-par.TextStart.X, y-par.TextStart.Y))
				tool.isMouseSelecting = true
				return false
			}
			if click.InRect(par.GetPaddedBBox(paragraphDraggingPadding)) {
				tool.draggingHandle.xHandleOffset = click.X - bbox.X
				tool.draggingHandle.yHandleOffset = click.Y - bbox.Y
				tool.draggingHandle.draggingParagraph = par
				return false
			}

		}
		newParagraph := pkg.NewTextParagraph(
			sdl.Point{X: x, Y: y},
			tool.textColor,
			tool.textFont,
			paragraphPadding,
		)
		tool.paragraphs = append(tool.paragraphs, newParagraph)
		tool.activeParagraph = newParagraph
		tool.cursorAnimation = pkg.NewLinearAnimation(255, 0, int(config.GetAppFPS()), cursorAnimationDuration, 0, true)
		tool.deselectText()
		tool.moveCursor(0)
		queue.Push(textParagraphCreatedAction{tool: tool, lastParagraph: newParagraph})
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		move := sdl.Point{X: x, Y: y}
		if tool.isMouseSelecting && !move.InRect(tool.activeParagraph.GetBBox()) {
			tool.isMouseSelecting = false
		}
		if tool.isMouseSelecting {
			tool.moveCursor(
				tool.activeParagraph.GetPositionByOffset(
					x-tool.activeParagraph.TextStart.X,
					y-tool.activeParagraph.TextStart.Y,
				),
			)
			sdl.SetCursor(gui.IBeamCursor)
			tool.iBeamCursorSet = true
			return false
		}
		if tool.draggingHandle.draggingParagraph != nil {
			tool.draggingHandle.draggingParagraph.TextStart = sdl.Point{
				X: move.X - tool.draggingHandle.xHandleOffset,
				Y: move.Y - tool.draggingHandle.yHandleOffset,
			}
			sdl.SetCursor(gui.SizeAllCursor)
			tool.sizeAllCursorSet = true
			return false
		}
		for _, par := range tool.paragraphs {
			if move.InRect(par.GetBBox()) {
				sdl.SetCursor(gui.IBeamCursor)
				tool.iBeamCursorSet = true
				return false
			}
			if move.InRect(par.GetPaddedBBox(paragraphDraggingPadding)) {
				sdl.SetCursor(gui.SizeAllCursor)
				tool.sizeAllCursorSet = true
				return false
			}
		}

		if tool.iBeamCursorSet || tool.sizeAllCursorSet {
			if tool.iBeamCursorSet {
				tool.iBeamCursorSet = false
			}
			if tool.sizeAllCursorSet {
				tool.sizeAllCursorSet = false
			}
			sdl.SetCursor(gui.ArrowCursor)
		}
		return false
	})

	callbacks.MouseUp = append(callbacks.MouseUp, func(button uint8, x, y int32) bool {
		if button == sdl.BUTTON_LEFT {
			if tool.isMouseSelecting {
				tool.isMouseSelecting = false
			}
			if tool.draggingHandle.draggingParagraph != nil {
				tool.draggingHandle.draggingParagraph = nil
			}
		}
		return false
	})

	callbacks.TextInput = append(callbacks.TextInput, func(rn rune) bool {
		if tool.activeParagraph == nil || !unicode.IsGraphic(rn) {
			return false
		}
		var newCursorPos int
		if tool.selection.selected {
			selStart, selEnd := tool.selection.selectionBounds()
			tool.replaceInParagraph(tool.activeParagraph, selStart, selEnd, queue, rn)
			newCursorPos = selStart + 1
		} else {
			tool.insertIntoParagraph(tool.activeParagraph, tool.cursorPos, queue, rn)
			newCursorPos = tool.cursorPos + 1
		}
		tool.moveCursorIgnoreSelection(newCursorPos)
		return false
	})

	callbacks.KeyDown = append(callbacks.KeyDown, func(keysym sdl.Keysym) bool {
		if tool.activeParagraph == nil {
			return false
		}
		activePar := tool.activeParagraph

		switch {
		case keysym.Sym == sdl.K_LSHIFT || keysym.Sym == sdl.K_RSHIFT:
			tool.isShiftSelecting = true
		case (keysym.Sym == sdl.K_BACKSPACE || keysym.Sym == sdl.K_DELETE) && tool.selection.selected:
			selStart, selEnd := tool.selection.selectionBounds()
			tool.moveCursor(selStart)
			tool.popFromParagraph(activePar, selStart, selEnd, queue)
			tool.deselectText()
		case keysym.Sym == sdl.K_BACKSPACE && tool.cursorPos > 0:
			tool.popFromParagraph(activePar, tool.cursorPos-1, tool.cursorPos, queue)
			tool.moveCursorIgnoreSelection(tool.cursorPos - 1)
		case keysym.Sym == sdl.K_DELETE && tool.cursorPos < len(activePar.Text):
			tool.popFromParagraph(activePar, tool.cursorPos, tool.cursorPos+1, queue)
		case keysym.Sym == sdl.K_LEFT && (keysym.Mod&sdl.KMOD_CTRL != 0):
			tool.moveCursor(activePar.ClosestLeftWordPos(tool.cursorPos))
		case keysym.Sym == sdl.K_RIGHT && (keysym.Mod&sdl.KMOD_CTRL != 0):
			tool.moveCursor(activePar.ClosestRightWordPos(tool.cursorPos))
		case keysym.Sym == sdl.K_LEFT:
			tool.moveCursor(tool.cursorPos - 1)
		case keysym.Sym == sdl.K_RIGHT:
			tool.moveCursor(tool.cursorPos + 1)
		case keysym.Sym == sdl.K_UP:
			tool.moveCursor(activePar.UpperLinePos(tool.cursorPos))
		case keysym.Sym == sdl.K_DOWN:
			tool.moveCursor(activePar.LowerLinePos(tool.cursorPos))
		case keysym.Sym == sdl.K_HOME:
			lines := activePar.GetLinesBoundaries()
			currentLine := lines[activePar.GetLineNumber(tool.cursorPos)]
			tool.moveCursor(currentLine[0])
		case keysym.Sym == sdl.K_END:
			lines := activePar.GetLinesBoundaries()
			currentLine := lines[activePar.GetLineNumber(tool.cursorPos)]
			tool.moveCursor(currentLine[1])
		case keysym.Sym == sdl.K_RETURN:
			tool.insertIntoParagraph(activePar, tool.cursorPos, queue, '\n')
			tool.moveCursor(tool.cursorPos + 1)
		case keysym.Sym == sdl.K_a && (keysym.Mod&sdl.KMOD_CTRL != 0):
			tool.moveCursor(len(activePar.Text))
			tool.selectText(0, len(activePar.Text))
		case keysym.Sym == sdl.K_x && (keysym.Mod&sdl.KMOD_CTRL != 0) && tool.selection.selected:
			selStart, selEnd := tool.selection.selectionBounds()
			sdl.SetClipboardText(string(activePar.Text[selStart:selEnd]))
			tool.popFromParagraph(activePar, selStart, selEnd, queue)
			tool.moveCursor(selStart)
			tool.deselectText()
		case keysym.Sym == sdl.K_c && (keysym.Mod&sdl.KMOD_CTRL != 0) && tool.selection.selected:
			selStart, selEnd := tool.selection.selectionBounds()
			sdl.SetClipboardText(string(activePar.Text[selStart:selEnd]))
			return true
		case keysym.Sym == sdl.K_v && (keysym.Mod&sdl.KMOD_CTRL != 0):
			cbString, err := sdl.GetClipboardText()
			if err != nil {
				break
			}
			text := []rune(cbString)
			var newCursorPos int
			if tool.selection.selected {
				selStart, selEnd := tool.selection.selectionBounds()
				tool.replaceInParagraph(activePar, selStart, selEnd, queue, text...)
				newCursorPos = selStart + len(text)
			} else {
				tool.insertIntoParagraph(activePar, tool.cursorPos, queue, []rune(text)...)
				newCursorPos = tool.cursorPos + len(text)
			}
			tool.moveCursor(newCursorPos)
		}
		return false
	})

	callbacks.KeyUp = append(callbacks.KeyUp, func(keysym sdl.Keysym) bool {
		if keysym.Sym == sdl.K_LSHIFT || keysym.Sym == sdl.K_RSHIFT {
			tool.isShiftSelecting = false
		}
		return false
	})

	callbacks.Quit = append(callbacks.Quit, func() bool {
		for _, par := range tool.paragraphs {
			if par.StringTexture != nil {
				par.StringTexture.Destroy()
			}
		}
		tool.textFont.Close()
		return false
	})

	return callbacks
}

func (tool TextTool) ToolSettings() []settings.ToolSetting {
	return tool.settings
}

func (tool TextTool) ToolColor() *sdl.Color {
	return &tool.textColor
}

func (tool *TextTool) OnToolDeactivated() {
	tool.activeParagraph = nil
	tool.draggingHandle.draggingParagraph = nil
	tool.selection.selected = false
	tool.isShiftSelecting = false
	tool.isMouseSelecting = false
}

func (tool *TextTool) moveCursor(newPos int) {
	newPos = pkg.Clamp(0, newPos, len(tool.activeParagraph.Text))
	if tool.isShiftSelecting || tool.isMouseSelecting {
		tool.selectText(tool.cursorPos, newPos)
	} else {
		tool.deselectText()
	}
	tool.cursorPos = newPos
}

func (tool *TextTool) moveCursorIgnoreSelection(newPos int) {
	newPos = pkg.Clamp(0, newPos, len(tool.activeParagraph.Text))
	tool.cursorPos = newPos
}

func (tool *TextTool) selectText(from int, to int) {
	offset := to - from
	if tool.selection.length == 0 {
		tool.selection.start = from
	}
	tool.selection.length += offset
	if tool.selection.length != 0 {
		tool.selection.selected = true
	}
}

func (tool *TextTool) deselectText() {
	tool.selection.selected = false
	tool.selection.length = 0
	tool.selection.start = 0
}

func (tool *TextTool) insertIntoParagraph(par *pkg.TextParagraph, insertAt int, queue *ActionsQueue, text ...rune) {
	queue.Push(
		textInsertedAction{
			tool: tool, ren: tool.ren,
			par: par, text: text,
			insertedAt: insertAt,
		},
	)
	par.InsertRunes(tool.ren, insertAt, text...)
}

func (tool *TextTool) replaceInParagraph(par *pkg.TextParagraph, replaceFrom int, replaceTo int, queue *ActionsQueue, newText ...rune) {
	queue.Push(
		textInsertedAction{
			tool: tool, ren: tool.ren,
			par: par, text: newText,
			replacedText: par.Text[replaceFrom:replaceTo],
			insertedAt:   replaceFrom,
		},
	)
	par.PopRunes(tool.ren, replaceFrom, replaceTo)
	par.InsertRunes(tool.ren, replaceFrom, newText...)
}

func (tool *TextTool) popFromParagraph(par *pkg.TextParagraph, popFrom int, popTo int, queue *ActionsQueue) {
	queue.Push(
		textRemovedAction{
			tool: tool, ren: tool.ren,
			par: par, text: par.Text[popFrom:popTo],
			removedFrom: popFrom,
		},
	)
	par.PopRunes(tool.ren, popFrom, popTo)
}

func (tool TextTool) RenderScreenshot(ren *sdl.Renderer) {
	for _, par := range tool.paragraphs {
		if par.StringTexture != nil {
			par.StringTexture.Draw(ren, &par.TextStart)
		}
	}
}

func (tool TextTool) RenderCurrentState(ren *sdl.Renderer) {
	for _, par := range tool.paragraphs {
		pkg.DrawRectangle(
			ren,
			par.GetBBox(),
			paragraphBoundariesColor,
		)
		if par.StringTexture != nil {
			par.StringTexture.Draw(ren, &par.TextStart)
		}
		if par == tool.activeParagraph {
			tool.renderCursor(ren)
			if tool.selection.selected {
				tool.renderSelection(ren)
			}
		}
	}
}

func (tool TextTool) renderCursor(ren *sdl.Renderer) {
	par := tool.activeParagraph
	xOffset, yOffset := par.GetOffsetByPosition(tool.cursorPos)
	cursorH := par.Font.Height()
	pkg.DrawThickLine(
		ren,
		&sdl.Point{X: par.TextStart.X + xOffset, Y: par.TextStart.Y + yOffset},
		&sdl.Point{X: par.TextStart.X + xOffset, Y: par.TextStart.Y + yOffset + int32(cursorH)},
		1,
		sdl.Color{R: cursorColor.R, G: cursorColor.G, B: cursorColor.B, A: uint8(tool.cursorAnimation.CurrentValue())},
	)
}

func (tool TextTool) renderSelection(ren *sdl.Renderer) {
	par := tool.activeParagraph
	selH := par.Font.LineSkip()
	selStart, selEnd := tool.selection.selectionBounds()
	lines := par.GetLinesBoundaries()
	for i, line := range lines {
		y := int32(i * selH)
		lineSelStart := pkg.Max(selStart, line[0])
		lineSelEnd := pkg.Min(selEnd, line[1]+1)
		if lineSelStart > line[1] || lineSelEnd < line[0] {
			continue
		}
		selOffset := 0
		if lineSelStart > 0 {
			selOffset, _ = pkg.SizeString(par.Font, string(par.Text[line[0]:lineSelStart]))
		}
		selW, _ := pkg.SizeString(par.Font, string(par.Text[lineSelStart:lineSelEnd]))
		pkg.DrawFilledRectangle(
			ren,
			&sdl.Rect{
				X: par.TextStart.X + int32(selOffset),
				Y: par.TextStart.Y + y,
				W: int32(selW), H: int32(selH),
			},
			selectionColor,
		)
	}
}

type textDraggingHandle struct {
	draggingParagraph *pkg.TextParagraph
	xHandleOffset     int32
	yHandleOffset     int32
}

type textSelection struct {
	start    int
	length   int
	selected bool
}

func (sel textSelection) selectionBounds() (int, int) {
	return pkg.Min(sel.start, sel.start+sel.length), pkg.Max(sel.start, sel.start+sel.length)
}

type textParagraphCreatedAction struct {
	tool          *TextTool
	lastParagraph *pkg.TextParagraph
}

func (action textParagraphCreatedAction) Undo() {
	action.tool.paragraphs = action.tool.paragraphs[:len(action.tool.paragraphs)-1]
}

func (action textParagraphCreatedAction) Redo() {
	action.tool.paragraphs = append(action.tool.paragraphs, action.lastParagraph)
}

type textInsertedAction struct {
	tool         *TextTool
	ren          *sdl.Renderer
	par          *pkg.TextParagraph
	text         []rune
	replacedText []rune
	insertedAt   int
}

func (action textInsertedAction) Undo() {
	action.par.PopRunes(
		action.ren,
		action.insertedAt,
		action.insertedAt+len(action.text),
	)
	if len(action.replacedText) > 0 {
		action.par.InsertRunes(action.ren, action.insertedAt, action.replacedText...)
	}
	action.tool.moveCursor(action.insertedAt + len(action.replacedText))
}

func (action textInsertedAction) Redo() {
	if len(action.replacedText) > 0 {
		action.par.PopRunes(action.ren, action.insertedAt, action.insertedAt+len(action.replacedText))
	}
	action.par.InsertRunes(action.ren, action.insertedAt, action.text...)
	action.tool.moveCursor(action.insertedAt + len(action.text))
}

type textRemovedAction struct {
	tool        *TextTool
	ren         *sdl.Renderer
	par         *pkg.TextParagraph
	text        []rune
	removedFrom int
}

func (action textRemovedAction) Undo() {
	action.par.InsertRunes(action.ren, action.removedFrom, action.text...)
	action.tool.moveCursor(action.removedFrom + len(action.text))
}

func (action textRemovedAction) Redo() {
	action.par.PopRunes(action.ren, action.removedFrom, action.removedFrom+len(action.text))
	action.tool.moveCursor(action.removedFrom)
}
