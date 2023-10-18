package pkg

import (
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/chai2010/webp"
	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/image/bmp"
)

func WriteSurfaceToPNG(surface *sdl.Surface, writer io.Writer) {
	if err := png.Encode(writer, surface); err != nil {
		panic(err)
	}
}

func WriteSurfaceToJPEG(surface *sdl.Surface, writer io.Writer) {
	if err := jpeg.Encode(writer, surface, &jpeg.Options{Quality: 100}); err != nil {
		panic(err)
	}
}

func WriteSurfaceToGIF(surface *sdl.Surface, writer io.Writer) {
	if err := gif.Encode(writer, surface, &gif.Options{NumColors: 256}); err != nil {
		panic(err)
	}
}

func WriteSurfaceToWEBP(surface *sdl.Surface, writer io.Writer) {
	if err := webp.Encode(writer, surface, &webp.Options{Lossless: true, Quality: 100}); err != nil {
		panic(err)
	}
}

func WriteSurfaceToBMP(surface *sdl.Surface, writer io.Writer) {
	if err := bmp.Encode(writer, surface); err != nil {
		panic(err)
	}
}
