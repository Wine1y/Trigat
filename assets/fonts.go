package assets

import (
	_ "embed"

	"github.com/Wine1y/trigat/pkg"
	"github.com/veandco/go-sdl2/ttf"
)

//go:embed fonts/defaultAppFont.ttf
var defaultFontData []byte

func GetAppFont(size int) *ttf.Font {
	return pkg.LoadFont(defaultFontData, size)
}
