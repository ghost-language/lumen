package color

import (
	"strconv"
	"strings"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
)

var ColorMethods = map[string]*object.LibraryFunction{}
var ColorProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(ColorMethods, "rgb", colorRgbMethod)
	modules.RegisterMethod(ColorMethods, "hex", colorHexMethod)

	modules.RegisterProperty(ColorProperties, "black", colorBlackProperty)
	modules.RegisterProperty(ColorProperties, "darkBlue", colorDarkBlueProperty)
	modules.RegisterProperty(ColorProperties, "darkPurple", colorDarkPurpleProperty)
	modules.RegisterProperty(ColorProperties, "darkGreen", colorDarkGreenProperty)
	modules.RegisterProperty(ColorProperties, "brown", colorBrownProperty)
	modules.RegisterProperty(ColorProperties, "darkGray", colorDarkGrayProperty)
	modules.RegisterProperty(ColorProperties, "lightGray", colorLightGrayProperty)
	modules.RegisterProperty(ColorProperties, "white", colorWhiteProperty)
	modules.RegisterProperty(ColorProperties, "red", colorRedProperty)
	modules.RegisterProperty(ColorProperties, "orange", colorOrangeProperty)
	modules.RegisterProperty(ColorProperties, "yellow", colorYellowProperty)
	modules.RegisterProperty(ColorProperties, "green", colorGreenProperty)
	modules.RegisterProperty(ColorProperties, "blue", colorBlueProperty)
	modules.RegisterProperty(ColorProperties, "lavender", colorLavenderProperty)
	modules.RegisterProperty(ColorProperties, "pink", colorPinkProperty)
	modules.RegisterProperty(ColorProperties, "peach", colorPeachProperty)
}

func colorRgbMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 3 {
		panic("wrong number of arguments. expected=3 or 4")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	red := uint8(args[0].(*object.Number).Value.IntPart())
	green := uint8(args[1].(*object.Number).Value.IntPart())
	blue := uint8(args[2].(*object.Number).Value.IntPart())
	alpha := uint8(255)

	if len(args) == 4 {
		alpha = uint8(args[3].(*object.Number).Value.IntPart())
	}

	color := new(Color)

	color.Red = red
	color.Green = green
	color.Blue = blue
	color.Alpha = alpha

	return color
}

func colorHexMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=1", len(args))
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
		panic("invalid hex value")
		// return object.NewError("invalid hex value")
	}

	color := new(Color)

	color.Red = red
	color.Green = green
	color.Blue = blue
	color.Alpha = alpha

	return color
}

func colorBlackProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 0
	color.Green = 0
	color.Blue = 0
	color.Alpha = 255

	return color
}

func colorDarkBlueProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 29
	color.Green = 43
	color.Blue = 83
	color.Alpha = 255

	return color
}

func colorDarkPurpleProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 126
	color.Green = 37
	color.Blue = 83
	color.Alpha = 255

	return color
}

func colorDarkGreenProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 0
	color.Green = 135
	color.Blue = 81
	color.Alpha = 255

	return color
}

func colorBrownProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 171
	color.Green = 82
	color.Blue = 54
	color.Alpha = 255

	return color
}

func colorDarkGrayProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 95
	color.Green = 87
	color.Blue = 79
	color.Alpha = 255

	return color
}

func colorLightGrayProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 194
	color.Green = 195
	color.Blue = 199
	color.Alpha = 255

	return color
}

func colorWhiteProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 255
	color.Green = 255
	color.Blue = 255
	color.Alpha = 255

	return color
}

func colorRedProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 255
	color.Green = 0
	color.Blue = 77
	color.Alpha = 255

	return color
}

func colorOrangeProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 255
	color.Green = 163
	color.Blue = 0
	color.Alpha = 255

	return color
}

func colorYellowProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 255
	color.Green = 236
	color.Blue = 39
	color.Alpha = 255

	return color
}

func colorGreenProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 0
	color.Green = 228
	color.Blue = 54
	color.Alpha = 255

	return color
}

func colorBlueProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 41
	color.Green = 173
	color.Blue = 255
	color.Alpha = 255

	return color
}

func colorLavenderProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 131
	color.Green = 118
	color.Blue = 156
	color.Alpha = 255

	return color
}

func colorPinkProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 255
	color.Green = 119
	color.Blue = 168
	color.Alpha = 255

	return color
}

func colorPeachProperty(scope *object.Scope, tok token.Token) object.Object {
	color := new(Color)

	color.Red = 255
	color.Green = 204
	color.Blue = 170
	color.Alpha = 255

	return color
}
