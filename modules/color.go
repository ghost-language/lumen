package modules

import (
	"math"
	"strconv"
	"strings"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
)

var ColorMethods = map[string]*object.LibraryFunction{}
var ColorProperties = map[string]*object.LibraryProperty{}

// named is the palette exposed as color properties. Having a handful of colors
// ready to hand keeps prototypes readable before a game settles on its own.
var named = map[string]*engine.Color{
	"black":       engine.NewColor(0, 0, 0, 255),
	"white":       engine.NewColor(255, 255, 255, 255),
	"transparent": engine.NewColor(0, 0, 0, 0),
	"red":         engine.NewColor(224, 60, 60, 255),
	"green":       engine.NewColor(72, 184, 96, 255),
	"blue":        engine.NewColor(64, 128, 224, 255),
	"yellow":      engine.NewColor(240, 200, 72, 255),
	"orange":      engine.NewColor(232, 136, 56, 255),
	"purple":      engine.NewColor(150, 96, 208, 255),
	"cyan":        engine.NewColor(72, 200, 208, 255),
	"magenta":     engine.NewColor(216, 88, 168, 255),
	"brown":       engine.NewColor(128, 88, 56, 255),
	"gray":        engine.NewColor(128, 128, 128, 255),
	"lightGray":   engine.NewColor(192, 192, 192, 255),
	"darkGray":    engine.NewColor(64, 64, 64, 255),
}

func init() {
	modules.RegisterMethod(ColorMethods, "rgb", colorRgbMethod)
	modules.RegisterMethod(ColorMethods, "rgba", colorRgbMethod)
	modules.RegisterMethod(ColorMethods, "hex", colorHexMethod)
	modules.RegisterMethod(ColorMethods, "hsl", colorHslMethod)

	for name, color := range named {
		modules.RegisterProperty(ColorProperties, name, namedColorProperty(color))
	}
}

// colorRgbMethod builds a color. Red, green, and blue run 0-255; the optional
// fourth argument is opacity, running 0 to 1.
func colorRgbMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("color.rgb", tok, args, 3, 4); err != nil {
		return err
	}

	components := make([]uint8, 0, 4)

	for index := range args {
		component, ok := args[index].(*object.Number)

		if !ok {
			return object.NewError("%d:%d: runtime error: color.rgb() expects number components. argument %d is %s", tok.Line, tok.Column, index+1, args[index].Type())
		}

		if index == 3 {
			components = append(components, engine.ColorAlpha(component))

			continue
		}

		components = append(components, engine.ColorComponent(component))
	}

	alpha := uint8(255)

	if len(components) == 4 {
		alpha = components[3]
	}

	return engine.NewColor(components[0], components[1], components[2], alpha)
}

// colorHexMethod parses #rgb, #rrggbb, and #rrggbbaa, with or without the hash.
func colorHexMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("color.hex", tok, args, 1); err != nil {
		return err
	}

	given, err := text("color.hex", tok, args, 0)

	if err != nil {
		return err
	}

	hex := strings.TrimPrefix(strings.TrimSpace(given), "#")

	// #rgb is shorthand for #rrggbb, so each digit is doubled.
	if len(hex) == 3 || len(hex) == 4 {
		expanded := make([]byte, 0, len(hex)*2)

		for index := 0; index < len(hex); index++ {
			expanded = append(expanded, hex[index], hex[index])
		}

		hex = string(expanded)
	}

	if len(hex) != 6 && len(hex) != 8 {
		return object.NewError("%d:%d: runtime error: color.hex() expects a 3, 4, 6, or 8 digit hex value. got=%s", tok.Line, tok.Column, given)
	}

	components := make([]uint8, 0, 4)

	for index := 0; index < len(hex); index += 2 {
		component, parseErr := strconv.ParseUint(hex[index:index+2], 16, 8)

		if parseErr != nil {
			return object.NewError("%d:%d: runtime error: color.hex() could not parse '%s'", tok.Line, tok.Column, given)
		}

		components = append(components, uint8(component))
	}

	alpha := uint8(255)

	if len(components) == 4 {
		alpha = components[3]
	}

	return engine.NewColor(components[0], components[1], components[2], alpha)
}

// colorHslMethod builds a color from hue in degrees and saturation and lightness
// in the 0-1 range. Hue is the natural way to walk a palette, which is what
// rainbow effects and damage flashes want.
func colorHslMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("color.hsl", tok, args, 3, 4); err != nil {
		return err
	}

	values, err := numbers("color.hsl", tok, args)

	if err != nil {
		return err
	}

	hue := values[0] / 360
	saturation := values[1]
	lightness := values[2]

	alpha := 1.0

	if len(values) == 4 {
		alpha = math.Max(0, math.Min(1, values[3]))
	}

	red, green, blue := hslToRgb(hue, saturation, lightness)

	return engine.NewColor(uint8(red*255), uint8(green*255), uint8(blue*255), uint8(alpha*255))
}

func hslToRgb(hue, saturation, lightness float64) (float64, float64, float64) {
	if saturation == 0 {
		return lightness, lightness, lightness
	}

	var q float64

	if lightness < 0.5 {
		q = lightness * (1 + saturation)
	} else {
		q = lightness + saturation - lightness*saturation
	}

	p := 2*lightness - q

	return hueToChannel(p, q, hue+1.0/3.0), hueToChannel(p, q, hue), hueToChannel(p, q, hue-1.0/3.0)
}

func hueToChannel(p, q, t float64) float64 {
	if t < 0 {
		t++
	}

	if t > 1 {
		t--
	}

	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	}

	return p
}

// namedColorProperty returns a fresh copy each time so a game that mutates a
// color it read from the palette cannot change the palette for everyone else.
func namedColorProperty(color *engine.Color) object.GoProperty {
	return func(scope *object.Scope, tok token.Token) object.Object {
		return engine.NewColor(color.Red, color.Green, color.Blue, color.Alpha)
	}
}
