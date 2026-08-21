package engine

import "github.com/veandco/go-sdl2/sdl"

// MouseButtonMask maps the button names a game uses onto SDL's button bitmask.
func MouseButtonMask(button string) (uint32, bool) {
	switch button {
	case "left", "1":
		return sdl.ButtonLMask(), true
	case "middle", "3":
		return sdl.ButtonMMask(), true
	case "right", "2":
		return sdl.ButtonRMask(), true
	case "x1", "4":
		return sdl.ButtonX1Mask(), true
	case "x2", "5":
		return sdl.ButtonX2Mask(), true
	}

	return 0, false
}

// MouseButtonName maps an SDL button index back to its Lumen name, for the
// mousepressed and mousereleased callbacks.
func MouseButtonName(button uint8) string {
	switch button {
	case sdl.BUTTON_LEFT:
		return "left"
	case sdl.BUTTON_MIDDLE:
		return "middle"
	case sdl.BUTTON_RIGHT:
		return "right"
	case sdl.BUTTON_X1:
		return "x1"
	case sdl.BUTTON_X2:
		return "x2"
	}

	return "unknown"
}

// SetInitialMouseState captures the pointer state before the first frame.
func (engine *Engine) SetInitialMouseState() {
	x, y, state := sdl.GetMouseState()

	engine.MouseX = x
	engine.MouseY = y
	engine.CurrentMouseState = state
	engine.PreviousMouseState = state
}

// AgeMouseState rolls the pointer snapshot forward and clears the wheel delta,
// which only describes the frame it arrived in. It runs before events are
// pumped, so the wheel events that arrive during the pump land in the frame
// that reads them.
func (engine *Engine) AgeMouseState() {
	engine.PreviousMouseState = engine.CurrentMouseState
	engine.WheelX = 0
	engine.WheelY = 0
}

// SampleMouseState reads the live pointer state. It runs after events are
// pumped: SDL's mouse state only reflects events it has already processed, so
// sampling any earlier would report the previous frame's position.
func (engine *Engine) SampleMouseState() {
	x, y, state := sdl.GetMouseState()

	engine.MouseX = x
	engine.MouseY = y
	engine.CurrentMouseState = state
}
