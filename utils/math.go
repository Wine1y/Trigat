package utils

import "github.com/veandco/go-sdl2/sdl"

type Integer interface {
	int | int8 | int16 | int32 | int64
}

type Uinteger interface {
	uint | uint8 | uint16 | uint32 | uint64
}
type Number interface {
	Integer | Uinteger
}

func Abs[I Integer](integer I) I {
	if integer > 0 {
		return integer
	}
	return integer * -1
}

func Max[N Number](first N, others ...N) N {
	var max N = first
	for _, number := range others {
		if number > max {
			max = number
		}
	}
	return max
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
