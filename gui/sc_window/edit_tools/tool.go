package editTools

import (
	"github.com/Wine1y/trigat/gui"
	"github.com/Wine1y/trigat/gui/sc_window/settings"
	"github.com/veandco/go-sdl2/sdl"
)

type ScreenshotEditTool interface {
	ToolCallbacks(*ActionsQueue) *gui.WindowCallbackSet
	RenderCurrentState(ren *sdl.Renderer)
	RenderScreenshot(ren *sdl.Renderer)
	ToolIcon() *sdl.Surface
	ToolSettings() []settings.ToolSetting
	ToolColor() *sdl.Color
}

type ScreenshotCropTool interface {
	ScreenshotEditTool
	CropScreenshot(surface *sdl.Surface) *sdl.Surface
}
