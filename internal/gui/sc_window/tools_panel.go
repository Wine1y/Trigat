package scWindow

import (
	"time"

	"github.com/Wine1y/trigat/internal/gui"
	editTools "github.com/Wine1y/trigat/internal/gui/sc_window/edit_tools"
	"github.com/Wine1y/trigat/pkg"
	"github.com/veandco/go-sdl2/sdl"
)

const panelIconSize int32 = 38
const panelIconPadding int32 = 16
const panelIconMargin int32 = 10
const panelPadding int32 = 6
const panelSeparatorWidth int32 = 1
const panelToolColorOutlineWidth int32 = 1

const panelTopMargin int32 = 20
const panelRoundingRadius int32 = 8
const panelToolSize int32 = panelIconSize + (panelIconPadding * 2)
const panelToolColorWidth int32 = panelToolSize / 4 * 3
const panelToolColorHeight int32 = 3
const panelToolColorPadding int32 = 6

const panelSettingsWidth int32 = 100
const panelSettingsShowDelay time.Duration = time.Millisecond * 650

var panelBackgroundColor = sdl.Color{R: 115, G: 115, B: 115, A: 150}
var panelActiveToolColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var panelHoverToolColor = sdl.Color{R: 100, G: 100, B: 100, A: 200}
var panelSeparatorColor = sdl.Color{R: 170, G: 170, B: 170, A: 255}
var panelToolColorOutlineColor = sdl.Color{R: 0, G: 0, B: 0, A: 40}

type ToolsPanel struct {
	tools             []*toolMeta
	currentTool       *toolMeta
	hoveredTool       *toolMeta
	hoveredAt         time.Time
	cropTool          editTools.ScreenshotCropTool
	actionsQueue      *editTools.ActionsQueue
	onNewToolSelected func(tool editTools.ScreenshotEditTool)
	handCursorSet     bool
	panelRect         *sdl.Rect
}

func NewToolsPanel(
	ren *sdl.Renderer,
	onNewToolSelected func(tool editTools.ScreenshotEditTool),
	saveCallback, copyCallback, searchCallback func(),
) *ToolsPanel {
	selectionTool := editTools.NewSelectionTool(ren, saveCallback, copyCallback, searchCallback)
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
		tools:             metas,
		actionsQueue:      editTools.NewActionsQueue(),
		cropTool:          selectionTool,
		onNewToolSelected: onNewToolSelected,
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
	pkg.DrawRoundedFilledRectangle(ren, panel.panelRect, panelRoundingRadius, panelBackgroundColor)
	for i, meta := range panel.tools {
		panel.drawTool(ren, meta)
		if i != len(panel.tools)-1 {
			pkg.DrawThickLine(
				ren,
				&sdl.Point{X: meta.toolBBox.X + meta.toolBBox.W + (panelIconMargin / 2), Y: meta.toolBBox.Y + panelIconPadding},
				&sdl.Point{X: meta.toolBBox.X + meta.toolBBox.W + (panelIconMargin / 2), Y: meta.toolBBox.Y + meta.toolBBox.H - panelIconPadding},
				panelSeparatorWidth, panelSeparatorColor,
			)
		}
	}
}

func (panel *ToolsPanel) drawTool(ren *sdl.Renderer, meta *toolMeta) {
	meta.texture.SetColorMod(255, 255, 255)
	if meta == panel.hoveredTool {
		pkg.DrawRoundedFilledRectangle(ren, &meta.toolBBox, panelRoundingRadius, panelHoverToolColor)
		toolSettings := panel.hoveredTool.tool.ToolSettings()
		if len(toolSettings) > 0 && time.Since(panel.hoveredAt) >= panelSettingsShowDelay {
			for _, setting := range toolSettings {
				setting.Render(ren)
			}
		}
	}
	if meta == panel.currentTool {
		pkg.DrawRoundedFilledRectangle(ren, &meta.toolBBox, panelRoundingRadius, panelActiveToolColor)
		meta.texture.SetColorMod(0, 0, 0)
	}
	if toolColor := meta.tool.ToolColor(); toolColor != nil {
		pkg.DrawFilledRectangle(
			ren,
			meta.colorBBox,
			*toolColor,
		)
		pkg.DrawThickRectangle(
			ren,
			&sdl.Rect{
				X: meta.colorBBox.X - 1, Y: meta.colorBBox.Y - 1,
				W: meta.colorBBox.W + 2, H: meta.colorBBox.H + 2,
			},
			panelToolColorOutlineWidth,
			panelToolColorOutlineColor,
		)
	}
	pkg.CopyTexture(ren, meta.texture, &meta.iconBBox, nil)
}

func (panel *ToolsPanel) SetToolsCallbacks(callbacks *gui.WindowCallbackSet) {
	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		click := sdl.Point{X: x, Y: y}
		for _, meta := range panel.tools {
			if click.InRect(&meta.toolBBox) && button == sdl.BUTTON_LEFT {
				panel.setActiveTool(meta)
				return true
			}
			if panel.hoveredTool == meta && time.Since(panel.hoveredAt) >= panelSettingsShowDelay {
				if click.InRect(&panel.hoveredTool.settingsBBox) && button == sdl.BUTTON_LEFT {
					panel.setActiveTool(meta)
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
					panel.hoveredAt = time.Now()
				}
				panel.hoveredTool = meta
				sdl.SetCursor(gui.HandCursor)
				panel.handCursorSet = true
				return false
			}
		}
		if panel.hoveredTool != nil {
			if panel.handCursorSet {
				sdl.SetCursor(gui.ArrowCursor)
			}
			if move.InRect(&panel.hoveredTool.settingsBBox) && time.Since(panel.hoveredAt) >= panelSettingsShowDelay {
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

	if panel.hoveredTool != nil && time.Since(panel.hoveredAt) >= panelSettingsShowDelay {
		for _, setting := range panel.hoveredTool.tool.ToolSettings() {
			callbacks.Append(setting.SettingCallbacks())
		}
	}
	if panel.currentTool != nil {
		callbacks.Append(panel.currentTool.tool.ToolCallbacks(panel.actionsQueue))
	}
}

func (panel *ToolsPanel) setActiveTool(toolMeta *toolMeta) {
	if panel.currentTool != nil {
		panel.currentTool.tool.OnToolDeactivated()
	}
	panel.currentTool = toolMeta
	panel.currentTool.tool.OnToolActivated()
	panel.onNewToolSelected(toolMeta.tool)
}

func (panel *ToolsPanel) resizePanel(viewportW, viewportH int32) {
	panelWidth := (panelToolSize * int32(len(panel.tools))) + (panelIconMargin * int32(len(panel.tools)-1)) + (panelPadding * 2)
	panelRect := sdl.Rect{
		X: (viewportW - panelWidth) / 2, Y: panelTopMargin,
		W: panelWidth, H: panelToolSize + (panelPadding * 2),
	}
	panel.panelRect = &panelRect
	for i, meta := range panel.tools {
		meta.iconBBox = sdl.Rect{
			X: panelRect.X + panelPadding + panelIconPadding + ((panelIconMargin + panelToolSize) * int32(i)),
			Y: panelRect.Y + panelPadding + panelIconPadding,
			W: panelIconSize,
			H: panelIconSize,
		}
		meta.toolBBox = sdl.Rect{
			X: meta.iconBBox.X - panelIconPadding,
			Y: meta.iconBBox.Y - panelIconPadding,
			W: panelToolSize, H: panelToolSize,
		}
		meta.colorBBox = &sdl.Rect{
			X: meta.toolBBox.X + (meta.toolBBox.W-panelToolColorWidth)/2,
			Y: meta.toolBBox.Y + meta.toolBBox.H - panelToolColorHeight - panelToolColorPadding,
			W: panelToolColorWidth, H: panelToolColorHeight,
		}
		var toolSettingsH int32 = 0
		for _, setting := range meta.tool.ToolSettings() {
			setting.SetLeftTop(&sdl.Point{
				X: meta.toolBBox.X - (panelSettingsWidth-meta.toolBBox.W)/2,
				Y: meta.toolBBox.Y + meta.toolBBox.H + panelPadding + toolSettingsH,
			})
			setting.SetWidth(panelSettingsWidth)
			toolSettingsH += setting.BBox().H
		}
		meta.settingsBBox = sdl.Rect{
			X: meta.toolBBox.X - (panelSettingsWidth-meta.toolBBox.W)/2,
			Y: meta.toolBBox.Y + meta.toolBBox.H,
			W: panelSettingsWidth, H: toolSettingsH,
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
	return toolMeta{
		tool:     tool,
		iconBBox: sdl.Rect{},
		texture:  pkg.CreateTextureFromSurface(ren, tool.ToolIcon()),
	}
}
