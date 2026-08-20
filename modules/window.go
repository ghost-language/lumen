package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
	"github.com/veandco/go-sdl2/sdl"
)

var WindowMethods = map[string]*object.LibraryFunction{}
var WindowProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(WindowMethods, "setTitle", windowSetTitleMethod)
	modules.RegisterMethod(WindowMethods, "setMode", windowSetModeMethod)
	modules.RegisterMethod(WindowMethods, "setSize", windowSetSizeMethod)
	modules.RegisterMethod(WindowMethods, "setFullscreen", windowSetFullscreenMethod)
	modules.RegisterMethod(WindowMethods, "toggleFullscreen", windowToggleFullscreenMethod)
	modules.RegisterMethod(WindowMethods, "setResizable", windowSetResizableMethod)
	modules.RegisterMethod(WindowMethods, "setBorderless", windowSetBorderlessMethod)
	modules.RegisterMethod(WindowMethods, "setVsync", windowSetVsyncMethod)
	modules.RegisterMethod(WindowMethods, "setIcon", windowSetIconMethod)
	modules.RegisterMethod(WindowMethods, "setPosition", windowSetPositionMethod)
	modules.RegisterMethod(WindowMethods, "center", windowCenterMethod)
	modules.RegisterMethod(WindowMethods, "maximize", windowMaximizeMethod)
	modules.RegisterMethod(WindowMethods, "minimize", windowMinimizeMethod)
	modules.RegisterMethod(WindowMethods, "restore", windowRestoreMethod)
	modules.RegisterMethod(WindowMethods, "getDesktopDimensions", windowGetDesktopDimensionsMethod)
	modules.RegisterMethod(WindowMethods, "getDimensions", windowGetDimensionsMethod)

	modules.RegisterProperty(WindowProperties, "fps", windowFpsProperty)
	modules.RegisterProperty(WindowProperties, "width", windowWidthProperty)
	modules.RegisterProperty(WindowProperties, "height", windowHeightProperty)
	modules.RegisterProperty(WindowProperties, "title", windowTitleProperty)
	modules.RegisterProperty(WindowProperties, "fullscreen", windowFullscreenProperty)
	modules.RegisterProperty(WindowProperties, "focused", windowFocusedProperty)
}

func windowSetTitleMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("window.setTitle", tok, args, 1); err != nil {
		return err
	}

	title, err := text("window.setTitle", tok, args, 0)

	if err != nil {
		return err
	}

	engine.Lumen.Title = title
	engine.Lumen.Window.SetTitle(title)

	return value.NULL
}

// windowSetModeMethod resizes the window and optionally goes fullscreen:
// window.setMode(width, height) or window.setMode(width, height, fullscreen).
func windowSetModeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("window.setMode", tok, args, 2, 3); err != nil {
		return err
	}

	width, err := integer("window.setMode", tok, args, 0)

	if err != nil {
		return err
	}

	height, err := integer("window.setMode", tok, args, 1)

	if err != nil {
		return err
	}

	fullscreen := engine.Lumen.IsFullscreen()

	if len(args) == 3 {
		given, err := boolean("window.setMode", tok, args, 2)

		if err != nil {
			return err
		}

		fullscreen = given
	}

	if modeErr := engine.Lumen.SetMode(int32(width), int32(height), fullscreen); modeErr != nil {
		return object.NewError("%d:%d: runtime error: window.setMode() %s", tok.Line, tok.Column, modeErr)
	}

	return value.NULL
}

func windowSetSizeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("window.setSize", tok, args, 2); err != nil {
		return err
	}

	return windowSetModeMethod(scope, tok, args...)
}

func windowSetFullscreenMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("window.setFullscreen", tok, args, 1); err != nil {
		return err
	}

	fullscreen, err := boolean("window.setFullscreen", tok, args, 0)

	if err != nil {
		return err
	}

	if modeErr := engine.Lumen.SetMode(engine.Lumen.Width, engine.Lumen.Height, fullscreen); modeErr != nil {
		return object.NewError("%d:%d: runtime error: window.setFullscreen() %s", tok.Line, tok.Column, modeErr)
	}

	return value.NULL
}

func windowToggleFullscreenMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	fullscreen := !engine.Lumen.IsFullscreen()

	if err := engine.Lumen.SetMode(engine.Lumen.Width, engine.Lumen.Height, fullscreen); err != nil {
		return object.NewError("%d:%d: runtime error: window.toggleFullscreen() %s", tok.Line, tok.Column, err)
	}

	return &object.Boolean{Value: fullscreen}
}

func windowSetResizableMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("window.setResizable", tok, args, 1); err != nil {
		return err
	}

	resizable, err := boolean("window.setResizable", tok, args, 0)

	if err != nil {
		return err
	}

	engine.Lumen.Window.SetResizable(resizable)

	return value.NULL
}

func windowSetBorderlessMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("window.setBorderless", tok, args, 1); err != nil {
		return err
	}

	borderless, err := boolean("window.setBorderless", tok, args, 0)

	if err != nil {
		return err
	}

	engine.Lumen.Window.SetBordered(!borderless)

	return value.NULL
}

// windowSetVsyncMethod turns vertical sync on or off. With it on, frames are
// paced by the display; with it off, the target frame rate does the pacing.
func windowSetVsyncMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("window.setVsync", tok, args, 1); err != nil {
		return err
	}

	enabled, err := boolean("window.setVsync", tok, args, 0)

	if err != nil {
		return err
	}

	if vsyncErr := engine.Lumen.Renderer.RenderSetVSync(enabled); vsyncErr != nil {
		return object.NewError("%d:%d: runtime error: window.setVsync() %s", tok.Line, tok.Column, vsyncErr)
	}

	return value.NULL
}

// windowSetIconMethod sets the window's icon from a loaded image.
func windowSetIconMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("window.setIcon", tok, args, 1); err != nil {
		return err
	}

	image, ok := args[0].(*engine.Image)

	if !ok {
		return object.NewError("%d:%d: runtime error: window.setIcon() expects an image. got=%s", tok.Line, tok.Column, args[0].Type())
	}

	if image.Surface == nil {
		return object.NewError("%d:%d: runtime error: window.setIcon() needs an image loaded from a file", tok.Line, tok.Column)
	}

	engine.Lumen.Window.SetIcon(image.Surface)

	return value.NULL
}

func windowSetPositionMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("window.setPosition", tok, args, 2); err != nil {
		return err
	}

	values, err := numbers("window.setPosition", tok, args)

	if err != nil {
		return err
	}

	engine.Lumen.Window.SetPosition(int32(values[0]), int32(values[1]))

	return value.NULL
}

func windowCenterMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.Window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)

	return value.NULL
}

func windowMaximizeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.Window.Maximize()

	return value.NULL
}

func windowMinimizeMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.Window.Minimize()

	return value.NULL
}

func windowRestoreMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	engine.Lumen.Window.Restore()

	return value.NULL
}

// windowGetDesktopDimensionsMethod reports the size of the display the window is
// on, which a game needs to pick a sensible fullscreen or startup size.
func windowGetDesktopDimensionsMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	index, err := engine.Lumen.Window.GetDisplayIndex()

	if err != nil {
		return object.NewError("%d:%d: runtime error: window.getDesktopDimensions() %s", tok.Line, tok.Column, err)
	}

	mode, err := sdl.GetDesktopDisplayMode(index)

	if err != nil {
		return object.NewError("%d:%d: runtime error: window.getDesktopDimensions() %s", tok.Line, tok.Column, err)
	}

	return integerList(int64(mode.W), int64(mode.H))
}

func windowGetDimensionsMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	width, height := engine.Lumen.Window.GetSize()

	return integerList(int64(width), int64(height))
}

func windowFpsProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewInt(int64(engine.Lumen.CurrentFps))
}

func windowWidthProperty(scope *object.Scope, tok token.Token) object.Object {
	width, _ := engine.Lumen.Window.GetSize()

	return object.NewInt(int64(width))
}

func windowHeightProperty(scope *object.Scope, tok token.Token) object.Object {
	_, height := engine.Lumen.Window.GetSize()

	return object.NewInt(int64(height))
}

func windowTitleProperty(scope *object.Scope, tok token.Token) object.Object {
	return &object.String{Value: engine.Lumen.Title}
}

func windowFullscreenProperty(scope *object.Scope, tok token.Token) object.Object {
	return &object.Boolean{Value: engine.Lumen.IsFullscreen()}
}

func windowFocusedProperty(scope *object.Scope, tok token.Token) object.Object {
	return &object.Boolean{Value: engine.Lumen.HasFocus}
}
