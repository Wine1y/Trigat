package settings

import (
	"math"

	"github.com/Wine1y/trigat/gui"
	"github.com/Wine1y/trigat/utils"
	"github.com/veandco/go-sdl2/sdl"
)

const colorPickerHeight int32 = 200
const gradientPadding int32 = 10
const pickerGradientHeight int32 = 80
const hueGradientHeight int32 = 20
const pickerThumbRadius int32 = 8
const pickerThumbThickness int32 = 3

var hueThumbColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var pickerThumbColorLight = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var pickerThumbColorDark = sdl.Color{R: 0, G: 0, B: 0, A: 255}

type ColorPickerSetting struct {
	*DefaultSetting
	currentColor          hslColor
	currentPickerGradient cachedGradient
	currentHueGradient    cachedGradient
	lastRenderer          *sdl.Renderer
	draggingHue           bool
	draggingPicker        bool
	onColorUpdated        func(color sdl.Color)
}

func NewColorPickerSetting(onColorUpdated func(color sdl.Color)) *ColorPickerSetting {
	return &ColorPickerSetting{
		DefaultSetting: NewDefaultSetting(colorPickerHeight),
		currentColor:   hslColor{H: 0, S: 1, L: 0.5},
		onColorUpdated: onColorUpdated,
	}
}

func (setting *ColorPickerSetting) Render(ren *sdl.Renderer) {
	if ren != setting.lastRenderer {
		setting.lastRenderer = ren
		setting.updateGradients()
	}
	utils.CopyTexture(ren, setting.currentPickerGradient.texture, setting.currentPickerGradient.bbox, nil)
	utils.CopyTexture(ren, setting.currentHueGradient.texture, setting.currentHueGradient.bbox, nil)

	utils.DrawThickLine(
		ren,
		&sdl.Point{
			X: setting.currentHueGradient.bbox.X + setting.currentColor.hueOffset(setting.currentHueGradient.bbox.W),
			Y: setting.currentHueGradient.bbox.Y - gradientPadding/2,
		},
		&sdl.Point{
			X: setting.currentHueGradient.bbox.X + setting.currentColor.hueOffset(setting.currentHueGradient.bbox.W),
			Y: setting.currentHueGradient.bbox.Y + setting.currentHueGradient.bbox.H - 1 + gradientPadding/2,
		},
		1, hueThumbColor,
	)

	var pickerThumbColor sdl.Color = pickerThumbColorLight
	if setting.currentColor.L > 0.7 || (setting.currentColor.L > 0.3 && setting.currentColor.S < 0.35) {
		pickerThumbColor = pickerThumbColorDark
	}

	utils.DrawThickCircle(
		ren,
		&sdl.Point{
			X: setting.currentPickerGradient.bbox.X + setting.currentColor.xOffset(setting.currentPickerGradient.bbox.W),
			Y: setting.currentPickerGradient.bbox.Y + setting.currentColor.yOffset(setting.currentPickerGradient.bbox.H),
		},
		pickerThumbRadius,
		pickerThumbThickness,
		pickerThumbColor,
	)
	utils.DrawThickRectangle(ren, &setting.bbox, 1, sdl.Color{R: 0, G: 255, B: 0, A: 255})
}

func (setting *ColorPickerSetting) SettingCallbacks() *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()

	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		if button != sdl.BUTTON_LEFT {
			return false
		}
		click := sdl.Point{X: x, Y: y}
		if click.InRect(setting.currentHueGradient.bbox) {
			setting.newHueValue(x - setting.currentHueGradient.bbox.X)
			setting.draggingHue = true
		}
		if click.InRect(setting.currentPickerGradient.bbox) {
			setting.newPickerValue(
				x-setting.currentPickerGradient.bbox.X,
				y-setting.currentPickerGradient.bbox.Y,
			)
			setting.draggingPicker = true
		}
		if click.InRect(&setting.bbox) {
			return true
		}
		return false
	})

	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		move := sdl.Point{X: x, Y: y}
		if !move.InRect(&setting.bbox) {
			setting.draggingHue = false
			setting.draggingPicker = false
		}
		if setting.draggingHue {
			setting.newHueValue(x - setting.currentHueGradient.bbox.X)
		}
		if setting.draggingPicker {
			setting.newPickerValue(
				x-setting.currentPickerGradient.bbox.X,
				y-setting.currentPickerGradient.bbox.Y,
			)
		}
		return false
	})

	callbacks.MouseUp = append(callbacks.MouseUp, func(button uint8, x, y int32) bool {
		if button == sdl.BUTTON_LEFT {
			setting.draggingHue = false
			setting.draggingPicker = false
		}
		return false
	})

	callbacks.SizeChange = append(callbacks.SizeChange, func(w, h int32) bool {
		setting.lastRenderer = nil
		return false
	})

	callbacks.Quit = append(callbacks.Quit, func() bool {
		if setting.currentHueGradient.texture != nil {
			setting.currentHueGradient.texture.Destroy()
		}
		if setting.currentPickerGradient.texture != nil {
			setting.currentPickerGradient.texture.Destroy()
		}
		return false
	})
	return callbacks
}

func (setting *ColorPickerSetting) SetWidth(width int32) {
	setting.DefaultSetting.SetWidth(width)
	setting.lastRenderer = nil
}

func (setting *ColorPickerSetting) newHueValue(hueOffset int32) {
	hueOffset = utils.Clamp(0, hueOffset, setting.currentHueGradient.bbox.W-1)
	h := 360 / (float64(setting.currentHueGradient.bbox.W) - 1)
	setting.currentColor.H = h * float64(hueOffset)
	setting.updatePickerGradient()
	setting.colorUpdated()
}

func (setting *ColorPickerSetting) newPickerValue(xOffset, yOffset int32) {
	xOffset = utils.Clamp(0, xOffset, setting.currentPickerGradient.bbox.W-1)
	yOffset = utils.Clamp(0, yOffset, setting.currentPickerGradient.bbox.H-1)
	h := 1 / (float64(setting.currentPickerGradient.bbox.W) - 1)
	s := h * float64(xOffset)
	l := 1 - (float64(yOffset) / float64(pickerGradientHeight))
	setting.currentColor.S = s
	setting.currentColor.L = l
	setting.colorUpdated()
}

func (setting *ColorPickerSetting) updateGradients() {
	if err := setting.updatePickerGradient(); err != nil {
		panic(err)
	}
	if err := setting.updateHueGradient(); err != nil {
		panic(err)
	}
}

func (setting *ColorPickerSetting) updatePickerGradient() error {
	pickerW := setting.bbox.W - (gradientPadding * 2)
	if setting.currentPickerGradient.texture != nil {
		setting.currentPickerGradient.texture.Destroy()
	}

	texture, err := setting.lastRenderer.CreateTexture(
		sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_TARGET,
		pickerW, pickerGradientHeight,
	)
	if err != nil {
		return err
	}
	if err = setting.lastRenderer.SetRenderTarget(texture); err != nil {
		return err
	}
	saturationLinspace := utils.Linspace(0, 1, int(pickerW))
	lightnessLinspace := utils.Linspace(1, 0, int(pickerGradientHeight))
	for y := int32(0); y < pickerGradientHeight; y++ {
		for x := int32(0); x < pickerW; x++ {
			color := hslToRGB(setting.currentColor.H, saturationLinspace[x], lightnessLinspace[y])
			utils.DrawPoint(setting.lastRenderer, &sdl.Point{X: x, Y: y}, color)
		}
	}
	if err = setting.lastRenderer.SetRenderTarget(nil); err != nil {
		panic(err)
	}
	setting.currentPickerGradient = cachedGradient{
		texture: texture,
		bbox: &sdl.Rect{
			X: setting.bbox.X + gradientPadding, Y: setting.bbox.Y + gradientPadding,
			W: pickerW, H: pickerGradientHeight,
		},
	}
	return nil
}

func (setting *ColorPickerSetting) updateHueGradient() error {
	hueW := setting.bbox.W - (gradientPadding * 2)
	if setting.currentHueGradient.texture != nil {
		setting.currentHueGradient.texture.Destroy()
	}

	texture, err := setting.lastRenderer.CreateTexture(
		sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_TARGET,
		hueW, hueGradientHeight,
	)
	if err != nil {
		return err
	}
	if err = setting.lastRenderer.SetRenderTarget(texture); err != nil {
		return err
	}
	hueLinspace := utils.Linspace(0, 360, int(hueW))
	for y := int32(0); y < hueGradientHeight; y++ {
		for x := int32(0); x < hueW; x++ {
			color := hslToRGB(hueLinspace[x], 1, 0.5)
			utils.DrawPoint(setting.lastRenderer, &sdl.Point{X: x, Y: y}, color)
		}
	}
	if err = setting.lastRenderer.SetRenderTarget(nil); err != nil {
		panic(err)
	}
	setting.currentHueGradient = cachedGradient{
		texture: texture,
		bbox: &sdl.Rect{
			X: setting.bbox.X + gradientPadding, Y: setting.bbox.Y + pickerGradientHeight + (gradientPadding * 2),
			W: hueW, H: hueGradientHeight,
		},
	}
	return nil
}

func (setting ColorPickerSetting) colorUpdated() {
	setting.onColorUpdated(
		hslToRGB(setting.currentColor.H, setting.currentColor.S, setting.currentColor.L),
	)
}

func (setting ColorPickerSetting) CurrentColor() sdl.Color {
	return hslToRGB(setting.currentColor.H, setting.currentColor.S, setting.currentColor.L)
}

type cachedGradient struct {
	texture *sdl.Texture
	bbox    *sdl.Rect
}

type hslColor struct {
	H float64
	S float64
	L float64
}

func (color hslColor) xOffset(pickerWidth int32) int32 {
	h := 1 / (float64(pickerWidth) - 1)
	return int32(color.S / h)
}

func (color hslColor) yOffset(pickerHeight int32) int32 {
	fHeight := float64(pickerHeight)
	return int32(fHeight - (fHeight * color.L))
}

func (color hslColor) hueOffset(hueWidth int32) int32 {
	h := 360 / (float64(hueWidth) - 1)
	return int32(color.H / h)
}

func hslToRGB(h, s, l float64) sdl.Color {
	C := (1 - utils.Abs(l*2-1)) * s
	H := h / 60
	X := C * (1 - utils.Abs(math.Mod(H, 2)-1))
	var r, g, b float64
	switch {
	case H >= 0 && H <= 1:
		r, g, b = C, X, 0
	case H >= 1 && H <= 2:
		r, g, b = X, C, 0
	case H >= 2 && H <= 3:
		r, g, b = 0, C, X
	case H >= 3 && H <= 4:
		r, g, b = 0, X, C
	case h >= 4 && H <= 5:
		r, g, b = X, 0, C
	case h >= 5 && H <= 6:
		r, g, b = C, 0, X
	}
	m := l - C/2
	return sdl.Color{R: uint8((r + m) * 255), G: uint8((g + m) * 255), B: uint8((b + m) * 255), A: 255}
}
