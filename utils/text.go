package utils

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func LoadFont(fontData []byte, size int) *ttf.Font {
	fontRW, err := sdl.RWFromMem(fontData)
	if err != nil {
		panic(err)
	}
	font, err := ttf.OpenFontRW(fontRW, 1, size)
	if err != nil {
		panic(err)
	}
	return font
}

func SizeString(font *ttf.Font, text string) (int, int) {
	w, h, err := font.SizeUTF8(text)
	if err != nil {
		panic(err)
	}
	return w, h
}

type StringTexture struct {
	Texture    *sdl.Texture
	TextWidth  int32
	TextHeight int32
}

func NewStringTexture(ren *sdl.Renderer, font *ttf.Font, text string, color sdl.Color) *StringTexture {
	surface, err := font.RenderUTF8BlendedWrapped(text, color, 0)
	if err != nil {
		panic(err)
	}
	defer surface.Free()
	texture, err := ren.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}
	return &StringTexture{
		Texture:    texture,
		TextWidth:  surface.W,
		TextHeight: surface.H,
	}
}

func (text *StringTexture) Draw(ren *sdl.Renderer, leftTop *sdl.Point) {
	CopyTexture(
		ren,
		text.Texture,
		&sdl.Rect{X: leftTop.X, Y: leftTop.Y, W: text.TextWidth, H: text.TextHeight},
		nil,
	)
}

func (text *StringTexture) Destroy() {
	text.Texture.Destroy()
}
