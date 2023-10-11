package settings

import (
	"github.com/Wine1y/trigat/internal/gui"
	"github.com/veandco/go-sdl2/sdl"
)

type ToolSetting interface {
	BBox() *sdl.Rect
	SetLeftTop(lt *sdl.Point)
	SetWidth(width int32)

	Render(ren *sdl.Renderer)
	SettingCallbacks() *gui.WindowCallbackSet
}

type DefaultSetting struct {
	bbox sdl.Rect
}

func NewDefaultSetting(settingHeight int32) *DefaultSetting {
	return &DefaultSetting{
		bbox: sdl.Rect{X: 0, Y: 0, W: 1, H: settingHeight},
	}
}

func (setting *DefaultSetting) SetLeftTop(lt *sdl.Point) {
	setting.bbox.X = lt.X
	setting.bbox.Y = lt.Y
}

func (setting *DefaultSetting) SetWidth(width int32) {
	setting.bbox.W = width
}

func (setting DefaultSetting) BBox() *sdl.Rect {
	return &setting.bbox
}
