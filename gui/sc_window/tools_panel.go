package scWindow

import (
	"github.com/Wine1y/trigat/gui"
	editTools "github.com/Wine1y/trigat/gui/sc_window/edit_tools"
	"github.com/Wine1y/trigat/utils"
	"github.com/veandco/go-sdl2/sdl"
)

const iconSize int32 = 32
const iconPadding int32 = 16
const iconMargin int32 = 10
const panelPadding int32 = 6
const panelSeparatorWidth int32 = 1

const panelTopMargin int32 = 20
const panelRoundingRadius int32 = 8
const toolSize int32 = iconSize + (iconPadding * 2)
const toolColorWidth int32 = toolSize / 4 * 3
const toolColorHeight int32 = 2
const toolColorPadding int32 = 6

const settingsWidth int32 = 100
const settingsShowDelayMs uint64 = 650

var panelBackgroundColor = sdl.Color{R: 115, G: 115, B: 115, A: 150}
var panelActiveToolColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var panelHoverToolColor = sdl.Color{R: 100, G: 100, B: 100, A: 200}
var panelSeparatorColor = sdl.Color{R: 170, G: 170, B: 170, A: 255}

type ToolsPanel struct {
	tools        []*toolMeta
	currentTool  *toolMeta
	hoveredTool  *toolMeta
	hoveredAt    uint64
	cropTool     editTools.ScreenshotCropTool
	actionsQueue *editTools.ActionsQueue
	panelRect    *sdl.Rect
}

func NewToolsPanel(ren *sdl.Renderer) *ToolsPanel {
	selectionTool := editTools.NewSelectionTool(ren)
	tools := []editTools.ScreenshotEditTool{
		selectionTool,
		editTools.NewPaintTool(),
		editTools.NewLinesTool(),
		editTools.NewRectsTool(),
		editTools.NewTextTool(ren),
		editTools.NewPipetteTool(ren),
	}
	metas := make([]*toolMeta, len(tools))
	for i, tool := range tools {
		meta := newToolMeta(tool, ren)
		metas[i] = &meta
	}
	panel := ToolsPanel{
		tools:        metas,
		actionsQueue: editTools.NewActionsQueue(),
		cropTool:     selectionTool,
	}
	if len(metas) > 0 {
		panel.currentTool = metas[0]
	}
	vp := ren.GetViewport()
	panel.resizePanel(vp.W, vp.H)
	return &panel
}

func (panel ToolsPanel) CurrentTool() editTools.ScreenshotEditTool {
	return panel.currentTool.tool
}

func (panel ToolsPanel) DrawToolsState(ren *sdl.Renderer) {
	for _, meta := range panel.tools {
		meta.tool.RenderCurrentState(ren)
	}
}

func (panel ToolsPanel) RenderScreenshot(ren *sdl.Renderer) {
	for _, meta := range panel.tools {
		meta.tool.RenderScreenshot(ren)
	}
}

func (panel ToolsPanel) CropScreenshot(surface *sdl.Surface) *sdl.Surface {
	if panel.cropTool != nil {
		return panel.cropTool.CropScreenshot(surface)
	}
	return surface
}

func (panel *ToolsPanel) UndoLastAction() {
	if panel.actionsQueue.CanUndo() {
		panel.actionsQueue.Undo()
	}
}

func (panel *ToolsPanel) RedoLastAction() {
	if panel.actionsQueue.CanRedo() {
		panel.actionsQueue.Redo()
	}
}

func (panel ToolsPanel) DrawPanel(ren *sdl.Renderer) {
	utils.DrawRoundedFilledRectangle(ren, panel.panelRect, panelRoundingRadius, panelBackgroundColor)
	for i, meta := range panel.tools {
		meta.texture.SetColorMod(255, 255, 255)
		if meta == panel.hoveredTool {
			utils.DrawRoundedFilledRectangle(ren, &meta.toolBBox, panelRoundingRadius, panelHoverToolColor)
			toolSettings := panel.hoveredTool.tool.ToolSettings()
			if len(toolSettings) > 0 && sdl.GetTicks64()-panel.hoveredAt >= settingsShowDelayMs {
				for _, setting := range toolSettings {
					setting.Render(ren)
				}
			}
		}
		if meta == panel.currentTool {
			utils.DrawRoundedFilledRectangle(ren, &meta.toolBBox, panelRoundingRadius, panelActiveToolColor)
			meta.texture.SetColorMod(0, 0, 0)
		}
		if toolColor := meta.tool.ToolColor(); toolColor != nil {
			utils.DrawRoundedFilledRectangle(
				ren,
				meta.colorBBox,
				2,
				*toolColor,
			)
		}
		if i != len(panel.tools)-1 {
			utils.DrawThickLine(
				ren,
				&sdl.Point{X: meta.toolBBox.X + meta.toolBBox.W + (iconMargin / 2), Y: meta.toolBBox.Y + iconPadding},
				&sdl.Point{X: meta.toolBBox.X + meta.toolBBox.W + (iconMargin / 2), Y: meta.toolBBox.Y + meta.toolBBox.H - iconPadding},
				panelSeparatorWidth, panelSeparatorColor,
			)
		}
		utils.CopyTexture(ren, meta.texture, &meta.iconBBox, nil)
	}
}

func (panel *ToolsPanel) SetToolsCallbacks(callbacks *gui.WindowCallbackSet) {
	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		click := sdl.Point{X: x, Y: y}
		for _, meta := range panel.tools {
			if click.InRect(&meta.toolBBox) {
				panel.currentTool = meta
				return true
			}
			if panel.hoveredTool == meta && sdl.GetTicks64()-panel.hoveredAt >= settingsShowDelayMs {
				if click.InRect(&panel.hoveredTool.settingsBBox) {
					panel.currentTool = meta
				}
			}
		}
		return click.InRect(panel.panelRect)
	})
	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		move := sdl.Point{X: x, Y: y}
		for _, meta := range panel.tools {
			if move.InRect(&meta.toolBBox) {
				if panel.hoveredTool != meta {
					panel.hoveredAt = sdl.GetTicks64()
				}
				panel.hoveredTool = meta
				sdl.SetCursor(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_HAND))
				return false
			}
		}
		if panel.hoveredTool != nil {
			sdl.SetCursor(sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW))
			if move.InRect(&panel.hoveredTool.settingsBBox) && sdl.GetTicks64()-panel.hoveredAt >= settingsShowDelayMs {
				return false
			}
			panel.hoveredTool = nil
		}
		return false
	})

	callbacks.KeyDown = append(callbacks.KeyDown, func(keysym sdl.Keysym) bool {
		switch {
		case keysym.Sym == sdl.K_z && (keysym.Mod&sdl.KMOD_CTRL != 0 && keysym.Mod&sdl.KMOD_ALT != 0):
			panel.RedoLastAction()
		case keysym.Sym == sdl.K_z && (keysym.Mod&sdl.KMOD_CTRL != 0):
			panel.UndoLastAction()
		}
		return false
	})

	callbacks.SizeChange = append(callbacks.SizeChange, func(w, h int32) bool {
		panel.resizePanel(w, h)
		return false
	})
	callbacks.Quit = append(callbacks.Quit, func() bool {
		for _, meta := range panel.tools {
			meta.texture.Destroy()
		}
		return false
	})

	if panel.hoveredTool != nil && sdl.GetTicks64()-panel.hoveredAt >= settingsShowDelayMs {
		for _, setting := range panel.hoveredTool.tool.ToolSettings() {
			callbacks.Append(setting.SettingCallbacks())
		}
	}
	if panel.currentTool != nil {
		callbacks.Append(panel.currentTool.tool.ToolCallbacks(panel.actionsQueue))
	}
}

func (panel *ToolsPanel) resizePanel(viewportW, viewportH int32) {
	panelWidth := (toolSize * int32(len(panel.tools))) + (iconMargin * int32(len(panel.tools)-1)) + (panelPadding * 2)
	panelRect := sdl.Rect{
		X: (viewportW - panelWidth) / 2, Y: panelTopMargin,
		W: panelWidth, H: toolSize + (panelPadding * 2),
	}
	panel.panelRect = &panelRect
	for i, meta := range panel.tools {
		meta.iconBBox = sdl.Rect{
			X: panelRect.X + panelPadding + iconPadding + ((iconMargin + toolSize) * int32(i)),
			Y: panelRect.Y + panelPadding + iconPadding,
			W: iconSize,
			H: iconSize,
		}
		meta.toolBBox = sdl.Rect{
			X: meta.iconBBox.X - iconPadding,
			Y: meta.iconBBox.Y - iconPadding,
			W: toolSize, H: toolSize,
		}
		if meta.tool.ToolColor() != nil {
			meta.colorBBox = &sdl.Rect{
				X: meta.toolBBox.X + (meta.toolBBox.W-toolColorWidth)/2,
				Y: meta.toolBBox.Y + meta.toolBBox.H - toolColorHeight - toolColorPadding,
				W: toolColorWidth, H: toolColorHeight,
			}
		}
		var toolSettingsH int32 = 0
		for _, setting := range meta.tool.ToolSettings() {
			setting.SetLeftTop(&sdl.Point{
				X: meta.toolBBox.X - (settingsWidth-meta.toolBBox.W)/2,
				Y: meta.toolBBox.Y + meta.toolBBox.H + panelPadding + toolSettingsH,
			})
			setting.SetWidth(settingsWidth)
			toolSettingsH += setting.BBox().H
		}
		meta.settingsBBox = sdl.Rect{
			X: meta.toolBBox.X - (settingsWidth-meta.toolBBox.W)/2,
			Y: meta.toolBBox.Y + meta.toolBBox.H,
			W: settingsWidth, H: toolSettingsH,
		}
	}
}

type toolMeta struct {
	tool         editTools.ScreenshotEditTool
	iconBBox     sdl.Rect
	toolBBox     sdl.Rect
	colorBBox    *sdl.Rect
	settingsBBox sdl.Rect
	texture      *sdl.Texture
}

func newToolMeta(tool editTools.ScreenshotEditTool, ren *sdl.Renderer) toolMeta {
	texture, err := ren.CreateTextureFromSurface(tool.ToolIcon())
	if err != nil {
		panic(err)
	}
	return toolMeta{
		tool:     tool,
		iconBBox: sdl.Rect{},
		texture:  texture,
	}
}
