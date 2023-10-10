package keyboard

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"github.com/veandco/go-sdl2/sdl"
)

var KeyboardMethods = map[string]*object.LibraryFunction{}
var KeyboardProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(KeyboardMethods, "isDown", keyboardIsDown)
	modules.RegisterMethod(KeyboardMethods, "isUp", keyboardIsUp)
	modules.RegisterMethod(KeyboardMethods, "isPressed", keyboardIsPressed)
	modules.RegisterMethod(KeyboardMethods, "isReleased", keyboardIsReleased)
}

func keyboardIsDown(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	scancode := sdl.GetScancodeFromName(args[0].String())

	if IsDown(scancode) {
		return value.TRUE
	}

	return value.FALSE
}

func keyboardIsUp(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	scancode := sdl.GetScancodeFromName(args[0].String())

	if IsUp(scancode) {
		return value.TRUE
	}

	return value.FALSE
}

func keyboardIsPressed(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	scancode := sdl.GetScancodeFromName(args[0].String())

	if IsPressed(scancode) {
		return value.TRUE
	}

	return value.FALSE
}

func keyboardIsReleased(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	scancode := sdl.GetScancodeFromName(args[0].String())

	if IsReleased(scancode) {
		return value.TRUE
	}

	return value.FALSE
}
