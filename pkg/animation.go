package pkg

import "time"

type Animation struct {
	startValue     int
	endValue       int
	totalFrames    int
	currentFrame   int
	repeats        int
	isCycled       bool
	timingFunction func(startValue, endValue int, totalFrames int, currentFrame int) int
}

func (animation *Animation) CurrentValue() int {
	currentRepeat := int(float64(animation.currentFrame) / float64(animation.totalFrames))
	if animation.repeats > 0 && currentRepeat >= animation.repeats {
		return animation.endValue
	}
	currentFrame := animation.currentFrame
	if currentRepeat < animation.repeats || animation.repeats == 0 {
		currentFrame -= animation.totalFrames * currentRepeat
	}
	var value int
	if animation.isCycled {
		value = animation.getCycledValue(currentFrame)
	} else {
		value = animation.getValue(currentFrame)
	}
	animation.currentFrame++
	return value
}

func (animation *Animation) getValue(frame int) int {
	return animation.timingFunction(
		animation.startValue, animation.endValue,
		animation.totalFrames, frame,
	)
}

func (animation *Animation) getCycledValue(frame int) int {
	half := animation.totalFrames / 2
	if animation.totalFrames%2 != 0 {
		half++
	}
	if frame < half {
		return linearTimingFunction(animation.startValue, animation.endValue, half, frame)
	} else {
		return linearTimingFunction(animation.endValue, animation.startValue, half, frame-half)
	}
}

func newAnimation(
	startValue,
	endValue,
	fps int,
	duration time.Duration,
	repeats int,
	isCycled bool,
) *Animation {
	return &Animation{
		startValue:     startValue,
		endValue:       endValue,
		totalFrames:    int(duration.Seconds() * float64(fps)),
		currentFrame:   0,
		repeats:        repeats,
		timingFunction: nil,
		isCycled:       isCycled,
	}
}

func NewLinearAnimation(startValue, endValue, fps int, duration time.Duration, repeats int, isCycled bool) *Animation {
	animation := newAnimation(startValue, endValue, fps, duration, repeats, isCycled)
	animation.timingFunction = linearTimingFunction
	return animation
}

func linearTimingFunction(startValue, endValue int, totalFrames int, currentFrame int) int {
	h := float64(endValue-startValue) / float64(totalFrames-1)
	return startValue + int(h*float64(currentFrame))
}
