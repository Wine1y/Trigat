package settings

import (
	"math"

	"github.com/Wine1y/trigat/internal/gui"
	"github.com/Wine1y/trigat/pkg"
	"github.com/veandco/go-sdl2/sdl"
)

const sliderHeight int32 = 30
const trackRadius int32 = 6
const trackHeight int32 = 8
const thumbHeight int32 = 16

var trackColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
var thumbFillColor = sdl.Color{R: 245, G: 245, B: 245, A: 255}
var thumbOutlineColor = sdl.Color{R: 0, G: 0, B: 0, A: 100}

type SliderSetting struct {
	*DefaultSetting
	minValue       uint
	maxValue       uint
	currentValue   uint
	track          sdl.Rect
	thumb          sdl.Rect
	isMoving       bool
	onValueUpdated func(value uint)
}

func NewSliderSetting(minValue, maxValue uint, onValueUpdated func(value uint)) *SliderSetting {
	if minValue == 0 {
		panic("Slider value can't be 0 or lower")
	}
	return &SliderSetting{
		DefaultSetting: NewDefaultSetting(sliderHeight),
		minValue:       minValue,
		maxValue:       maxValue,
		currentValue:   (minValue + maxValue) / 2,
		isMoving:       false,
		onValueUpdated: onValueUpdated,
	}
}

func (setting SliderSetting) Render(ren *sdl.Renderer) {
	pkg.DrawRoundedFilledRectangle(
		ren,
		&setting.track,
		trackRadius,
		trackColor,
	)
	pkg.DrawFilledCircle(
		ren,
		&sdl.Point{X: setting.thumb.X + thumbHeight/2, Y: setting.thumb.Y + thumbHeight/2},
		thumbHeight/2,
		thumbFillColor,
	)
	pkg.DrawCircle(
		ren,
		&sdl.Point{X: setting.thumb.X + thumbHeight/2, Y: setting.thumb.Y + thumbHeight/2},
		thumbHeight/2,
		thumbOutlineColor,
	)
}

func (setting *SliderSetting) SettingCallbacks() *gui.WindowCallbackSet {
	callbacks := gui.NewWindowCallbackSet()
	callbacks.MouseDown = append(callbacks.MouseDown, func(button uint8, x, y int32) bool {
		click := sdl.Point{X: x, Y: y}
		if button == sdl.BUTTON_LEFT && click.InRect(&setting.thumb) {
			setting.isMoving = true
		} else if button == sdl.BUTTON_LEFT && click.InRect(&setting.track) {
			setting.updateValue(x)
		}
		if click.InRect(&setting.bbox) {
			return true
		}
		return false
	})
	callbacks.MouseUp = append(callbacks.MouseUp, func(button uint8, x, y int32) bool {
		if button == sdl.BUTTON_LEFT && setting.isMoving {
			setting.isMoving = false
		}
		return false
	})
	callbacks.MouseMove = append(callbacks.MouseMove, func(x, y int32) bool {
		move := sdl.Point{X: x, Y: y}
		if !move.InRect(&setting.bbox) && setting.isMoving {
			setting.isMoving = false
		}
		if setting.isMoving {
			setting.updateValue(x)
		}
		return false
	})
	return callbacks
}

func (setting *SliderSetting) SetLeftTop(lt *sdl.Point) {
	setting.DefaultSetting.SetLeftTop(lt)
	setting.resize()
}

func (setting *SliderSetting) SetWidth(width int32) {
	setting.DefaultSetting.SetWidth(width)
	setting.resize()
}

func (setting *SliderSetting) resize() {
	trackX := setting.bbox.X + trackRadius
	trackY := setting.bbox.Y + (setting.bbox.H-trackHeight)/2
	trackW := setting.bbox.W - trackRadius*2
	pixelsPerValue := float64(trackW) / float64(setting.maxValue-setting.minValue)
	thumbOffset := float64(setting.currentValue-1) * pixelsPerValue
	thumbCenter := sdl.Point{X: int32(float64(trackX) + thumbOffset), Y: trackY + trackHeight/2}

	setting.track = sdl.Rect{
		X: trackX, Y: trackY, W: trackW, H: trackHeight,
	}

	setting.thumb = sdl.Rect{
		X: thumbCenter.X - thumbHeight/2, Y: thumbCenter.Y - thumbHeight/2,
		W: thumbHeight, H: thumbHeight,
	}
}

func (setting *SliderSetting) updateValue(x int32) {
	trackX := setting.bbox.X + trackRadius
	trackW := setting.bbox.W - trackRadius*2
	if x < trackX {
		x = trackX
	}
	if x > trackX+trackW {
		x = trackX + trackW
	}
	valuePerPixel := float64(setting.maxValue-setting.minValue) / float64(trackW)
	setting.currentValue = uint(math.Round(float64(x-trackX)*valuePerPixel)) + 1
	setting.onValueUpdated(setting.currentValue)
	setting.resize()
}

func (slider SliderSetting) CurrentValue() uint {
	return slider.currentValue
}
