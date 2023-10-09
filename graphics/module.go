package graphics

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/renderer"
	"github.com/veandco/go-sdl2/sdl"
)

var GraphicsMethods = map[string]*object.LibraryFunction{}
var GraphicsProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(GraphicsMethods, "scale", graphicsScale)
	modules.RegisterMethod(GraphicsMethods, "setColor", graphicsSetColor)
	modules.RegisterMethod(GraphicsMethods, "rectangle", graphicsRectangle)
	modules.RegisterMethod(GraphicsMethods, "filledRectangle", graphicsFilledRectangle)
	modules.RegisterMethod(GraphicsMethods, "line", graphicsLine)
	modules.RegisterMethod(GraphicsMethods, "clear", graphicsClear)
	modules.RegisterMethod(GraphicsMethods, "pixel", graphicsPixel)
}

func graphicsScale(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 2 {
		panic("wrong number of arguments. expected=2")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())

	renderer.Renderer.SetScale(float32(x), float32(y))

	return nil
}

func graphicsSetColor(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 3 {
		panic("wrong number of arguments. expected=3")
		// return object.NewError("wrong number of arguments. got=%d, expected=4", len(args))
	}

	red := uint8(args[0].(*object.Number).Value.IntPart())
	green := uint8(args[1].(*object.Number).Value.IntPart())
	blue := uint8(args[2].(*object.Number).Value.IntPart())
	alpha := uint8(255)

	if len(args) == 4 {
		alpha = uint8(args[3].(*object.Number).Value.IntPart())
	}

	renderer.Renderer.SetDrawColor(red, green, blue, alpha)

	return nil
}

func graphicsRectangle(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 4 {
		panic("wrong number of arguments. expected=4")
		// return object.NewError("wrong number of arguments. got=%d, expected=4", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	w := int32(args[2].(*object.Number).Value.IntPart())
	h := int32(args[3].(*object.Number).Value.IntPart())

	rectangle := sdl.Rect{X: x, Y: y, W: w, H: h}

	renderer.Renderer.DrawRect(&rectangle)

	return nil
}

func graphicsFilledRectangle(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 4 {
		panic("wrong number of arguments. expected=4")
		// return object.NewError("wrong number of arguments. got=%d, expected=4", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	w := int32(args[2].(*object.Number).Value.IntPart())
	h := int32(args[3].(*object.Number).Value.IntPart())

	rectangle := sdl.Rect{X: x, Y: y, W: w, H: h}

	renderer.Renderer.FillRect(&rectangle)

	return nil
}

func graphicsLine(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 4 {
		panic("wrong number of arguments. expected=4")
		// return object.NewError("wrong number of arguments. got=%d, expected=4", len(args))
	}

	x1 := int32(args[0].(*object.Number).Value.IntPart())
	y1 := int32(args[1].(*object.Number).Value.IntPart())
	x2 := int32(args[2].(*object.Number).Value.IntPart())
	y2 := int32(args[3].(*object.Number).Value.IntPart())

	renderer.Renderer.DrawLine(x1, y1, x2, y2)

	return nil
}

func graphicsClear(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	renderer.Renderer.Clear()

	return nil
}

func graphicsPixel(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 2 {
		panic("wrong number of arguments. expected=2")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())

	renderer.Renderer.DrawPoint(x, y)

	return nil
}
