package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
	"github.com/shopspring/decimal"
	"github.com/veandco/go-sdl2/sdl"
)

var MouseMethods = map[string]*object.LibraryFunction{}
var MouseProperties = map[string]*object.LibraryProperty{}

func init() {
	// Methods
	modules.RegisterMethod(MouseMethods, "showCursor", mouseShowCursorMethod)
	modules.RegisterMethod(MouseMethods, "hideCursor", mouseHideCursorMethod)
	modules.RegisterMethod(MouseMethods, "isButtonDown", mouseIsButtonDownMethod)
	modules.RegisterMethod(MouseMethods, "isButtonUp", mouseIsButtonUpMethod)
	modules.RegisterMethod(MouseMethods, "wasButtonPressed", mouseWasButtonPressedMethod)
	modules.RegisterMethod(MouseMethods, "wasButtonReleased", mouseWasButtonReleasedMethod)

	// Properties
	modules.RegisterProperty(MouseProperties, "x", mouseXProperty)
	modules.RegisterProperty(MouseProperties, "y", mouseYProperty)
}

func mouseShowCursorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	sdl.ShowCursor(sdl.ENABLE)

	return nil
}

func mouseHideCursorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	sdl.ShowCursor(sdl.DISABLE)

	return nil
}

func mouseIsButtonDownMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	var mask uint32

	button := args[0].(*object.String)

	switch button.Value {
	case "left":
		mask = 1
	case "middle":
		mask = 2
	case "right":
		mask = 4
	}

	isDown := engine.Lumen.CurrentMouseState & mask

	return &object.Boolean{Value: isDown != 0}
}

func mouseIsButtonUpMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	var mask uint32

	button := args[0].(*object.String)

	switch button.Value {
	case "left":
		mask = 1
	case "middle":
		mask = 2
	case "right":
		mask = 4
	}

	isDown := engine.Lumen.CurrentMouseState & mask

	return &object.Boolean{Value: isDown == 0}
}

func mouseWasButtonPressedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	var mask uint32

	button := args[0].(*object.String)

	switch button.Value {
	case "left":
		mask = 1
	case "middle":
		mask = 2
	case "right":
		mask = 4
	}

	isDown := engine.Lumen.CurrentMouseState & mask
	wasDown := engine.Lumen.PreviousMouseState & mask

	return &object.Boolean{Value: isDown != 0 && wasDown == 0}
}

func mouseWasButtonReleasedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	var mask uint32

	button := args[0].(*object.String)

	switch button.Value {
	case "left":
		mask = 1
	case "middle":
		mask = 2
	case "right":
		mask = 4
	}

	isDown := engine.Lumen.CurrentMouseState & mask
	wasDown := engine.Lumen.PreviousMouseState & mask

	return &object.Boolean{Value: isDown == 0 && wasDown != 0}
}

func mouseXProperty(scope *object.Scope, tok token.Token) object.Object {
	x, _, _ := sdl.GetMouseState()

	return &object.Number{Value: decimal.NewFromInt(int64(x))}
}

func mouseYProperty(scope *object.Scope, tok token.Token) object.Object {
	_, y, _ := sdl.GetMouseState()

	return &object.Number{Value: decimal.NewFromInt(int64(y))}
}
