package editTools

import (
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/internal/gui/sc_window/settings"
	"github.com/veandco/go-sdl2/sdl"
)

type ScreenshotEditTool interface {
	ToolCallbacks(*ActionsQueue) *gui.WindowCallbackSet
	RenderCurrentState(ren *sdl.Renderer)
	RenderScreenshot(ren *sdl.Renderer)
	ToolIcon() *sdl.Surface
	ToolSettings() []settings.ToolSetting
	ToolColor() *sdl.Color
	OnToolActivated()
	OnToolDeactivated()
}

type ScreenshotCropTool interface {
	ScreenshotEditTool
	CropScreenshot(surface *sdl.Surface) *sdl.Surface
}

type DefaultScreenshotEditTool struct {
}

func (tool DefaultScreenshotEditTool) OnToolActivated() {

}

func (tool DefaultScreenshotEditTool) OnToolDeactivated() {

}

func (tool DefaultScreenshotEditTool) ToolSettings() []settings.ToolSetting {
	return nil
}

func (tool DefaultScreenshotEditTool) ToolColor() *sdl.Color {
	return nil
}
