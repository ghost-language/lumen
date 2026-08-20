package engine

import (
	"fmt"
	"math"

	"ghostlang.org/x/ghost/object"
	"github.com/veandco/go-sdl2/sdl"
)

// Color represents an RGBA color.
//
// Lumen's convention is that the red, green, and blue channels run 0-255, and
// alpha runs 0-1. Splitting the two ranges is deliberate: a single range that
// tries to accept both, guessing from whether a value looks like a fraction,
// cannot tell 1 (nearly transparent) from 1.0 (fully opaque), and gets it
// silently wrong either way. Channels are how image editors and hex codes
// describe color; alpha is how everyone describes opacity.
type Color struct {
	Red   uint8
	Green uint8
	Blue  uint8
	Alpha uint8
}

// NewColor builds a color from 0-255 components.
func NewColor(red, green, blue, alpha uint8) *Color {
	return &Color{Red: red, Green: green, Blue: blue, Alpha: alpha}
}

// SDL converts the color into the struct SDL's drawing calls take.
func (color *Color) SDL() sdl.Color {
	return sdl.Color{R: color.Red, G: color.Green, B: color.Blue, A: color.Alpha}
}

// String represents the color object's value as a string.
func (color *Color) String() string {
	return fmt.Sprintf("Color: rgba(%d, %d, %d, %d)", color.Red, color.Green, color.Blue, color.Alpha)
}

// Type returns the color object type.
func (color *Color) Type() object.Type {
	return COLOR
}

// Method defines the set of methods available on color objects.
func (color *Color) Method(method string, args []object.Object) (object.Object, bool) {
	switch method {
	case "getRed":
		return object.NewInt(int64(color.Red)), true
	case "getGreen":
		return object.NewInt(int64(color.Green)), true
	case "getBlue":
		return object.NewInt(int64(color.Blue)), true
	case "getAlpha":
		return object.NewInt(int64(color.Alpha)), true
	case "toHex":
		return &object.String{Value: color.Hex()}, true
	case "withAlpha":
		return color.withAlpha(args)
	case "lerp":
		return color.lerp(args)
	case "toString":
		return &object.String{Value: color.String()}, true
	}

	return nil, false
}

// Hex renders the color as an #rrggbbaa string.
func (color *Color) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", color.Red, color.Green, color.Blue, color.Alpha)
}

// =============================================================================
// Object methods

// withAlpha returns a copy of the color with a replaced alpha component.
func (color *Color) withAlpha(args []object.Object) (object.Object, bool) {
	if len(args) != 1 {
		return object.NewError("color.withAlpha() expects 1 argument. got=%d", len(args)), true
	}

	alpha, ok := args[0].(*object.Number)

	if !ok {
		return object.NewError("color.withAlpha() expects a number"), true
	}

	return NewColor(color.Red, color.Green, color.Blue, ColorAlpha(alpha)), true
}

// lerp blends the color toward another color by the given amount (0-1).
func (color *Color) lerp(args []object.Object) (object.Object, bool) {
	if len(args) != 2 {
		return object.NewError("color.lerp() expects 2 arguments. got=%d", len(args)), true
	}

	target, ok := args[0].(*Color)

	if !ok {
		return object.NewError("color.lerp() expects a color as its first argument"), true
	}

	amount, ok := args[1].(*object.Number)

	if !ok {
		return object.NewError("color.lerp() expects a number as its second argument"), true
	}

	blend := math.Max(0, math.Min(1, amount.Float64()))

	mix := func(from, to uint8) uint8 {
		return uint8(math.Round(float64(from) + (float64(to)-float64(from))*blend))
	}

	return NewColor(
		mix(color.Red, target.Red),
		mix(color.Green, target.Green),
		mix(color.Blue, target.Blue),
		mix(color.Alpha, target.Alpha),
	), true
}

// ColorComponent converts a Ghost number into a 0-255 color channel.
func ColorComponent(number *object.Number) uint8 {
	return uint8(math.Max(0, math.Min(255, math.Round(number.Float64()))))
}

// ColorAlpha converts a Ghost number into an alpha byte, reading it as an
// opacity between 0 and 1.
func ColorAlpha(number *object.Number) uint8 {
	return uint8(math.Max(0, math.Min(1, number.Float64())) * 255)
}
