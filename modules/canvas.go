package modules

import (
	"math"

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
	modules.RegisterMethod(CanvasMethods, "filledRectangle", canvasFilledRectangleMethod)
	modules.RegisterMethod(CanvasMethods, "circle", canvasCircleMethod)
	modules.RegisterMethod(CanvasMethods, "filledCircle", canvasFilledCircleMethod)
	modules.RegisterMethod(CanvasMethods, "line", canvasLineMethod)
	modules.RegisterMethod(CanvasMethods, "point", canvasPointMethod)
	modules.RegisterMethod(CanvasMethods, "clear", canvasClearMethod)
	modules.RegisterMethod(CanvasMethods, "setColor", canvasSetColorMethod)
	modules.RegisterMethod(CanvasMethods, "setFont", canvasSetFontMethod)
	modules.RegisterMethod(CanvasMethods, "resetFont", canvasResetFontMethod)
	modules.RegisterMethod(CanvasMethods, "print", canvasPrintMethod)
	modules.RegisterMethod(CanvasMethods, "scale", canvasScaleMethod)
	modules.RegisterMethod(CanvasMethods, "translate", canvasTranslateMethod)

	// Properties
	// modules.RegisterProperty(CanvasProperties, "fps", windowFpsProperty)
}

func canvasRectangleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 4 {
		return object.NewError("wrong number of arguments. got=%d, want=4", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	w := int32(args[2].(*object.Number).Value.IntPart())
	h := int32(args[3].(*object.Number).Value.IntPart())

	rectangle := sdl.Rect{X: x, Y: y, W: w, H: h}

	engine.Lumen.Renderer.DrawRect(&rectangle)

	return nil
}

func canvasFilledRectangleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 4 {
		return object.NewError("wrong number of arguments. got=%d, want=4", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	w := int32(args[2].(*object.Number).Value.IntPart())
	h := int32(args[3].(*object.Number).Value.IntPart())

	x, y = engine.Lumen.ApplyOffset(x, y)

	rectangle := sdl.Rect{X: x, Y: y, W: w, H: h}

	engine.Lumen.Renderer.FillRect(&rectangle)

	return nil
}

func canvasCircleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError("wrong number of arguments. got=%d, want=3", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	r := int32(args[2].(*object.Number).Value.IntPart())

	d := int32(math.Pi - float64(2*r))

	x0 := int32(0)
	y0 := int32(r)

	x, y = engine.Lumen.ApplyOffset(x, y)
	x0, y0 = engine.Lumen.ApplyOffset(x0, y0)

	for x0 <= y0 {
		engine.Lumen.Renderer.DrawPoint(x+x0, y+y0)
		engine.Lumen.Renderer.DrawPoint(x-x0, y-y0)
		engine.Lumen.Renderer.DrawPoint(x-x0, y+y0)
		engine.Lumen.Renderer.DrawPoint(x+x0, y-y0)
		engine.Lumen.Renderer.DrawPoint(x+y0, y+x0)
		engine.Lumen.Renderer.DrawPoint(x-y0, y-x0)
		engine.Lumen.Renderer.DrawPoint(x-y0, y+x0)
		engine.Lumen.Renderer.DrawPoint(x+y0, y-x0)

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

func canvasFilledCircleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError("wrong number of arguments. got=%d, want=3", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())
	r := int32(args[2].(*object.Number).Value.IntPart())

	d := int32(math.Pi - float64(2*r))

	x0 := int32(0)
	y0 := int32(r)

	x, y = engine.Lumen.ApplyOffset(x, y)
	x0, y0 = engine.Lumen.ApplyOffset(x0, y0)

	for x0 <= y0 {
		engine.Lumen.Renderer.DrawLine(x-x0, y-y0, x+x0, y-y0)
		engine.Lumen.Renderer.DrawLine(x-x0, y+y0, x+x0, y+y0)
		engine.Lumen.Renderer.DrawLine(x-y0, y-x0, x+y0, y-x0)
		engine.Lumen.Renderer.DrawLine(x-y0, y+x0, x+y0, y+x0)

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

func canvasLineMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 4 {
		return object.NewError("wrong number of arguments. got=%d, want=4", len(args))
	}

	x1 := int32(args[0].(*object.Number).Value.IntPart())
	y1 := int32(args[1].(*object.Number).Value.IntPart())
	x2 := int32(args[2].(*object.Number).Value.IntPart())
	y2 := int32(args[3].(*object.Number).Value.IntPart())

	x1, y1 = engine.Lumen.ApplyOffset(x1, y1)
	x2, y2 = engine.Lumen.ApplyOffset(x2, y2)

	engine.Lumen.Renderer.DrawLine(x1, y1, x2, y2)

	return nil
}

func canvasPointMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 2 {
		return object.NewError("wrong number of arguments. got=%d, want=2", len(args))
	}

	x := int32(args[0].(*object.Number).Value.IntPart())
	y := int32(args[1].(*object.Number).Value.IntPart())

	x, y = engine.Lumen.ApplyOffset(x, y)

	engine.Lumen.Renderer.DrawPoint(x, y)

	return nil
}

func canvasClearMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.Renderer.Clear()

	return nil
}

func canvasSetColorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	color := args[0].(*engine.Color)

	// Set the color
	engine.Lumen.Renderer.SetDrawColor(color.Red, color.Green, color.Blue, color.Alpha)

	return nil
}

func canvasSetFontMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	font := args[0].(*engine.Font)

	// Set the font
	engine.Lumen.CurrentFont = font

	return nil
}

func canvasResetFontMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 0 {
		return object.NewError("wrong number of arguments. got=%d, want=0", len(args))
	}

	// Set the font
	engine.Lumen.CurrentFont = engine.Lumen.DefaultFont

	return nil
}

func canvasPrintMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 3 {
		return object.NewError("wrong number of arguments. got=%d, want=3", len(args))
	}

	engine.Lumen.CurrentFont.Method("print", args)

	return nil
}

func canvasScaleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	var x, y float32

	if len(args) == 2 {
		x, _ = args[0].(*object.Number).Value.BigFloat().Float32()
		y, _ = args[1].(*object.Number).Value.BigFloat().Float32()
	} else {
		x, _ = args[0].(*object.Number).Value.BigFloat().Float32()
		y, _ = args[0].(*object.Number).Value.BigFloat().Float32()
	}

	engine.Lumen.Renderer.SetScale(x, y)

	return nil
}

func canvasTranslateMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	var x, y float32

	if len(args) == 2 {
		x, _ = args[0].(*object.Number).Value.BigFloat().Float32()
		y, _ = args[1].(*object.Number).Value.BigFloat().Float32()
	} else {
		x, _ = args[0].(*object.Number).Value.BigFloat().Float32()
		y, _ = args[0].(*object.Number).Value.BigFloat().Float32()
	}

	engine.Lumen.OffsetX = int32(x)
	engine.Lumen.OffsetY = int32(y)

	return nil
}

// func windowFpsProperty(scope *object.Scope, tok token.Token) object.Object {
// 	return &object.Number{Value: decimal.NewFromInt(int64(engine.Lumen.CurrentFps))}
// }
