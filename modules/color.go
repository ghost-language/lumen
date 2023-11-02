package modules

import (
	"strconv"
	"strings"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
)

var ColorMethods = map[string]*object.LibraryFunction{}
var ColorProperties = map[string]*object.LibraryProperty{}

func init() {
	// Methods
	modules.RegisterMethod(ColorMethods, "rgb", colorRgbMethod)
	modules.RegisterMethod(ColorMethods, "hex", colorHexMethod)

	// Properties
	modules.RegisterProperty(ColorProperties, "black", colorBlackProperty)
	modules.RegisterProperty(ColorProperties, "white", colorWhiteProperty)
}

func colorRgbMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 3 {
		return object.NewError("wrong number of arguments. got=%d, want=3", len(args))
	}

	red := uint8(args[0].(*object.Number).Value.IntPart())
	green := uint8(args[1].(*object.Number).Value.IntPart())
	blue := uint8(args[2].(*object.Number).Value.IntPart())
	alpha := uint8(255)

	if len(args) == 4 {
		alpha = uint8(args[3].(*object.Number).Value.IntPart())
	}

	color := new(engine.Color)

	color.Red = red
	color.Green = green
	color.Blue = blue
	color.Alpha = alpha

	return color
}

func colorHexMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	hex := args[0].(*object.String).Value

	hex = strings.TrimPrefix(hex, "#")

	var red, green, blue, alpha uint8

	// Default alpha value
	alpha = 255

	if len(hex) == 3 {
		r, _ := strconv.ParseInt(string(hex[0])+string(hex[0]), 16, 32)
		g, _ := strconv.ParseInt(string(hex[1])+string(hex[1]), 16, 32)
		b, _ := strconv.ParseInt(string(hex[2])+string(hex[2]), 16, 32)

		red = uint8(r)
		green = uint8(g)
		blue = uint8(b)
	} else if len(hex) == 6 {
		r, _ := strconv.ParseInt(hex[0:2], 16, 32)
		g, _ := strconv.ParseInt(hex[2:4], 16, 32)
		b, _ := strconv.ParseInt(hex[4:6], 16, 32)

		red = uint8(r)
		green = uint8(g)
		blue = uint8(b)
	} else if len(hex) == 8 {
		r, _ := strconv.ParseInt(hex[0:2], 16, 32)
		g, _ := strconv.ParseInt(hex[2:4], 16, 32)
		b, _ := strconv.ParseInt(hex[4:6], 16, 32)
		a, _ := strconv.ParseInt(hex[6:8], 16, 32)

		red = uint8(r)
		green = uint8(g)
		blue = uint8(b)
		alpha = uint8(a)
	} else {
		return object.NewError("invalid hex value")
	}

	color := new(engine.Color)

	color.Red = red
	color.Green = green
	color.Blue = blue
	color.Alpha = alpha

	return color
}

func colorBlackProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(engine.Color)

	color.Red = 0
	color.Green = 0
	color.Blue = 0
	color.Alpha = 255

	return color
}

func colorWhiteProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(engine.Color)

	color.Red = 255
	color.Green = 255
	color.Blue = 255
	color.Alpha = 255

	return color
}
