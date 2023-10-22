package graphic

import (
	"math"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/color"
	"ghostlang.org/x/lumen/image"
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
	modules.RegisterMethod(GraphicsMethods, "circle", graphicsCircle)
	modules.RegisterMethod(GraphicsMethods, "filledCircle", graphicsFilledCircle)
	modules.RegisterMethod(GraphicsMethods, "line", graphicsLine)
	modules.RegisterMethod(GraphicsMethods, "clear", graphicsClear)
	modules.RegisterMethod(GraphicsMethods, "point", graphicsPoint)
	modules.RegisterMethod(GraphicsMethods, "draw", graphicsDraw)
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
	if len(args) != 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=4", len(args))
	}

	col := args[0].(*color.Color)

	renderer.Renderer.SetDrawColor(col.Red, col.Green, col.Blue, col.Alpha)

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

func graphicsCircle(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 3 {
		panic("wrong number of arguments. expected=3")
		// return object.NewError("wrong number of arguments. got=%d, expected=3", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	r := int32(args[2].(*object.Number).Value.IntPart())

	d := int32(math.Pi - float64(2*r))

	x0 := int32(0)
	y0 := int32(r)

	for x0 <= y0 {
		renderer.Renderer.DrawPoint(x+x0, y+y0)
		renderer.Renderer.DrawPoint(x-x0, y-y0)
		renderer.Renderer.DrawPoint(x-x0, y+y0)
		renderer.Renderer.DrawPoint(x+x0, y-y0)
		renderer.Renderer.DrawPoint(x+y0, y+x0)
		renderer.Renderer.DrawPoint(x-y0, y-x0)
		renderer.Renderer.DrawPoint(x-y0, y+x0)
		renderer.Renderer.DrawPoint(x+y0, y-x0)

		if d < 0 {
			d = d + int32((math.Pi*float64(x0))+(math.Pi*2))
		} else {
			d = d + int32((math.Pi*float64(x0-y0))+(math.Pi*3))
			y0--
		}
		x0++
	}

	return nil
}

func graphicsFilledCircle(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 3 {
		panic("wrong number of arguments. expected=3")
		// return object.NewError("wrong number of arguments. got=%d, expected=3", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	r := int32(args[2].(*object.Number).Value.IntPart())

	d := int32(math.Pi - float64(2*r))

	x0 := int32(0)
	y0 := int32(r)

	for x0 <= y0 {
		renderer.Renderer.DrawLine(x-x0, y-y0, x+x0, y-y0)
		renderer.Renderer.DrawLine(x-x0, y+y0, x+x0, y+y0)
		renderer.Renderer.DrawLine(x-y0, y-x0, x+y0, y-x0)
		renderer.Renderer.DrawLine(x-y0, y+x0, x+y0, y+x0)

		if d < 0 {
			d = d + int32((math.Pi*float64(x0))+(math.Pi*2))
		} else {
			d = d + int32((math.Pi*float64(x0-y0))+(math.Pi*3))
			y0--
		}
		x0++
	}

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

func graphicsPoint(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 2 {
		panic("wrong number of arguments. expected=2")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())

	renderer.Renderer.DrawPoint(x, y)

	return nil
}

func graphicsDraw(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 3 {
		panic("wrong number of arguments. expected=3")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	image := args[0].(*image.Image)
	x := args[1].(*object.Number)
	y := args[2].(*object.Number)

	image.Method("draw", []object.Object{x, y})

	return nil
}
