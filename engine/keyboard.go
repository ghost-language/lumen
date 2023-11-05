package engine

import "github.com/veandco/go-sdl2/sdl"

func (engine *Engine) SetInitialKeyboardState() {
	engine.CurrentKeyboardState = sdl.GetKeyboardState()
	engine.PreviousKeyboardState = make([]uint8, len(engine.CurrentKeyboardState))
}

func (engine *Engine) UpdateKeyboardState() {
	copy(engine.PreviousKeyboardState, engine.CurrentKeyboardState)
}
