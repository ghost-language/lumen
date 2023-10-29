package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
	"github.com/veandco/go-sdl2/sdl"
)

var CanvasMethods = map[string]*object.LibraryFunction{}
var CanvasProperties = map[string]*object.LibraryProperty{}

func init() {
	// Methods
	modules.RegisterMethod(CanvasMethods, "rectangle", canvasRectangleMethod)

	// Properties
	// modules.RegisterProperty(CanvasProperties, "fps", windowFpsProperty)
}

func canvasRectangleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 4 {
		return object.NewError("wrong number of arguments. got=%d, want=4", len(args))
		// return object.NewError("wrong number of arguments. got=%d, expected=4", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	w := int32(args[2].(*object.Number).Value.IntPart())
	h := int32(args[3].(*object.Number).Value.IntPart())

	rectangle := sdl.Rect{X: x, Y: y, W: w, H: h}

	engine.Lumen.Renderer.DrawRect(&rectangle)

	return nil
}

// func windowFpsProperty(scope *object.Scope, tok token.Token) object.Object {
// 	return &object.Number{Value: decimal.NewFromInt(int64(engine.Lumen.CurrentFps))}
// }
