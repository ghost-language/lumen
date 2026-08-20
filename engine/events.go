package engine

import (
	"ghostlang.org/x/ghost/object"
	"github.com/veandco/go-sdl2/sdl"
)

// handleEvents drains SDL's event queue and forwards anything a game might care
// about to its Ghost callbacks.
func (engine *Engine) handleEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			engine.handleQuitEvent()
		case *sdl.KeyboardEvent:
			engine.handleKeyboardEvent(e)
		case *sdl.TextInputEvent:
			engine.handleTextInputEvent(e)
		case *sdl.MouseButtonEvent:
			engine.handleMouseButtonEvent(e)
		case *sdl.MouseMotionEvent:
			engine.handleMouseMotionEvent(e)
		case *sdl.MouseWheelEvent:
			engine.handleMouseWheelEvent(e)
		case *sdl.WindowEvent:
			engine.handleWindowEvent(e)
		case *sdl.ControllerDeviceEvent:
			engine.handleControllerDeviceEvent(e)
		}
	}
}

// handleQuitEvent lets the game veto shutdown by returning true from quit(),
// which is how LOVE's love.quit callback works.
func (engine *Engine) handleQuitEvent() {
	result := engine.callback("quit")

	if cancelled, ok := result.(*object.Boolean); ok && cancelled.Value {
		return
	}

	engine.Quit()
}

func (engine *Engine) handleKeyboardEvent(event *sdl.KeyboardEvent) {
	name := sdl.GetScancodeName(event.Keysym.Scancode)

	if event.Type == sdl.KEYDOWN {
		engine.callback("keypressed",
			&object.String{Value: name},
			&object.Boolean{Value: event.Repeat > 0},
		)

		return
	}

	engine.callback("keyreleased", &object.String{Value: name})
}

func (engine *Engine) handleTextInputEvent(event *sdl.TextInputEvent) {
	engine.callback("textinput", &object.String{Value: event.GetText()})
}

func (engine *Engine) handleMouseButtonEvent(event *sdl.MouseButtonEvent) {
	name := &object.String{Value: MouseButtonName(event.Button)}
	x := object.NewInt(int64(event.X))
	y := object.NewInt(int64(event.Y))

	if event.Type == sdl.MOUSEBUTTONDOWN {
		engine.callback("mousepressed", x, y, name, object.NewInt(int64(event.Clicks)))

		return
	}

	engine.callback("mousereleased", x, y, name)
}

func (engine *Engine) handleMouseMotionEvent(event *sdl.MouseMotionEvent) {
	engine.callback("mousemoved",
		object.NewInt(int64(event.X)),
		object.NewInt(int64(event.Y)),
		object.NewInt(int64(event.XRel)),
		object.NewInt(int64(event.YRel)),
	)
}

func (engine *Engine) handleMouseWheelEvent(event *sdl.MouseWheelEvent) {
	engine.WheelX += event.X
	engine.WheelY += event.Y

	engine.callback("wheelmoved",
		object.NewInt(int64(event.X)),
		object.NewInt(int64(event.Y)),
	)
}

func (engine *Engine) handleWindowEvent(event *sdl.WindowEvent) {
	switch event.Event {
	case sdl.WINDOWEVENT_SIZE_CHANGED, sdl.WINDOWEVENT_RESIZED:
		engine.Width = event.Data1
		engine.Height = event.Data2

		engine.callback("resize",
			object.NewInt(int64(event.Data1)),
			object.NewInt(int64(event.Data2)),
		)
	case sdl.WINDOWEVENT_FOCUS_GAINED:
		engine.HasFocus = true

		engine.callback("focus", &object.Boolean{Value: true})
	case sdl.WINDOWEVENT_FOCUS_LOST:
		engine.HasFocus = false

		engine.callback("focus", &object.Boolean{Value: false})
	}
}

func (engine *Engine) handleControllerDeviceEvent(event *sdl.ControllerDeviceEvent) {
	switch event.Type {
	case sdl.CONTROLLERDEVICEADDED:
		engine.OpenJoysticks()

		engine.callback("joystickadded", object.NewInt(int64(len(engine.Joysticks))))
	case sdl.CONTROLLERDEVICEREMOVED:
		engine.OpenJoysticks()

		engine.callback("joystickremoved", object.NewInt(int64(len(engine.Joysticks))))
	}
}

// callback invokes an optional Ghost function. Games only define the callbacks
// they need, so a missing function is not an error.
func (engine *Engine) callback(name string, args ...object.Object) object.Object {
	if engine.Ghost == nil {
		return nil
	}

	if _, ok := engine.Ghost.Scope.Environment.Get(name); !ok {
		return nil
	}

	result := engine.Ghost.Call(name, args)

	if err, ok := result.(*object.Error); ok {
		engine.reportError(name, err)
	}

	return result
}
