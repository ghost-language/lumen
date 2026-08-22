package engine

import (
	"fmt"
	"math"

	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/value"
)

// AnimationMode decides what an animation does when it runs off its last frame.
type AnimationMode int

const (
	// AnimationLoop restarts from the first frame and never finishes.
	AnimationLoop AnimationMode = iota

	// AnimationOnce holds on the last frame and reports itself finished.
	AnimationOnce

	// AnimationPingpong plays forwards then backwards, and never finishes.
	AnimationPingpong
)

// AnimationModeFromName maps the mode names a game writes onto the enum.
func AnimationModeFromName(name string) (AnimationMode, bool) {
	switch name {
	case "loop":
		return AnimationLoop, true
	case "once":
		return AnimationOnce, true
	case "pingpong", "pingPong", "bounce":
		return AnimationPingpong, true
	}

	return AnimationLoop, false
}

// Name returns the mode's name as a game writes it.
func (mode AnimationMode) Name() string {
	switch mode {
	case AnimationOnce:
		return "once"
	case AnimationPingpong:
		return "pingpong"
	}

	return "loop"
}

// Animation plays a sequence of a spritesheet's frames against a clock.
//
// The clock is driven by delta time rather than by frames drawn. Counting draws
// instead — as the obvious version of this does, and as two of Lumen's own
// examples did — ties playback speed to frame rate, so the same walk cycle runs
// at half speed on a machine managing 30 FPS.
//
// An animation carries its own playhead, so two characters showing the same
// walk cycle need one each. clone() is the cheap way to get one.
type Animation struct {
	Sheet    *Spritesheet
	Frames   []int32
	Duration float64
	Mode     AnimationMode

	elapsed float64
	speed   float64
	playing bool
}

// NewAnimation builds an animation over a sheet's frames, at a number of
// seconds per frame.
func NewAnimation(sheet *Spritesheet, frames []int32, duration float64, mode AnimationMode) *Animation {
	return &Animation{
		Sheet:    sheet,
		Frames:   frames,
		Duration: duration,
		Mode:     mode,
		speed:    1,
		playing:  true,
	}
}

// Update advances the playhead by a number of seconds.
func (animation *Animation) Update(delta float64) {
	if !animation.playing || animation.Finished() {
		return
	}

	animation.elapsed += delta * animation.speed

	if animation.elapsed < 0 {
		animation.elapsed = 0
	}
}

// Length returns how long one pass through the animation takes, in seconds. A
// ping-pong pass includes the way back, minus the two frames it does not repeat.
func (animation *Animation) Length() float64 {
	return float64(animation.steps()) * animation.Duration
}

// Finished reports whether a 'once' animation has reached its last frame.
// Looping and ping-ponging animations never finish.
func (animation *Animation) Finished() bool {
	if animation.Mode != AnimationOnce || animation.Duration <= 0 {
		return false
	}

	return animation.elapsed >= animation.Length()
}

// Frame returns the sheet frame currently showing.
func (animation *Animation) Frame() int32 {
	count := len(animation.Frames)

	if count == 0 {
		return 0
	}

	// A duration of zero is a still pose rather than a division by zero.
	if animation.Duration <= 0 {
		return animation.Frames[0]
	}

	step := int(math.Floor(animation.elapsed / animation.Duration))

	if step < 0 {
		step = 0
	}

	switch animation.Mode {
	case AnimationOnce:
		if step >= count {
			step = count - 1
		}

	case AnimationPingpong:
		// One pass out and back visits 2n-2 steps: the first and last frames are
		// the turning points and are not played twice in a row.
		period := animation.steps()

		step = step % period

		if step >= count {
			step = period - step
		}

	default:
		step = step % count
	}

	return animation.Frames[step]
}

// Draw renders the frame the animation is currently showing.
func (animation *Animation) Draw(arguments DrawArguments) {
	animation.Sheet.Draw(animation.Frame(), arguments)
}

// Clone returns a copy of the animation with its own playhead, sharing the sheet
// and the frame list it was made from.
func (animation *Animation) Clone() *Animation {
	copied := *animation
	copied.elapsed = 0
	copied.playing = true

	return &copied
}

// steps returns how many playhead steps one pass through the animation takes.
func (animation *Animation) steps() int {
	count := len(animation.Frames)

	if animation.Mode == AnimationPingpong && count > 1 {
		return 2*count - 2
	}

	if count == 0 {
		return 1
	}

	return count
}

// String represents the animation object's value as a string.
func (animation *Animation) String() string {
	return fmt.Sprintf("Animation: %d frames, %gs each, %s", len(animation.Frames), animation.Duration, animation.Mode.Name())
}

// Type returns the animation object type.
func (animation *Animation) Type() object.Type {
	return ANIMATION
}

// Method defines the set of methods available on animation objects.
func (animation *Animation) Method(method string, args []object.Object) (object.Object, bool) {
	switch method {
	case "update":
		return animation.update(args)
	case "draw":
		return animation.draw(args)
	case "play":
		animation.playing = true

		return value.NULL, true
	case "pause":
		animation.playing = false

		return value.NULL, true
	case "resume":
		animation.playing = true

		return value.NULL, true
	case "stop":
		animation.playing = false
		animation.elapsed = 0

		return value.NULL, true
	case "reset":
		animation.elapsed = 0
		animation.playing = true

		return value.NULL, true
	case "seek":
		return animation.seek(args)
	case "clone":
		return animation.Clone(), true
	case "isPlaying":
		return &object.Boolean{Value: animation.playing && !animation.Finished()}, true
	case "isPaused":
		return &object.Boolean{Value: !animation.playing}, true
	case "isFinished":
		return &object.Boolean{Value: animation.Finished()}, true
	case "getFrame":
		return object.NewInt(int64(animation.Frame())), true
	case "getFrameCount":
		return object.NewInt(int64(len(animation.Frames))), true
	case "getElapsed":
		return object.NewFloat(animation.elapsed), true
	case "getLength":
		return object.NewFloat(animation.Length()), true
	case "getSheet":
		return animation.Sheet, true
	case "getDuration":
		return object.NewFloat(animation.Duration), true
	case "setDuration":
		return animation.setDuration(args)
	case "getSpeed":
		return object.NewFloat(animation.speed), true
	case "setSpeed":
		return animation.setSpeed(args)
	case "getMode":
		return &object.String{Value: animation.Mode.Name()}, true
	case "setMode":
		return animation.setMode(args)
	case "toString":
		return &object.String{Value: animation.String()}, true
	}

	return nil, false
}

// =============================================================================
// Object methods

// update advances the playhead. Called with no argument it uses the time the
// last frame took, which is what a game wants often enough to be the default.
func (animation *Animation) update(args []object.Object) (object.Object, bool) {
	delta := Lumen.Delta

	if len(args) > 0 {
		given, err := Float("animation.update", args, 0)

		if err != nil {
			return err, true
		}

		delta = given
	}

	animation.Update(delta)

	return value.NULL, true
}

// draw renders the current frame: animation.draw(x, y, rotation, sx, sy, ox, oy).
func (animation *Animation) draw(args []object.Object) (object.Object, bool) {
	arguments, err := ParseDrawArguments("animation.draw", args, 0)

	if err != nil {
		return err, true
	}

	animation.Draw(arguments)

	return value.NULL, true
}

// seek moves the playhead to a number of seconds into the animation.
func (animation *Animation) seek(args []object.Object) (object.Object, bool) {
	seconds, err := Float("animation.seek", args, 0)

	if err != nil {
		return err, true
	}

	if seconds < 0 {
		seconds = 0
	}

	animation.elapsed = seconds

	return value.NULL, true
}

// setDuration changes how long each frame is held.
func (animation *Animation) setDuration(args []object.Object) (object.Object, bool) {
	duration, err := Float("animation.setDuration", args, 0)

	if err != nil {
		return err, true
	}

	animation.Duration = duration

	return value.NULL, true
}

// setSpeed multiplies the rate the playhead advances at. A speed of 2 plays
// twice as fast; a negative speed rewinds.
func (animation *Animation) setSpeed(args []object.Object) (object.Object, bool) {
	speed, err := Float("animation.setSpeed", args, 0)

	if err != nil {
		return err, true
	}

	animation.speed = speed

	return value.NULL, true
}

// setMode changes what happens at the end of the sequence.
func (animation *Animation) setMode(args []object.Object) (object.Object, bool) {
	if len(args) != 1 {
		return object.NewError("animation.setMode() expects 1 argument. got=%d", len(args)), true
	}

	name, ok := args[0].(*object.String)

	if !ok {
		return object.NewError("animation.setMode() expects a string. got=%s", args[0].Type()), true
	}

	mode, valid := AnimationModeFromName(name.Value)

	if !valid {
		return object.NewError("animation.setMode() expects 'loop', 'once', or 'pingpong'. got=%s", name.Value), true
	}

	animation.Mode = mode

	return value.NULL, true
}
