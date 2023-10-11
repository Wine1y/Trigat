package pkg

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

func DrawPoint(ren *sdl.Renderer, point *sdl.Point, color sdl.Color) {
	ren.SetDrawColor(color.R, color.G, color.B, color.A)
	ren.DrawPoint(point.X, point.Y)
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

func DrawRectangle(ren *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	gfx.RectangleColor(ren, rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H, color)
}

func DrawFilledRectangle(ren *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	gfx.BoxColor(ren, rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H, color)
}

func DrawThickRectangle(ren *sdl.Renderer, rect *sdl.Rect, width int32, color sdl.Color) {
	for r := int32(0); r < width; r++ {
		DrawRectangle(ren, &sdl.Rect{X: rect.X - r/2, Y: rect.Y - r/2, W: rect.W + r, H: rect.H + r}, color)
	}
}

func DrawFilledCircle(ren *sdl.Renderer, center *sdl.Point, radius int32, color sdl.Color) {
	gfx.FilledCircleColor(ren, center.X, center.Y, radius, color)
}

func DrawCircle(ren *sdl.Renderer, center *sdl.Point, radius int32, color sdl.Color) {
	gfx.AACircleColor(ren, center.X, center.Y, radius, color)
}

func DrawThickCircle(ren *sdl.Renderer, center *sdl.Point, radius int32, width int32, color sdl.Color) {
	for r := radius; r >= radius-width; r-- {
		DrawCircle(ren, center, r, color)
	}
}

func CopyTexture(ren *sdl.Renderer, texture *sdl.Texture, dst *sdl.Rect, blendMode *sdl.BlendMode) {
	if blendMode != nil {
		texture.SetBlendMode(*blendMode)
	}
	ren.Copy(texture, nil, dst)
}
