package editTools

import (
	_ "embed"
	"unicode"

	"github.com/Wine1y/trigat/gui"
	"github.com/Wine1y/trigat/gui/sc_window/settings"
	"github.com/Wine1y/trigat/utils"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var textColor = sdl.Color{R: 255, G: 0, B: 0, A: 255}

//go:embed icons/text_tool.png
var textIconData []byte
var textRgbIcon = utils.LoadPNGSurface(textIconData)

//go:embed font.ttf
var defaultFontData []byte

type TextTool struct {
	paragraphs []*paragraph
	ren        *sdl.Renderer
	font       *ttf.Font
}

func NewTextTool(renderer *sdl.Renderer) *TextTool {
	return &TextTool{
		paragraphs: make([]*paragraph, 0),
		ren:        renderer,
		font:       utils.LoadFont(defaultFontData, 14),
	}
}

func (tool TextTool) ToolIcon() *sdl.Surface {
	return textRgbIcon
}

func (tool *TextTool) ToolCallbacks(queue *ActionsQueue) *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		newParagraph := paragraph{text: make([]rune, 0, 10), textStart: sdl.Point{X: x, Y: y}}
		tool.paragraphs = append(tool.paragraphs, &newParagraph)
		queue.Push(TextAction{tool: tool, lastParagraph: &newParagraph})
		return false
	})

	callbacks.TextInput = append(callbacks.TextInput, func(rn rune) bool {
		if !unicode.IsGraphic(rn) || len(tool.paragraphs) <= 0 {
			return false
		}
		lastParagraph := tool.paragraphs[len(tool.paragraphs)-1]
		lastParagraph.text = append(lastParagraph.text, rn)
		lastParagraph.updateTexture(tool.ren, tool.font)
		return false
	})

	callbacks.KeyDown = append(callbacks.KeyDown, func(keysym sdl.Keysym) bool {
		if keysym.Sym == sdl.K_BACKSPACE && len(tool.paragraphs) > 0 {
			lastParagraph := tool.paragraphs[len(tool.paragraphs)-1]
			if len(lastParagraph.text) > 0 {
				lastParagraph.text = lastParagraph.text[:len(lastParagraph.text)-1]
				lastParagraph.updateTexture(tool.ren, tool.font)
			}
		}
		return false
	})

	callbacks.Quit = append(callbacks.Quit, func() bool {
		for _, par := range tool.paragraphs {
			if par.stringTexture != nil {
				par.stringTexture.Destroy()
			}
		}
		tool.font.Close()
		return false
	})

	return callbacks
}

func (tool TextTool) RenderCurrentState(ren *sdl.Renderer) {
	for _, par := range tool.paragraphs {
		if par.stringTexture != nil {
			par.stringTexture.Draw(ren, &par.textStart)
		}
	}
}

func (tool TextTool) RenderScreenshot(ren *sdl.Renderer) {
	if ren == tool.ren {
		tool.RenderCurrentState(ren)
		return
	}
	for _, par := range tool.paragraphs {
		text := utils.NewStringTexture(ren, tool.font, string(par.text), textColor)
		text.Draw(ren, &par.textStart)
	}
}

func (tool TextTool) ToolSettings() []settings.ToolSetting {
	return nil
}

type paragraph struct {
	text          []rune
	stringTexture *utils.StringTexture
	textStart     sdl.Point
}

func (par *paragraph) updateTexture(ren *sdl.Renderer, font *ttf.Font) {
	if par.stringTexture != nil {
		par.stringTexture.Destroy()
	}
	if len(par.text) == 0 {
		par.stringTexture = nil
		return
	}
	par.stringTexture = utils.NewStringTexture(ren, font, string(par.text), textColor)
}

type TextAction struct {
	tool          *TextTool
	lastParagraph *paragraph
}

func (action TextAction) Undo() {
	action.tool.paragraphs = action.tool.paragraphs[:len(action.tool.paragraphs)-1]
}

func (action TextAction) Redo() {
	action.tool.paragraphs = append(action.tool.paragraphs, action.lastParagraph)
}
