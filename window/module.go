package window

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"github.com/shopspring/decimal"
)

var WindowMethods = map[string]*object.LibraryFunction{}
var WindowProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(WindowMethods, "title", windowTitle)

	modules.RegisterProperty(WindowProperties, "title", windowTitleProperty)
	modules.RegisterProperty(WindowProperties, "width", windowWidthProperty)
	modules.RegisterProperty(WindowProperties, "height", windowHeightProperty)
}

func windowTitle(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	Window.Window.SetTitle(args[0].(*object.String).Value)

	return nil
}

func windowTitleProperty(scope *object.Scope, tok token.Token) object.Object {
	return &object.String{Value: Window.Window.GetTitle()}
}

func windowWidthProperty(scope *object.Scope, tok token.Token) object.Object {
	width, _ := Window.Window.GetSize()

	return &object.Number{Value: decimal.NewFromInt(int64(width))}
}

func windowHeightProperty(scope *object.Scope, tok token.Token) object.Object {
	_, height := Window.Window.GetSize()

	return &object.Number{Value: decimal.NewFromInt(int64(height))}
}
