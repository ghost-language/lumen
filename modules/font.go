package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
)

var FontMethods = map[string]*object.LibraryFunction{}
var FontProperties = map[string]*object.LibraryProperty{}

func init() {
	// Methods
	modules.RegisterMethod(FontMethods, "load", fontLoadMethod)

	// Properties
	// modules.RegisterProperty(FontProperties, "fps", windowFpsProperty)
}

func fontLoadMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	font := engine.NewFont(args[0].String(), int(args[1].(*object.Number).Value.IntPart()))

	return font
}

// func windowFpsProperty(scope *object.Scope, tok token.Token) object.Object {
// 	return &object.Number{Value: decimal.NewFromInt(int64(engine.Lumen.CurrentFps))}
// }
