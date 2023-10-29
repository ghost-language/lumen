package lumen

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func (lumen *Lumen) handleQuitEvent(event *sdl.QuitEvent) {
	fmt.Println("Quitting Lumen...")

	lumen.Quit()
}

func (lumen *Lumen) handleKeyboardEvent(event *sdl.KeyboardEvent) {
	lumen.KeyboardState = sdl.GetKeyboardState()
}
