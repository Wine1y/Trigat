package editTools

import (
	_ "embed"
	"fmt"

	"github.com/Wine1y/trigat/assets"
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/pkg"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const selectionThickness int32 = 2
const selectionTooltipMargin int32 = 4
const selectionTooltipPadding int32 = 4
const actionIconSize int32 = 16
const actionMargin int32 = 4
const selectionTooltipBackroundCornerRadius int32 = 4

var selectionTooltipBackgroundColor = sdl.Color{R: 0, G: 0, B: 0, A: 130}
var selectionTooltipForegroundColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var selectionBorderColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var selectionFillColor = sdl.Color{R: 255, G: 255, B: 255, A: 50}

type SelectionTool struct {
	ren            *sdl.Renderer
	isDragging     bool
	isShiftPressed bool
	selection      *sdl.Rect
	lastCursorPos  sdl.Point
	sizeTooltip    *selectionSizeTooltip
	actionsTooltip *selectionActionsTooltip
	handCursorSet  bool
	DefaultScreenshotEditTool
}

func NewSelectionTool(renderer *sdl.Renderer, saveCallback, copyCallback, searchCallback func()) *SelectionTool {
	return &SelectionTool{
		isDragging:     false,
		isShiftPressed: false,
		sizeTooltip:    &selectionSizeTooltip{font: assets.GetAppFont(14)},
		actionsTooltip: NewSelectionActionsTooltip(renderer, saveCallback, copyCallback, searchCallback),
		ren:            renderer,
	}
}

func (tool SelectionTool) ToolIcon() *sdl.Surface {
	return assets.SelectionIcon
}

func (tool *SelectionTool) OnToolDeactivated() {
	tool.isShiftPressed = false
	tool.isDragging = false
}

func (tool *SelectionTool) ToolCallbacks(_ *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		if action, actionHovered := tool.actionsTooltip.getActionAt(x, y); actionHovered {
			action.callback()
			return false
		}
		tool.selection = &sdl.Rect{X: x, Y: y, W: 1, H: 1}
		tool.updateTooltips()
		tool.isDragging = true
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		tool.lastCursorPos.X, tool.lastCursorPos.Y = x, y
		if tool.isDragging {
			sel := tool.selection
			sel.W = x - sel.X
			sel.H = y - sel.Y
			if tool.isShiftPressed {
				pkg.RectIntoSquare(sel)
			}
			tool.updateTooltips()
			return false
		}
		if _, actionHovered := tool.actionsTooltip.getActionAt(x, y); actionHovered {
			sdl.SetCursor(gui.HandCursor)
			tool.handCursorSet = true
		} else if tool.handCursorSet {
			sdl.SetCursor(gui.ArrowCursor)
			tool.handCursorSet = false
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

	callbacks.KeyDown = append(callbacks.KeyDown, func(keysym sdl.Keysym) bool {
		if keysym.Sym == sdl.K_LSHIFT || keysym.Sym == sdl.K_RSHIFT {
			if tool.isDragging {
				pkg.RectIntoSquare(tool.selection)
				tool.updateTooltips()
			}
			tool.isShiftPressed = true
		}
		if keysym.Sym == sdl.K_a && (keysym.Mod&sdl.KMOD_CTRL != 0) {
			vp := tool.ren.GetViewport()
			tool.selection = &vp
			tool.updateTooltips()
		}

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
			tool.updateTooltips()
		}
		tool.isShiftPressed = false
		return false
	})

	callbacks.Quit = append(callbacks.Quit, func() bool {
		tool.destroyTooltips()
		return false
	})

	return callbacks
}

func (tool *SelectionTool) updateTooltips() {
	tool.sizeTooltip.updateTooltip(tool.ren, tool.selection)
	tool.actionsTooltip.updateTooltip(tool.selection)
}

func (tool *SelectionTool) destroyTooltips() {
	tool.sizeTooltip.destroy()
	tool.actionsTooltip.destroy()
}

func (tool SelectionTool) RenderCurrentState(ren *sdl.Renderer) {
	if tool.selection != nil {
		sel := tool.selection
		pkg.DrawFilledRectangle(ren, sel, selectionFillColor)
		pkg.DrawThickRectangle(ren, sel, selectionThickness, selectionBorderColor)
		tool.sizeTooltip.draw(ren)
	}
	tool.actionsTooltip.draw(ren)
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

type selectionSizeTooltip struct {
	texture *pkg.StringTexture
	bbox    sdl.Rect
	font    *ttf.Font
}

func (tooltip *selectionSizeTooltip) updateTooltip(ren *sdl.Renderer, selection *sdl.Rect) {
	text := fmt.Sprintf("%v x %v", pkg.Abs(selection.W), pkg.Abs(selection.H))
	textW, textH := pkg.SizeString(tooltip.font, text)

	tooltip.bbox.W = int32(textW) + selectionTooltipPadding*2
	tooltip.bbox.H = int32(textH) + selectionTooltipPadding*2

	tooltip.bbox.X = selection.X
	if selection.W < 0 {
		tooltip.bbox.X = selection.X + selection.W
	}

	tooltip.bbox.Y = selection.Y + selection.H + selectionTooltipMargin
	if selection.H < 0 {
		tooltip.bbox.Y = selection.Y + selectionTooltipMargin
	}

	vp := ren.GetViewport()
	if tooltip.bbox.Y+tooltip.bbox.H > vp.H {
		tooltip.bbox.Y -= (tooltip.bbox.H + selectionTooltipMargin*2 + selectionThickness)
		tooltip.bbox.X += (selectionThickness + selectionTooltipMargin)
	}
	tooltip.texture = pkg.NewStringTexture(ren, tooltip.font, text, selectionTooltipForegroundColor)
}

func (tooltip *selectionSizeTooltip) draw(ren *sdl.Renderer) {
	if tooltip.texture != nil {
		pkg.DrawRoundedFilledRectangle(
			ren,
			&tooltip.bbox,
			selectionTooltipBackroundCornerRadius,
			selectionTooltipBackgroundColor,
		)
		tooltip.texture.Draw(
			ren,
			&sdl.Point{
				X: tooltip.bbox.X + selectionTooltipPadding,
				Y: tooltip.bbox.Y + selectionTooltipPadding,
			},
		)
	}
}

func (tooltip *selectionSizeTooltip) destroy() {
	if tooltip.texture != nil {
		tooltip.texture.Destroy()
	}
	tooltip.font.Close()
}

type selectionActionsTooltip struct {
	actions     []*tooltipAction
	bbox        sdl.Rect
	inSelection bool
}

func NewSelectionActionsTooltip(ren *sdl.Renderer, saveCallback, copyCallback, searchCallback func()) *selectionActionsTooltip {
	tooltip := selectionActionsTooltip{
		actions: []*tooltipAction{
			{texture: pkg.CreateTextureFromSurface(ren, assets.SearchIcon), callback: searchCallback},
			{texture: pkg.CreateTextureFromSurface(ren, assets.CopyIcon), callback: copyCallback},
			{texture: pkg.CreateTextureFromSurface(ren, assets.SaveIcon), callback: saveCallback},
		},
	}

	vp := ren.GetViewport()
	tooltip.bbox.W = int32(len(tooltip.actions))*actionIconSize + int32(len(tooltip.actions)-1)*actionMargin + selectionTooltipPadding*2
	tooltip.bbox.H = actionIconSize + selectionTooltipPadding*2
	tooltip.bbox.X = vp.W - tooltip.bbox.W - selectionTooltipMargin
	tooltip.bbox.Y = selectionTooltipMargin
	tooltip.updateActionsPositions()
	return &tooltip
}

func (tooltip *selectionActionsTooltip) updateTooltip(selection *sdl.Rect) {
	tooltip.bbox.X = selection.X + selection.W - tooltip.bbox.W
	if selection.W < 0 {
		tooltip.bbox.X = selection.X - tooltip.bbox.W
	}

	tooltip.bbox.Y = selection.Y - tooltip.bbox.H - selectionTooltipMargin
	if selection.H < 0 {
		tooltip.bbox.Y = selection.Y + selection.H - tooltip.bbox.H - selectionTooltipMargin
	}

	if tooltip.bbox.Y < 0 {
		tooltip.bbox.Y += (tooltip.bbox.H + selectionTooltipMargin*2 + selectionThickness)
		tooltip.bbox.X -= (selectionThickness + selectionTooltipMargin)
		tooltip.inSelection = true
	} else {
		tooltip.inSelection = false
	}
	tooltip.updateActionsPositions()
}

func (tooltip *selectionActionsTooltip) updateActionsPositions() {
	for i := int32(0); i < int32(len(tooltip.actions)); i++ {
		action := tooltip.actions[i]
		action.bbox.X = tooltip.bbox.X + i*actionIconSize + i*actionMargin + selectionTooltipPadding
		action.bbox.Y = tooltip.bbox.Y + selectionTooltipPadding
		action.bbox.W, action.bbox.H = actionIconSize, actionIconSize
	}
}

func (tooltip *selectionActionsTooltip) draw(ren *sdl.Renderer) {
	pkg.DrawRoundedFilledRectangle(
		ren,
		&sdl.Rect{
			X: tooltip.bbox.X, Y: tooltip.bbox.Y,
			W: tooltip.bbox.W, H: tooltip.bbox.H,
		},
		selectionTooltipBackroundCornerRadius,
		selectionTooltipBackgroundColor,
	)
	for _, action := range tooltip.actions {
		pkg.CopyTexture(ren, action.texture, &action.bbox, nil)
	}
}

func (tooltip selectionActionsTooltip) getActionAt(x, y int32) (*tooltipAction, bool) {
	point := sdl.Point{X: x, Y: y}
	for _, action := range tooltip.actions {
		if point.InRect(&action.bbox) {
			return action, true
		}
	}
	return nil, false
}

func (tooltip *selectionActionsTooltip) destroy() {
	for _, action := range tooltip.actions {
		if action.texture != nil {
			action.texture.Destroy()
		}
	}
}

type tooltipAction struct {
	texture  *sdl.Texture
	bbox     sdl.Rect
	callback func()
}
