package lumen

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func (engine *Engine) handleQuitEvent(event *sdl.QuitEvent) {
	fmt.Println("Quitting Lumen...")

	engine.Quit()
}

func (engine *Engine) handleKeyboardEvent(event *sdl.KeyboardEvent) {
	engine.KeyboardState = sdl.GetKeyboardState()
}
