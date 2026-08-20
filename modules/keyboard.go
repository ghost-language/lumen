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
	modules.RegisterMethod(KeyboardMethods, "isDown", keyboardIsDownMethod)
	modules.RegisterMethod(KeyboardMethods, "isUp", keyboardIsUpMethod)
	modules.RegisterMethod(KeyboardMethods, "wasPressed", keyboardWasPressedMethod)
	modules.RegisterMethod(KeyboardMethods, "wasReleased", keyboardWasReleasedMethod)
	modules.RegisterMethod(KeyboardMethods, "startTextInput", keyboardStartTextInputMethod)
	modules.RegisterMethod(KeyboardMethods, "stopTextInput", keyboardStopTextInputMethod)
	modules.RegisterMethod(KeyboardMethods, "isTextInputActive", keyboardIsTextInputActiveMethod)
}

// keyboardIsDownMethod reports whether any of the named keys is held. Several
// keys can be passed at once, which is how a game maps one action onto both the
// arrow keys and WASD without repeating itself.
func keyboardIsDownMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return anyKey("keyboard.isDown", tok, args, func(scancode sdl.Scancode) bool {
		return engine.Lumen.IsKeyDown(scancode)
	})
}

func keyboardIsUpMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return anyKey("keyboard.isUp", tok, args, func(scancode sdl.Scancode) bool {
		return !engine.Lumen.IsKeyDown(scancode)
	})
}

// keyboardWasPressedMethod reports whether a key went down during this frame,
// which is what menus and dialogue advances should use so one keypress counts
// once rather than once per frame the key is held.
func keyboardWasPressedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return anyKey("keyboard.wasPressed", tok, args, func(scancode sdl.Scancode) bool {
		return engine.Lumen.IsKeyDown(scancode) && !engine.Lumen.WasKeyDown(scancode)
	})
}

func keyboardWasReleasedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return anyKey("keyboard.wasReleased", tok, args, func(scancode sdl.Scancode) bool {
		return !engine.Lumen.IsKeyDown(scancode) && engine.Lumen.WasKeyDown(scancode)
	})
}

// keyboardStartTextInputMethod turns on text events, which drive the textinput()
// callback. It also enables the platform's on-screen keyboard where there is one.
func keyboardStartTextInputMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	sdl.StartTextInput()

	return value.NULL
}

func keyboardStopTextInputMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	sdl.StopTextInput()

	return value.NULL
}

func keyboardIsTextInputActiveMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return &object.Boolean{Value: sdl.IsTextInputActive()}
}

// anyKey resolves each key name and reports whether the predicate holds for any
// of them. An unrecognised key name is an error rather than a silent false,
// because a typo in a key name is otherwise almost impossible to spot.
func anyKey(name string, tok token.Token, args []object.Object, predicate func(sdl.Scancode) bool) object.Object {
	if len(args) == 0 {
		return object.NewError("%d:%d: runtime error: %s() expects at least one key name", tok.Line, tok.Column, name)
	}

	for index := range args {
		key, err := text(name, tok, args, index)

		if err != nil {
			return err
		}

		scancode, ok := engine.Scancode(key)

		if !ok {
			return object.NewError("%d:%d: runtime error: %s() does not recognise the key '%s'", tok.Line, tok.Column, name, key)
		}

		if predicate(scancode) {
			return value.TRUE
		}
	}

	return value.FALSE
}
