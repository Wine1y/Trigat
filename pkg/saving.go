package pkg

import (
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"

	"github.com/chai2010/webp"
	"github.com/sqweek/dialog"
	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/image/bmp"
)

var SavingMethods []SavingMethod = []SavingMethod{
	{Name: "PNG", AllowedExtensions: []string{".png"}, WritingFunction: WriteSurfaceToPNG},
	{Name: "JPEG", AllowedExtensions: []string{".jpeg", ".jpg"}, WritingFunction: WriteSurfaceToJPEG},
	{Name: "GIF", AllowedExtensions: []string{".gif"}, WritingFunction: WriteSurfaceToGIF},
	{Name: "WEBP", AllowedExtensions: []string{".webp"}, WritingFunction: WriteSurfaceToWEBP},
	{Name: "BMP", AllowedExtensions: []string{".bmp"}, WritingFunction: WriteSurfaceToBMP},
}

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

func RequestSavingOptions(
	dialogTitle,
	dialogStartFileName string,
) (
	options *SavingOptions,
	success bool,
) {
	dialogBuilder := dialog.File()
	dialogBuilder.Title(dialogTitle)
	dialogBuilder.SetStartFile(fmt.Sprintf("%s%s", dialogStartFileName, SavingMethods[0].AllowedExtensions[0]))
	for _, method := range SavingMethods {
		dialogBuilder.Filter(method.Name, method.AllowedExtensions...)
	}
	path, err := dialogBuilder.Save()
	if err != nil {
		return nil, false
	}
	ext := filepath.Ext(path)
	for _, method := range SavingMethods {
		for _, allowedExt := range method.AllowedExtensions {
			if allowedExt == ext {
				return &SavingOptions{
					Filepath: path,
					Method:   method,
				}, true
			}
		}
	}
	return nil, false
}

type SavingMethod struct {
	Name              string
	AllowedExtensions []string
	WritingFunction   func(surface *sdl.Surface, writer io.Writer)
}

type SavingOptions struct {
	Filepath string
	Method   SavingMethod
}
