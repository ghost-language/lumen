package window

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
)

var WindowMethods = map[string]*object.LibraryFunction{}
var WindowProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(WindowMethods, "title", windowTitleMethod)

	modules.RegisterProperty(WindowProperties, "fps", windowFpsProperty)
}

func windowTitleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return nil
}

func windowFpsProperty(scope *object.Scope, tok token.Token) object.Object {
	return nil
}
