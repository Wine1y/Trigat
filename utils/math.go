package utils

import "github.com/veandco/go-sdl2/sdl"

func Abs[I int | int8 | int16 | int32 | int64](integer I) I {
	if integer > 0 {
		return integer
	}
	return integer * -1
}

func RectIntoSquare(rect *sdl.Rect) {
	squareSize := (Abs(rect.W) + Abs(rect.H)) / 2
	if rect.W > 0 {
		rect.W = squareSize
	} else {
		rect.W = -squareSize
	}
	if rect.H > 0 {
		rect.H = squareSize
	} else {
		rect.H = -squareSize
	}
}
