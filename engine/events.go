package engine

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func (engine *Engine) handleEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			engine.handleQuitEvent(e)
		case *sdl.KeyboardEvent:
			engine.handleKeyboardEvent(e)
		}
	}
}

func (engine *Engine) handleQuitEvent(event *sdl.QuitEvent) {
	fmt.Println("Quitting Lumen...")

	engine.Quit()
}

func (engine *Engine) handleKeyboardEvent(event *sdl.KeyboardEvent) {
	engine.UpdateKeyboardState()
}
