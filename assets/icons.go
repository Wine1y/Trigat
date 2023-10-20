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

//go:embed icons/copy_action.png
var copyIconData []byte
var CopyIcon = pkg.LoadPNGSurface(copyIconData)

//go:embed icons/save_action.png
var saveIconData []byte
var SaveIcon = pkg.LoadPNGSurface(saveIconData)

//go:embed icons/search_action.png
var searchIconData []byte
var SearchIcon = pkg.LoadPNGSurface(searchIconData)

//go:embed icons/trigat_icon.ico
var TrigatIconData []byte
