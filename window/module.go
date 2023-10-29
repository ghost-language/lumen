package window

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"github.com/shopspring/decimal"
	"github.com/veandco/go-sdl2/sdl"
)

var WindowMethods = map[string]*object.LibraryFunction{}
var WindowProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(WindowMethods, "title", windowTitle)
	modules.RegisterMethod(WindowMethods, "bordered", windowsBordered)
	modules.RegisterMethod(WindowMethods, "borderless", windowsBorderless)
	modules.RegisterMethod(WindowMethods, "fullscreen", windowsFullscreen)
	modules.RegisterMethod(WindowMethods, "resize", windowsResize)

	modules.RegisterProperty(WindowProperties, "title", windowTitleProperty)
	modules.RegisterProperty(WindowProperties, "width", windowWidthProperty)
	modules.RegisterProperty(WindowProperties, "height", windowHeightProperty)
	modules.RegisterProperty(WindowProperties, "fps", windowFpsProperty)
}

func windowTitle(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	Window.Window.SetTitle(args[0].(*object.String).Value)

	return nil
}

func windowsBordered(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	Window.Window.SetBordered(true)

	return nil
}

func windowsBorderless(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	Window.Window.SetBordered(false)

	return nil
}

func windowsFullscreen(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	Window.Window.SetFullscreen(sdl.WINDOW_FULLSCREEN)

	return nil
}

func windowsResize(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 2 {
		panic("wrong number of arguments. expected=2")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	width := args[0].(*object.Number).Value.IntPart()
	height := args[1].(*object.Number).Value.IntPart()

	Window.Window.SetSize(int32(width), int32(height))

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

func windowFpsProperty(scope *object.Scope, tok token.Token) object.Object {
	return &object.Number{Value: decimal.NewFromInt(int64(Window.Fps))}
}
