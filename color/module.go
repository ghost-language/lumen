package color

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
)

var ColorMethods = map[string]*object.LibraryFunction{}
var ColorProperties = map[string]*object.LibraryProperty{}

func init() {
	// modules.RegisterMethod(ColorMethods, "isDown", keyboardIsDown)

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
