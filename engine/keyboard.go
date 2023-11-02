package engine

import "github.com/veandco/go-sdl2/sdl"

func (engine *Engine) UpdateKeyboardState() {
	engine.PreviousKeyboardState = engine.CurrentKeyboardState
	engine.CurrentKeyboardState = sdl.GetKeyboardState()
}
