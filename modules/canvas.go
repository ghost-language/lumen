package modules

import (
	"math"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
)

var CanvasMethods = map[string]*object.LibraryFunction{}
var CanvasProperties = map[string]*object.LibraryProperty{}

func init() {
	// Shapes
	modules.RegisterMethod(CanvasMethods, "rectangle", canvasRectangleMethod)
	modules.RegisterMethod(CanvasMethods, "filledRectangle", canvasFilledRectangleMethod)
	modules.RegisterMethod(CanvasMethods, "circle", canvasCircleMethod)
	modules.RegisterMethod(CanvasMethods, "filledCircle", canvasFilledCircleMethod)
	modules.RegisterMethod(CanvasMethods, "ellipse", canvasEllipseMethod)
	modules.RegisterMethod(CanvasMethods, "filledEllipse", canvasFilledEllipseMethod)
	modules.RegisterMethod(CanvasMethods, "arc", canvasArcMethod)
	modules.RegisterMethod(CanvasMethods, "filledArc", canvasFilledArcMethod)
	modules.RegisterMethod(CanvasMethods, "polygon", canvasPolygonMethod)
	modules.RegisterMethod(CanvasMethods, "filledPolygon", canvasFilledPolygonMethod)
	modules.RegisterMethod(CanvasMethods, "line", canvasLineMethod)
	modules.RegisterMethod(CanvasMethods, "point", canvasPointMethod)

	// State
	modules.RegisterMethod(CanvasMethods, "clear", canvasClearMethod)
	modules.RegisterMethod(CanvasMethods, "setColor", canvasSetColorMethod)
	modules.RegisterMethod(CanvasMethods, "getColor", canvasGetColorMethod)
	modules.RegisterMethod(CanvasMethods, "setBackgroundColor", canvasSetBackgroundColorMethod)
	modules.RegisterMethod(CanvasMethods, "setLineWidth", canvasSetLineWidthMethod)
	modules.RegisterMethod(CanvasMethods, "getLineWidth", canvasGetLineWidthMethod)
	modules.RegisterMethod(CanvasMethods, "setPointSize", canvasSetPointSizeMethod)
	modules.RegisterMethod(CanvasMethods, "setBlendMode", canvasSetBlendModeMethod)
	modules.RegisterMethod(CanvasMethods, "setScissor", canvasSetScissorMethod)
	modules.RegisterMethod(CanvasMethods, "clearScissor", canvasClearScissorMethod)

	// Text
	modules.RegisterMethod(CanvasMethods, "setFont", canvasSetFontMethod)
	modules.RegisterMethod(CanvasMethods, "getFont", canvasGetFontMethod)
	modules.RegisterMethod(CanvasMethods, "resetFont", canvasResetFontMethod)
	modules.RegisterMethod(CanvasMethods, "print", canvasPrintMethod)
	modules.RegisterMethod(CanvasMethods, "printf", canvasPrintfMethod)

	// Transform stack
	modules.RegisterMethod(CanvasMethods, "push", canvasPushMethod)
	modules.RegisterMethod(CanvasMethods, "pop", canvasPopMethod)
	modules.RegisterMethod(CanvasMethods, "origin", canvasOriginMethod)
	modules.RegisterMethod(CanvasMethods, "scale", canvasScaleMethod)
	modules.RegisterMethod(CanvasMethods, "translate", canvasTranslateMethod)
	modules.RegisterMethod(CanvasMethods, "rotate", canvasRotateMethod)
	modules.RegisterMethod(CanvasMethods, "shear", canvasShearMethod)
	modules.RegisterMethod(CanvasMethods, "toScreen", canvasToScreenMethod)
	modules.RegisterMethod(CanvasMethods, "toWorld", canvasToWorldMethod)

	// Render targets
	modules.RegisterMethod(CanvasMethods, "newTarget", canvasNewTargetMethod)
	modules.RegisterMethod(CanvasMethods, "setTarget", canvasSetTargetMethod)
	modules.RegisterMethod(CanvasMethods, "newQuad", canvasNewQuadMethod)
	modules.RegisterMethod(CanvasMethods, "screenshot", canvasScreenshotMethod)

	// Properties
	modules.RegisterProperty(CanvasProperties, "width", canvasWidthProperty)
	modules.RegisterProperty(CanvasProperties, "height", canvasHeightProperty)
}

// =============================================================================
// Shapes

func canvasRectangleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return drawRectangle("canvas.rectangle", tok, args, false)
}

func canvasFilledRectangleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return drawRectangle("canvas.filledRectangle", tok, args, true)
}

// drawRectangle builds the rectangle as four corner points rather than handing
// SDL a rect, so a rotated or sheared transform draws a rotated rectangle
// instead of an axis-aligned approximation of one.
func drawRectangle(name string, tok token.Token, args []object.Object, filled bool) object.Object {
	if err := arity(name, tok, args, 4); err != nil {
		return err
	}

	values, err := numbers(name, tok, args)

	if err != nil {
		return err
	}

	x, y, width, height := values[0], values[1], values[2], values[3]

	points := []float64{x, y, x + width, y, x + width, y + height, x, y + height}

	if filled {
		engine.Lumen.FillPolygon(points)
	} else {
		engine.Lumen.DrawPolyline(points, true)
	}

	return value.NULL
}

func canvasCircleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return drawCircle("canvas.circle", tok, args, false)
}

func canvasFilledCircleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return drawCircle("canvas.filledCircle", tok, args, true)
}

func drawCircle(name string, tok token.Token, args []object.Object, filled bool) object.Object {
	if err := arityRange(name, tok, args, 3, 4); err != nil {
		return err
	}

	values, err := numbers(name, tok, args)

	if err != nil {
		return err
	}

	segments := 0

	if len(values) == 4 {
		segments = int(values[3])
	}

	points := engine.Lumen.EllipsePoints(values[0], values[1], values[2], values[2], segments)

	if filled {
		engine.Lumen.FillPolygon(points)
	} else {
		engine.Lumen.DrawPolyline(points, true)
	}

	return value.NULL
}

func canvasEllipseMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return drawEllipse("canvas.ellipse", tok, args, false)
}

func canvasFilledEllipseMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return drawEllipse("canvas.filledEllipse", tok, args, true)
}

func drawEllipse(name string, tok token.Token, args []object.Object, filled bool) object.Object {
	if err := arityRange(name, tok, args, 4, 5); err != nil {
		return err
	}

	values, err := numbers(name, tok, args)

	if err != nil {
		return err
	}

	segments := 0

	if len(values) == 5 {
		segments = int(values[4])
	}

	points := engine.Lumen.EllipsePoints(values[0], values[1], values[2], values[3], segments)

	if filled {
		engine.Lumen.FillPolygon(points)
	} else {
		engine.Lumen.DrawPolyline(points, true)
	}

	return value.NULL
}

func canvasArcMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return drawArc("canvas.arc", tok, args, false)
}

func canvasFilledArcMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return drawArc("canvas.filledArc", tok, args, true)
}

func drawArc(name string, tok token.Token, args []object.Object, filled bool) object.Object {
	if err := arityRange(name, tok, args, 5, 6); err != nil {
		return err
	}

	values, err := numbers(name, tok, args)

	if err != nil {
		return err
	}

	segments := 0

	if len(values) == 6 {
		segments = int(values[5])
	}

	points := engine.Lumen.ArcPoints(values[0], values[1], values[2], values[3], values[4], segments, true)

	if filled {
		engine.Lumen.FillPolygon(points)
	} else {
		engine.Lumen.DrawPolyline(points, true)
	}

	return value.NULL
}

func canvasPolygonMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	points, err := polygonPoints("canvas.polygon", tok, args)

	if err != nil {
		return err
	}

	engine.Lumen.DrawPolyline(points, true)

	return value.NULL
}

func canvasFilledPolygonMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	points, err := polygonPoints("canvas.filledPolygon", tok, args)

	if err != nil {
		return err
	}

	engine.Lumen.FillPolygon(points)

	return value.NULL
}

// polygonPoints accepts either a flat run of coordinates or a single list of
// them, so both canvas.polygon(x1, y1, x2, y2, x3, y3) and
// canvas.polygon([x1, y1, x2, y2, x3, y3]) read naturally.
func polygonPoints(name string, tok token.Token, args []object.Object) ([]float64, *object.Error) {
	if len(args) == 1 {
		elements, ok := args[0].(*object.List)

		if !ok {
			return nil, object.NewError("%d:%d: runtime error: %s() expects a list of coordinates or a run of numbers", tok.Line, tok.Column, name)
		}

		return numbers(name, tok, elements.Elements)
	}

	if len(args) < 6 || len(args)%2 != 0 {
		return nil, object.NewError("%d:%d: runtime error: %s() expects at least 3 x/y pairs. got=%d value(s)", tok.Line, tok.Column, name, len(args))
	}

	return numbers(name, tok, args)
}

func canvasLineMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 4 || len(args)%2 != 0 {
		return object.NewError("%d:%d: runtime error: canvas.line() expects pairs of x/y coordinates. got=%d value(s)", tok.Line, tok.Column, len(args))
	}

	points, err := numbers("canvas.line", tok, args)

	if err != nil {
		return err
	}

	engine.Lumen.DrawPolyline(points, false)

	return value.NULL
}

func canvasPointMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 2 || len(args)%2 != 0 {
		return object.NewError("%d:%d: runtime error: canvas.point() expects pairs of x/y coordinates. got=%d value(s)", tok.Line, tok.Column, len(args))
	}

	points, err := numbers("canvas.point", tok, args)

	if err != nil {
		return err
	}

	engine.Lumen.DrawPoints(points)

	return value.NULL
}

// =============================================================================
// State

func canvasClearMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	color := engine.Lumen.Graphics.BackgroundColor

	if len(args) == 1 {
		given, ok := args[0].(*engine.Color)

		if !ok {
			return object.NewError("%d:%d: runtime error: canvas.clear() expects a color. got=%s", tok.Line, tok.Column, args[0].Type())
		}

		color = given
	}

	engine.Lumen.Renderer.SetDrawColor(color.Red, color.Green, color.Blue, color.Alpha)
	engine.Lumen.Renderer.Clear()

	return value.NULL
}

func canvasSetColorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	color, err := colorArgument("canvas.setColor", tok, args)

	if err != nil {
		return err
	}

	engine.Lumen.Graphics.Color = color

	return value.NULL
}

func canvasGetColorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return engine.Lumen.Graphics.Color
}

func canvasSetBackgroundColorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	color, err := colorArgument("canvas.setBackgroundColor", tok, args)

	if err != nil {
		return err
	}

	engine.Lumen.Graphics.BackgroundColor = color

	return value.NULL
}

// colorArgument accepts either a color object or loose r/g/b/a components, so
// canvas.setColor(color.white) and canvas.setColor(255, 255, 255) both work.
func colorArgument(name string, tok token.Token, args []object.Object) (*engine.Color, *object.Error) {
	if len(args) == 1 {
		color, ok := args[0].(*engine.Color)

		if !ok {
			return nil, object.NewError("%d:%d: runtime error: %s() expects a color. got=%s", tok.Line, tok.Column, name, args[0].Type())
		}

		return color, nil
	}

	if len(args) != 3 && len(args) != 4 {
		return nil, object.NewError("%d:%d: runtime error: %s() expects a color, or 3 to 4 components. got=%d", tok.Line, tok.Column, name, len(args))
	}

	components := make([]uint8, 0, 4)

	for index := range args {
		component, ok := args[index].(*object.Number)

		if !ok {
			return nil, object.NewError("%d:%d: runtime error: %s() expects number components. argument %d is %s", tok.Line, tok.Column, name, index+1, args[index].Type())
		}

		if index == 3 {
			components = append(components, engine.ColorAlpha(component))

			continue
		}

		components = append(components, engine.ColorComponent(component))
	}

	alpha := uint8(255)

	if len(components) == 4 {
		alpha = components[3]
	}

	return engine.NewColor(components[0], components[1], components[2], alpha), nil
}

func canvasSetLineWidthMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.setLineWidth", tok, args, 1); err != nil {
		return err
	}

	width, err := number("canvas.setLineWidth", tok, args, 0)

	if err != nil {
		return err
	}

	engine.Lumen.Graphics.LineWidth = math.Max(width, 0)

	return value.NULL
}

func canvasGetLineWidthMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return object.NewFloat(engine.Lumen.Graphics.LineWidth)
}

func canvasSetPointSizeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.setPointSize", tok, args, 1); err != nil {
		return err
	}

	size, err := number("canvas.setPointSize", tok, args, 0)

	if err != nil {
		return err
	}

	engine.Lumen.Graphics.PointSize = math.Max(size, 0)

	return value.NULL
}

func canvasSetBlendModeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.setBlendMode", tok, args, 1); err != nil {
		return err
	}

	name, err := text("canvas.setBlendMode", tok, args, 0)

	if err != nil {
		return err
	}

	mode, ok := engine.BlendModeFromName(name)

	if !ok {
		return object.NewError("%d:%d: runtime error: canvas.setBlendMode() expects 'alpha', 'add', 'multiply', or 'none'. got=%s", tok.Line, tok.Column, name)
	}

	engine.Lumen.Graphics.BlendMode = mode

	return value.NULL
}

func canvasSetScissorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.setScissor", tok, args, 4); err != nil {
		return err
	}

	values, err := numbers("canvas.setScissor", tok, args)

	if err != nil {
		return err
	}

	engine.Lumen.SetScissor(values[0], values[1], values[2], values[3])

	return value.NULL
}

func canvasClearScissorMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.ClearScissor()

	return value.NULL
}

// =============================================================================
// Text

func canvasSetFontMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.setFont", tok, args, 1); err != nil {
		return err
	}

	font, ok := args[0].(*engine.Font)

	if !ok {
		return object.NewError("%d:%d: runtime error: canvas.setFont() expects a font. got=%s", tok.Line, tok.Column, args[0].Type())
	}

	engine.Lumen.CurrentFont = font

	return value.NULL
}

func canvasGetFontMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return engine.Lumen.CurrentFont
}

func canvasResetFontMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.CurrentFont = engine.Lumen.DefaultFont

	return value.NULL
}

func canvasPrintMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 3 {
		return object.NewError("%d:%d: runtime error: canvas.print() expects at least a string, x, and y. got=%d", tok.Line, tok.Column, len(args))
	}

	result, _ := engine.Lumen.CurrentFont.Method("print", args)

	return result
}

func canvasPrintfMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) < 4 {
		return object.NewError("%d:%d: runtime error: canvas.printf() expects at least a string, x, y, and wrap limit. got=%d", tok.Line, tok.Column, len(args))
	}

	result, _ := engine.Lumen.CurrentFont.Method("printf", args)

	return result
}

// =============================================================================
// Transform stack

func canvasPushMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	all := false

	if len(args) == 1 {
		name, err := text("canvas.push", tok, args, 0)

		if err != nil {
			return err
		}

		all = name == "all"
	}

	engine.Lumen.Graphics.Push(all)

	return value.NULL
}

func canvasPopMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.Graphics.Pop()

	return value.NULL
}

// canvasOriginMethod throws away whatever transforms have been applied and goes
// back to drawing in the game's own coordinate space, which is the base
// transform rather than the identity: a game that has fixed its logical size
// still wants its scaling after asking for the origin back.
func canvasOriginMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.Graphics.SetTransform(engine.Lumen.Graphics.Base)

	return value.NULL
}

func canvasScaleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("canvas.scale", tok, args, 1, 2); err != nil {
		return err
	}

	values, err := numbers("canvas.scale", tok, args)

	if err != nil {
		return err
	}

	x := values[0]
	y := values[0]

	if len(values) == 2 {
		y = values[1]
	}

	graphics := engine.Lumen.Graphics
	graphics.SetTransform(graphics.Transform().Scale(x, y))

	return value.NULL
}

func canvasTranslateMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.translate", tok, args, 2); err != nil {
		return err
	}

	values, err := numbers("canvas.translate", tok, args)

	if err != nil {
		return err
	}

	graphics := engine.Lumen.Graphics
	graphics.SetTransform(graphics.Transform().Translate(values[0], values[1]))

	return value.NULL
}

func canvasRotateMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.rotate", tok, args, 1); err != nil {
		return err
	}

	radians, err := number("canvas.rotate", tok, args, 0)

	if err != nil {
		return err
	}

	graphics := engine.Lumen.Graphics
	graphics.SetTransform(graphics.Transform().Rotate(radians))

	return value.NULL
}

func canvasShearMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.shear", tok, args, 2); err != nil {
		return err
	}

	values, err := numbers("canvas.shear", tok, args)

	if err != nil {
		return err
	}

	graphics := engine.Lumen.Graphics
	graphics.SetTransform(graphics.Transform().Shear(values[0], values[1]))

	return value.NULL
}

// canvasToScreenMethod maps a point from the current transform's space into
// window coordinates.
func canvasToScreenMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.toScreen", tok, args, 2); err != nil {
		return err
	}

	values, err := numbers("canvas.toScreen", tok, args)

	if err != nil {
		return err
	}

	x, y := engine.Lumen.Graphics.Transform().Apply(values[0], values[1])

	return list(x, y)
}

// canvasToWorldMethod maps a window coordinate back through the current
// transform, which is how a game turns a mouse position into a world position
// while a camera is active.
func canvasToWorldMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.toWorld", tok, args, 2); err != nil {
		return err
	}

	values, err := numbers("canvas.toWorld", tok, args)

	if err != nil {
		return err
	}

	inverse, ok := engine.Lumen.Graphics.Transform().Inverse()

	if !ok {
		return object.NewError("%d:%d: runtime error: canvas.toWorld() cannot invert the current transform", tok.Line, tok.Column)
	}

	x, y := inverse.Apply(values[0], values[1])

	return list(x, y)
}

// =============================================================================
// Render targets

func canvasNewTargetMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.newTarget", tok, args, 2); err != nil {
		return err
	}

	width, err := integer("canvas.newTarget", tok, args, 0)

	if err != nil {
		return err
	}

	height, err := integer("canvas.newTarget", tok, args, 1)

	if err != nil {
		return err
	}

	target, targetErr := engine.NewTarget(int32(width), int32(height))

	if targetErr != nil {
		return object.NewError("%d:%d: runtime error: canvas.newTarget() %s", tok.Line, tok.Column, targetErr)
	}

	return target
}

// canvasSetTargetMethod routes drawing into a target, or back to the window when
// called with no arguments.
func canvasSetTargetMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) == 0 {
		engine.Lumen.SetTarget(nil)

		return value.NULL
	}

	if err := arity("canvas.setTarget", tok, args, 1); err != nil {
		return err
	}

	if _, ok := args[0].(*object.Null); ok {
		engine.Lumen.SetTarget(nil)

		return value.NULL
	}

	target, ok := args[0].(*engine.Target)

	if !ok {
		return object.NewError("%d:%d: runtime error: canvas.setTarget() expects a target. got=%s", tok.Line, tok.Column, args[0].Type())
	}

	engine.Lumen.SetTarget(target)

	return value.NULL
}

func canvasNewQuadMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.newQuad", tok, args, 4); err != nil {
		return err
	}

	values, err := numbers("canvas.newQuad", tok, args)

	if err != nil {
		return err
	}

	return engine.NewQuad(int32(values[0]), int32(values[1]), int32(values[2]), int32(values[3]))
}

// canvasScreenshotMethod saves the current frame to a PNG in the game's save
// directory. It reads back from the renderer after the frame is drawn, so it is
// best called at the end of draw().
func canvasScreenshotMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("canvas.screenshot", tok, args, 1); err != nil {
		return err
	}

	name, err := text("canvas.screenshot", tok, args, 0)

	if err != nil {
		return err
	}

	path, pathErr := savePath(name)

	if pathErr != nil {
		return object.NewError("%d:%d: runtime error: canvas.screenshot() %s", tok.Line, tok.Column, pathErr)
	}

	if shotErr := engine.Lumen.Screenshot(path); shotErr != nil {
		return object.NewError("%d:%d: runtime error: canvas.screenshot() %s", tok.Line, tok.Column, shotErr)
	}

	return &object.String{Value: path}
}

// =============================================================================
// Properties

// canvasWidthProperty reports the width of whatever is being drawn into, which
// is the render target when one is set and the window otherwise.
func canvasWidthProperty(scope *object.Scope, tok token.Token) object.Object {
	if target := engine.Lumen.Graphics.Target; target != nil {
		return object.NewInt(int64(target.Width))
	}

	width, _, _ := engine.Lumen.Renderer.GetOutputSize()

	return object.NewInt(int64(width))
}

func canvasHeightProperty(scope *object.Scope, tok token.Token) object.Object {
	if target := engine.Lumen.Graphics.Target; target != nil {
		return object.NewInt(int64(target.Height))
	}

	_, height, _ := engine.Lumen.Renderer.GetOutputSize()

	return object.NewInt(int64(height))
}
