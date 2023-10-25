package pkg

import (
	"fmt"
	"math"
	"unicode"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type TextParagraph struct {
	Text          []rune
	StringTexture *StringTexture
	TextStart     sdl.Point
	Color         sdl.Color
	Font          *ttf.Font
	padding       int32
}

func NewTextParagraph(
	textStart sdl.Point,
	textColor sdl.Color,
	textFont *ttf.Font,
	paragraphPadding int32,
) *TextParagraph {
	return &TextParagraph{
		Text:      make([]rune, 0, 1),
		TextStart: textStart,
		Color:     textColor,
		Font:      textFont,
		padding:   paragraphPadding,
	}
}

func (par TextParagraph) GetBBox() *sdl.Rect {
	w, h := int32(par.Font.Height()), int32(par.Font.Height())
	if par.StringTexture != nil {
		h = par.StringTexture.TextHeight
		lines := par.GetLinesBoundaries()
		for _, line := range lines {
			lineW, _ := SizeString(par.Font, string(par.Text[line[0]:line[1]]))
			if int32(lineW) > w {
				w = int32(lineW)
			}
		}
	}
	if len(par.Text) > 0 && par.Text[len(par.Text)-1] == '\n' {
		h += int32(par.Font.Height())
	}
	return &sdl.Rect{
		X: par.TextStart.X - par.padding, Y: par.TextStart.Y - par.padding,
		W: w + par.padding*2, H: h + par.padding*2,
	}
}

func (par TextParagraph) GetPaddedBBox(padding int32) *sdl.Rect {
	bbox := par.GetBBox()
	return &sdl.Rect{
		X: bbox.X - padding, Y: bbox.Y - padding,
		W: bbox.W + padding*2, H: bbox.H + padding*2,
	}
}

func (par *TextParagraph) SetColor(color sdl.Color) {
	par.Color = color
}

func (par *TextParagraph) GetLinesBoundaries() [][2]int {
	lines := make([][2]int, 0, 1)
	i, lineStart := 0, 0
	for i < len(par.Text) {
		if par.Text[i] == '\n' || i == len(par.Text)-1 {
			lines = append(lines, [2]int{lineStart, i})
			lineStart = i + 1
		}
		i++
	}
	if len(lines) > 0 {
		lines[len(lines)-1][1]++
	}
	return lines
}

func (par TextParagraph) GetLineNumber(position int) int {
	lines := par.GetLinesBoundaries()
	for i, lineBoundaries := range lines {
		if position >= lineBoundaries[0] && position <= lineBoundaries[1] {
			return i
		}
	}
	panic(fmt.Sprintf("Invalid position %v, can't determine line number for lines: %v.", position, lines))
}

func (par *TextParagraph) InsertRunes(ren *sdl.Renderer, pos int, runes ...rune) {
	newText := make([]rune, 0, len(par.Text)+1)
	newText = append(newText, par.Text[:pos]...)
	newText = append(newText, runes...)
	if len(par.Text) > pos {
		newText = append(newText, par.Text[pos:]...)
	}
	par.Text = newText
	par.updateTexture(ren)
}

func (par *TextParagraph) PopRunes(ren *sdl.Renderer, startPos, endPos int) {
	startPos = Clamp(0, startPos, len(par.Text)-1)
	endPos = Clamp(startPos+1, endPos, len(par.Text))
	newText := make([]rune, 0, len(par.Text)-(endPos-startPos))
	newText = append(newText, par.Text[0:startPos]...)
	if endPos < len(par.Text) {
		newText = append(newText, par.Text[endPos:len(par.Text)]...)
	}
	par.Text = newText
	par.updateTexture(ren)
}

func (par TextParagraph) GetOffsetByPosition(position int) (int32, int32) {
	lines := par.GetLinesBoundaries()
	if len(lines) == 0 {
		return int32(0), int32(0)
	}
	lineNumber := par.GetLineNumber(position)
	if position == len(par.Text) && par.Text[position-1] == '\n' {
		return int32(0), int32((lineNumber + 1) * par.Font.LineSkip())
	}
	x, _ := SizeString(par.Font, string(par.Text[lines[lineNumber][0]:position]))
	return int32(x), int32(lineNumber * par.Font.LineSkip())
}

func (par TextParagraph) GetPositionByOffset(xOffset int32, yOffset int32) int {
	lineNumber := int(math.Floor(float64(yOffset) / float64(par.Font.LineSkip())))
	lineNumber = Max(0, lineNumber)
	lines := par.GetLinesBoundaries()
	if lineNumber >= len(lines) {
		return len(par.Text)
	}
	lineStart, lineEnd := lines[lineNumber][0], lines[lineNumber][1]
	w, _ := SizeString(par.Font, string(par.Text[lineStart:lineEnd]))
	if int(xOffset) >= w {
		return lineEnd + 1
	}
	start, end := lineStart, lineEnd+1
	for {
		pos := (start + end) / 2
		if start == pos || end == pos {
			return pos
		}
		w, _ := SizeString(par.Font, string(par.Text[lineStart:pos]))
		if w > int(xOffset) {
			end = pos
		} else {
			start = pos
		}
	}
}

func (par TextParagraph) ClosestLeftWordPos(pos int) int {
	inWord := false
PreWordLoop:
	for pos > 0 {
		numberOrLetter := unicode.In(par.Text[pos-1], unicode.Letter, unicode.Number)
		switch {
		case inWord && !numberOrLetter:
			break PreWordLoop
		case numberOrLetter:
			inWord = true
		}
		pos--
	}
	return pos
}

func (par TextParagraph) ClosestRightWordPos(pos int) int {
	inWord := false
	outWord := false
PostWordLoop:
	for pos < len(par.Text) {
		numberOrLetter := unicode.In(par.Text[pos], unicode.Letter, unicode.Number)
		switch {
		case outWord && numberOrLetter:
			break PostWordLoop
		case inWord && !numberOrLetter:
			outWord = true
		case numberOrLetter:
			inWord = true
		}
		pos++
	}
	return pos
}

func (par TextParagraph) UpperLinePos(position int) int {
	lineNumber := par.GetLineNumber(position)
	lines := par.GetLinesBoundaries()
	if lineNumber == 0 || len(lines) == 0 {
		return 0
	}
	upperLine := lines[lineNumber-1]
	lineOffset := position - lines[lineNumber][0]
	upperLinePos := upperLine[0] + lineOffset
	return Min(upperLinePos, upperLine[1])
}

func (par TextParagraph) LowerLinePos(position int) int {
	lineNumber := par.GetLineNumber(position)
	lines := par.GetLinesBoundaries()
	if lineNumber == len(lines)-1 || len(lines) == 0 {
		return len(par.Text)
	}
	lowerLine := lines[lineNumber+1]
	lineOffset := position - lines[lineNumber][0]
	lowerLinePos := lowerLine[0] + lineOffset
	lowerLinePos = Min(lowerLinePos, lowerLine[1])
	return lowerLinePos
}

func (par *TextParagraph) updateTexture(ren *sdl.Renderer) {
	if par.StringTexture != nil {
		par.StringTexture.Destroy()
	}
	if len(par.Text) == 0 {
		par.StringTexture = nil
		return
	}
	par.StringTexture = NewStringTexture(ren, par.Font, string(par.Text), par.Color)
}
