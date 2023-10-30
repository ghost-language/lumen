package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"github.com/shopspring/decimal"
	"github.com/veandco/go-sdl2/sdl"
)

var MouseMethods = map[string]*object.LibraryFunction{}
var MouseProperties = map[string]*object.LibraryProperty{}

func init() {
	// Methods
	modules.RegisterMethod(MouseMethods, "showCursor", mouseShowCursorMethod)
	modules.RegisterMethod(MouseMethods, "hideCursor", mouseHideCursorMethod)

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

func mouseXProperty(scope *object.Scope, tok token.Token) object.Object {
	x, _, _ := sdl.GetMouseState()

	return &object.Number{Value: decimal.NewFromInt(int64(x))}
}

func mouseYProperty(scope *object.Scope, tok token.Token) object.Object {
	_, y, _ := sdl.GetMouseState()

	return &object.Number{Value: decimal.NewFromInt(int64(y))}
}
