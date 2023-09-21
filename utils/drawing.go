package utils

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

func LoadPNGSurface(pngData []byte) *sdl.Surface {
	rwOps, err := sdl.RWFromMem(pngData)
	if err != nil {
		panic(err)
	}
	defer rwOps.Close()
	surface, err := img.LoadPNGRW(rwOps)
	if err != nil {
		panic(err)
	}
	return surface
}

func DrawRoundedFilledRectangle(ren *sdl.Renderer, rect *sdl.Rect, radius int32, color sdl.Color) {
	gfx.RoundedBoxColor(
		ren,
		rect.X, rect.Y,
		rect.X+rect.W, rect.Y+rect.H,
		radius, color,
	)
}

func DrawThickLine(ren *sdl.Renderer, p1 *sdl.Point, p2 *sdl.Point, width int32, color sdl.Color) {
	gfx.ThickLineColor(ren, p1.X, p1.Y, p2.X, p2.Y, width, color)
}

func DrawFilledRectangle(ren *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	gfx.BoxColor(ren, rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H, color)
}

func DrawThickRectangle(ren *sdl.Renderer, rect *sdl.Rect, width int32, color sdl.Color) {
	lt, rt := &sdl.Point{X: rect.X, Y: rect.Y}, &sdl.Point{X: rect.X + rect.W, Y: rect.Y}
	lb, rb := &sdl.Point{X: rect.X, Y: rect.Y + rect.H}, &sdl.Point{X: rect.X + rect.W, Y: rect.Y + rect.H}
	DrawThickLine(ren, lt, lb, width, color)
	DrawThickLine(ren, rt, rb, width, color)
	DrawThickLine(ren, lt, rt, width, color)
	DrawThickLine(ren, lb, rb, width, color)
}

func DrawFilledCircle(ren *sdl.Renderer, center *sdl.Point, radius int32, color sdl.Color) {
	gfx.FilledCircleColor(ren, center.X, center.Y, radius, color)
}

func CopyTexture(ren *sdl.Renderer, texture *sdl.Texture, dst *sdl.Rect, blendMode *sdl.BlendMode) {
	if blendMode != nil {
		texture.SetBlendMode(*blendMode)
	}
	ren.Copy(texture, nil, dst)
}
