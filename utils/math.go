package utils

import "github.com/veandco/go-sdl2/sdl"

type SInteger interface {
	int | int8 | int16 | int32 | int64
}

type Uinteger interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type Float interface {
	float32 | float64
}

type Integer interface {
	SInteger | Uinteger
}

type Number interface {
	Integer | Float
}

func Abs[I SInteger | Float](integer I) I {
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

func Min[N Number](first N, others ...N) N {
	var min N = first
	for _, number := range others {
		if number < min {
			min = number
		}
	}
	return min
}

func Clamp[N Number](min N, number N, max N) N {
	if number > max {
		number = max
	}
	if number < min {
		number = min
	}
	return number
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

func Linspace(start, stop float64, num int) []float64 {
	num_f := float64(num)
	h := (stop - start) / (num_f - 1)
	res := make([]float64, num)
	for i := 0; i < len(res); i++ {
		res[i] = start + h*float64(i)
	}
	return res
}
