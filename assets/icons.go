package assets

import (
	_ "embed"

	"github.com/Wine1y/trigat/pkg"
)

//go:embed icons/line_tool.png
var lineIconData []byte
var LineIcon = pkg.LoadPNGSurface(lineIconData)

//go:embed icons/paint_tool.png
var paintIconData []byte
var PaintIcon = pkg.LoadPNGSurface(paintIconData)

//go:embed icons/pipette_tool.png
var pipetteIconData []byte
var PipetteIcon = pkg.LoadPNGSurface(pipetteIconData)

//go:embed icons/rect_tool.png
var rectIconData []byte
var RectIcon = pkg.LoadPNGSurface(rectIconData)

//go:embed icons/selection_tool.png
var selectionIconData []byte
var SelectionIcon = pkg.LoadPNGSurface(selectionIconData)

//go:embed icons/text_tool.png
var textIconData []byte
var TextIcon = pkg.LoadPNGSurface(textIconData)
