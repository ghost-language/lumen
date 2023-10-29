package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
	"github.com/shopspring/decimal"
)

var WindowMethods = map[string]*object.LibraryFunction{}
var WindowProperties = map[string]*object.LibraryProperty{}

func init() {
	// Methods
	modules.RegisterMethod(WindowMethods, "setTitle", windowTitleMethod)

	// Properties
	modules.RegisterProperty(WindowProperties, "fps", windowFpsProperty)
}

func windowTitleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	title, ok := args[0].(*object.String)

	if !ok {
		return object.NewError("argument to `title` must be a string, got %s", args[0].Type())
	}

	engine.Lumen.Title = title.Value

	return nil
}

func windowFpsProperty(scope *object.Scope, tok token.Token) object.Object {
	return &object.Number{Value: decimal.NewFromInt(int64(engine.Lumen.CurrentFps))}
}
