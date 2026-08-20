package engine

import "github.com/veandco/go-sdl2/sdl"

// SetInitialKeyboardState captures the keyboard state before the first frame.
// SDL hands back a slice that stays live for the lifetime of the program, so
// edge detection works by keeping a copy of the previous frame's snapshot
// alongside it.
func (engine *Engine) SetInitialKeyboardState() {
	engine.CurrentKeyboardState = sdl.GetKeyboardState()
	engine.PreviousKeyboardState = make([]uint8, len(engine.CurrentKeyboardState))

	copy(engine.PreviousKeyboardState, engine.CurrentKeyboardState)
}

// AgeKeyboardState rolls the current snapshot into the previous one. It runs
// once per frame, before events are pumped, so wasPressed() and wasReleased()
// describe exactly one frame's worth of change.
//
// There is no matching sample step: SDL hands back a slice it keeps updating in
// place, so the current state is already live once events have been pumped.
func (engine *Engine) AgeKeyboardState() {
	copy(engine.PreviousKeyboardState, engine.CurrentKeyboardState)
}

// Scancode resolves a key name to its SDL scancode. Names are matched the way
// SDL matches them, so 'left', 'Left', and 'LEFT' all work.
func Scancode(name string) (sdl.Scancode, bool) {
	scancode := sdl.GetScancodeFromName(name)

	if scancode == sdl.SCANCODE_UNKNOWN {
		return scancode, false
	}

	return scancode, true
}

// IsKeyDown reports whether a key is held this frame.
func (engine *Engine) IsKeyDown(scancode sdl.Scancode) bool {
	if int(scancode) >= len(engine.CurrentKeyboardState) {
		return false
	}

	return engine.CurrentKeyboardState[scancode] == 1
}

// WasKeyDown reports whether a key was held last frame.
func (engine *Engine) WasKeyDown(scancode sdl.Scancode) bool {
	if int(scancode) >= len(engine.PreviousKeyboardState) {
		return false
	}

	return engine.PreviousKeyboardState[scancode] == 1
}
