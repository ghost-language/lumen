package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
	"github.com/veandco/go-sdl2/sdl"
)

var MouseMethods = map[string]*object.LibraryFunction{}
var MouseProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(MouseMethods, "showCursor", mouseShowCursorMethod)
	modules.RegisterMethod(MouseMethods, "hideCursor", mouseHideCursorMethod)
	modules.RegisterMethod(MouseMethods, "isVisible", mouseIsVisibleMethod)
	modules.RegisterMethod(MouseMethods, "isButtonDown", mouseIsButtonDownMethod)
	modules.RegisterMethod(MouseMethods, "isButtonUp", mouseIsButtonUpMethod)
	modules.RegisterMethod(MouseMethods, "wasButtonPressed", mouseWasButtonPressedMethod)
	modules.RegisterMethod(MouseMethods, "wasButtonReleased", mouseWasButtonReleasedMethod)
	modules.RegisterMethod(MouseMethods, "getPosition", mouseGetPositionMethod)
	modules.RegisterMethod(MouseMethods, "setPosition", mouseSetPositionMethod)
	modules.RegisterMethod(MouseMethods, "getWorldPosition", mouseGetWorldPositionMethod)
	modules.RegisterMethod(MouseMethods, "setRelativeMode", mouseSetRelativeModeMethod)
	modules.RegisterMethod(MouseMethods, "setGrabbed", mouseSetGrabbedMethod)

	modules.RegisterProperty(MouseProperties, "x", mouseXProperty)
	modules.RegisterProperty(MouseProperties, "y", mouseYProperty)
	modules.RegisterProperty(MouseProperties, "wheel", mouseWheelProperty)
	modules.RegisterProperty(MouseProperties, "wheelX", mouseWheelXProperty)
}

func mouseShowCursorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	sdl.ShowCursor(sdl.ENABLE)

	return value.NULL
}

func mouseHideCursorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	sdl.ShowCursor(sdl.DISABLE)

	return value.NULL
}

func mouseIsVisibleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	state, err := sdl.ShowCursor(sdl.QUERY)

	if err != nil {
		return object.NewError("%d:%d: runtime error: mouse.isVisible() %s", tok.Line, tok.Column, err)
	}

	return &object.Boolean{Value: state == sdl.ENABLE}
}

func mouseIsButtonDownMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return buttonState("mouse.isButtonDown", tok, args, func(current, previous uint32, mask uint32) bool {
		return current&mask != 0
	})
}

func mouseIsButtonUpMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return buttonState("mouse.isButtonUp", tok, args, func(current, previous uint32, mask uint32) bool {
		return current&mask == 0
	})
}

func mouseWasButtonPressedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return buttonState("mouse.wasButtonPressed", tok, args, func(current, previous uint32, mask uint32) bool {
		return current&mask != 0 && previous&mask == 0
	})
}

func mouseWasButtonReleasedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return buttonState("mouse.wasButtonReleased", tok, args, func(current, previous uint32, mask uint32) bool {
		return current&mask == 0 && previous&mask != 0
	})
}

// mouseGetPositionMethod reports the pointer in the coordinate space the game
// draws in, which is the window until the game fixes a logical size.
func mouseGetPositionMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	x, y := engine.Lumen.MousePosition()

	return list(x, y)
}

func mouseSetPositionMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("mouse.setPosition", tok, args, 2); err != nil {
		return err
	}

	values, err := numbers("mouse.setPosition", tok, args)

	if err != nil {
		return err
	}

	x, y := engine.Lumen.ToWindow(values[0], values[1])

	engine.Lumen.Window.WarpMouseInWindow(x, y)

	return value.NULL
}

// mouseGetWorldPositionMethod maps the pointer through the inverse of the
// current transform, so a game with a camera can ask where the mouse is in the
// world rather than on the screen.
func mouseGetWorldPositionMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	inverse, ok := engine.Lumen.Graphics.Transform().Inverse()

	if !ok {
		return object.NewError("%d:%d: runtime error: mouse.getWorldPosition() cannot invert the current transform", tok.Line, tok.Column)
	}

	// The transform maps the game's coordinates onto the renderer's pixels, so
	// inverting it has to start from the pointer in those same pixels.
	x, y := inverse.Apply(engine.Lumen.MousePixels())

	return list(x, y)
}

// mouseSetRelativeModeMethod hides the pointer and reports only movement deltas,
// which is what a game wants while the player is dragging a camera around.
func mouseSetRelativeModeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("mouse.setRelativeMode", tok, args, 1); err != nil {
		return err
	}

	enabled, err := boolean("mouse.setRelativeMode", tok, args, 0)

	if err != nil {
		return err
	}

	sdl.SetRelativeMouseMode(enabled)

	return value.NULL
}

func mouseSetGrabbedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("mouse.setGrabbed", tok, args, 1); err != nil {
		return err
	}

	grabbed, err := boolean("mouse.setGrabbed", tok, args, 0)

	if err != nil {
		return err
	}

	engine.Lumen.Window.SetGrab(grabbed)

	return value.NULL
}

func mouseXProperty(scope *object.Scope, tok token.Token) object.Object {
	x, _ := engine.Lumen.MousePosition()

	return object.NewFloat(x)
}

func mouseYProperty(scope *object.Scope, tok token.Token) object.Object {
	_, y := engine.Lumen.MousePosition()

	return object.NewFloat(y)
}

// mouseWheelProperty reports how far the wheel turned during this frame.
// Positive values scroll away from the player.
func mouseWheelProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewInt(int64(engine.Lumen.WheelY))
}

func mouseWheelXProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewInt(int64(engine.Lumen.WheelX))
}

// buttonState resolves a button name and applies a predicate to the current and
// previous button masks.
func buttonState(name string, tok token.Token, args []object.Object, predicate func(current, previous, mask uint32) bool) object.Object {
	if err := arity(name, tok, args, 1); err != nil {
		return err
	}

	button, err := text(name, tok, args, 0)

	if err != nil {
		return err
	}

	mask, ok := engine.MouseButtonMask(button)

	if !ok {
		return object.NewError("%d:%d: runtime error: %s() does not recognise the button '%s'", tok.Line, tok.Column, name, button)
	}

	return &object.Boolean{Value: predicate(engine.Lumen.CurrentMouseState, engine.Lumen.PreviousMouseState, mask)}
}
