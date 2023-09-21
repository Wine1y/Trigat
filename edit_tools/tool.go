package editTools

import (
	"github.com/Wine1y/trigat/gui"
	"github.com/veandco/go-sdl2/sdl"
)

type ScreenshotEditTool interface {
	ToolCallbacks(*ActionsQueue) *gui.WindowCallbackSet
	RenderCurrentState(ren *sdl.Renderer)
	RenderScreenshot(ren *sdl.Renderer)
	ToolIcon() *sdl.Surface
}

type ScreenshotCropTool interface {
	ScreenshotEditTool
	CropScreenshot(surface *sdl.Surface) *sdl.Surface
}
