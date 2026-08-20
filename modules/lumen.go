package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
)

var LumenMethods = map[string]*object.LibraryFunction{}
var LumenProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(LumenMethods, "quit", lumenQuitMethod)
	modules.RegisterMethod(LumenMethods, "setTargetFps", lumenSetTargetFpsMethod)
	modules.RegisterMethod(LumenMethods, "getTargetFps", lumenGetTargetFpsMethod)

	modules.RegisterProperty(LumenProperties, "version", lumenVersionProperty)
}

// lumenQuitMethod ends the game. The loop finishes the frame it is in before
// shutting down, so a game never stops halfway through drawing.
func lumenQuitMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.Quit()

	return value.NULL
}

// lumenSetTargetFpsMethod caps the frame rate. Zero removes the cap and lets
// the game run as fast as vsync allows.
func lumenSetTargetFpsMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("lumen.setTargetFps", tok, args, 1); err != nil {
		return err
	}

	fps, err := integer("lumen.setTargetFps", tok, args, 0)

	if err != nil {
		return err
	}

	if fps < 0 {
		return object.NewError("%d:%d: runtime error: lumen.setTargetFps() expects a positive number, or 0 for uncapped. got=%d", tok.Line, tok.Column, fps)
	}

	engine.Lumen.TargetFps = uint64(fps)

	return value.NULL
}

func lumenGetTargetFpsMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return object.NewInt(int64(engine.Lumen.TargetFps))
}

func lumenVersionProperty(scope *object.Scope, tok token.Token) object.Object {
	return &object.String{Value: engine.Version}
}
