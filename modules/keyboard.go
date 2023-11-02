package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
	"github.com/veandco/go-sdl2/sdl"
)

var KeyboardMethods = map[string]*object.LibraryFunction{}
var KeyboardProperties = map[string]*object.LibraryProperty{}

func init() {
	// Methods
	modules.RegisterMethod(KeyboardMethods, "isDown", keyboardIsDownMethod)
	modules.RegisterMethod(KeyboardMethods, "isUp", keyboardIsUpMethod)
	modules.RegisterMethod(KeyboardMethods, "wasPressed", keyboardWasPressedMethod)
	modules.RegisterMethod(KeyboardMethods, "wasReleased", keyboardWasReleasedMethod)
}

func keyboardIsDownMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	scancode := sdl.GetScancodeFromName(args[0].String())

	if engine.Lumen.CurrentKeyboardState[scancode] == 1 {
		return value.TRUE
	}

	return value.FALSE
}

func keyboardIsUpMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	scancode := sdl.GetScancodeFromName(args[0].String())

	if engine.Lumen.CurrentKeyboardState[scancode] == 0 {
		return value.TRUE
	}

	return value.FALSE
}

func keyboardWasPressedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	scancode := sdl.GetScancodeFromName(args[0].String())

	if engine.Lumen.CurrentKeyboardState[scancode] == 1 && engine.Lumen.PreviousKeyboardState[scancode] == 0 {
		return value.TRUE
	}

	return value.FALSE
}

func keyboardWasReleasedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	scancode := sdl.GetScancodeFromName(args[0].String())

	if engine.Lumen.CurrentKeyboardState[scancode] == 0 && engine.Lumen.PreviousKeyboardState[scancode] == 1 {
		return value.TRUE
	}

	return value.FALSE
}
