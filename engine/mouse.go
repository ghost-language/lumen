package engine

import "github.com/veandco/go-sdl2/sdl"

func (engine *Engine) SetInitialMouseState() {
	_, _, state := sdl.GetMouseState()

	engine.CurrentMouseState = state
}

func (engine *Engine) UpdateMouseState() {
	engine.PreviousMouseState = engine.CurrentMouseState

	_, _, state := sdl.GetMouseState()

	engine.CurrentMouseState = state
}
