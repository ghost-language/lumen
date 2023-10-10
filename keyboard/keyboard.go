package keyboard

import (
	"github.com/veandco/go-sdl2/sdl"
)

var State []uint8
var PreviousState []uint8

func IsDown(scancode sdl.Scancode) bool {
	return State[scancode] == 1
}

func IsUp(scancode sdl.Scancode) bool {
	return !IsDown(scancode)
}

func IsPressed(scancode sdl.Scancode) bool {
	return State[scancode] == 1 && PreviousState[scancode] == 0
}

func IsReleased(scancode sdl.Scancode) bool {
	return State[scancode] == 0 && PreviousState[scancode] == 1
}

func UpdateState() {
	PreviousState = State
	// copy(PreviousState, State)
}
