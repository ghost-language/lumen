package engine

import "github.com/veandco/go-sdl2/sdl"

// Joystick wraps an SDL game controller together with the button snapshots that
// make wasPressed()/wasReleased() possible.
type Joystick struct {
	ID         sdl.JoystickID
	Name       string
	Controller *sdl.GameController

	current  [sdl.CONTROLLER_BUTTON_MAX]bool
	previous [sdl.CONTROLLER_BUTTON_MAX]bool
}

// OpenJoysticks connects every game controller SDL can see. It runs at startup
// and again whenever a controller is plugged in or unplugged.
func (engine *Engine) OpenJoysticks() {
	engine.CloseJoysticks()

	for index := 0; index < sdl.NumJoysticks(); index++ {
		if !sdl.IsGameController(index) {
			continue
		}

		controller := sdl.GameControllerOpen(index)

		if controller == nil {
			continue
		}

		joystick := &Joystick{
			ID:         controller.Joystick().InstanceID(),
			Name:       controller.Name(),
			Controller: controller,
		}

		engine.Joysticks = append(engine.Joysticks, joystick)
	}
}

// CloseJoysticks disconnects every open controller.
func (engine *Engine) CloseJoysticks() {
	for _, joystick := range engine.Joysticks {
		if joystick.Controller != nil {
			joystick.Controller.Close()
		}
	}

	engine.Joysticks = nil
}

// AgeJoystickState rolls each controller's button snapshot forward one frame.
func (engine *Engine) AgeJoystickState() {
	for _, joystick := range engine.Joysticks {
		joystick.previous = joystick.current
	}
}

// SampleJoystickState reads the live button state of every controller. Like the
// mouse, SDL only updates it during the event pump, so this runs afterwards.
func (engine *Engine) SampleJoystickState() {
	for _, joystick := range engine.Joysticks {
		for button := sdl.GameControllerButton(0); button < sdl.CONTROLLER_BUTTON_MAX; button++ {
			joystick.current[button] = joystick.Controller.Button(button) == sdl.PRESSED
		}
	}
}

// Joystick returns the controller at the given index, counting from 1 to match
// how Ghost code refers to players.
func (engine *Engine) Joystick(index int) *Joystick {
	if index < 1 || index > len(engine.Joysticks) {
		return nil
	}

	return engine.Joysticks[index-1]
}

// IsDown reports whether a controller button is held this frame.
func (joystick *Joystick) IsDown(button sdl.GameControllerButton) bool {
	return joystick.current[button]
}

// WasPressed reports whether a button went down this frame.
func (joystick *Joystick) WasPressed(button sdl.GameControllerButton) bool {
	return joystick.current[button] && !joystick.previous[button]
}

// WasReleased reports whether a button came up this frame.
func (joystick *Joystick) WasReleased(button sdl.GameControllerButton) bool {
	return !joystick.current[button] && joystick.previous[button]
}

// Axis returns an axis position. Sticks report -1 to 1 and triggers 0 to 1.
func (joystick *Joystick) Axis(axis sdl.GameControllerAxis) float64 {
	return float64(joystick.Controller.Axis(axis)) / 32767.0
}

// ControllerButton resolves a button name such as 'a', 'start', or 'dpup'.
func ControllerButton(name string) (sdl.GameControllerButton, bool) {
	button := sdl.GameControllerGetButtonFromString(name)

	if button == sdl.CONTROLLER_BUTTON_INVALID {
		return button, false
	}

	return button, true
}

// ControllerAxis resolves an axis name such as 'leftx' or 'triggerright'.
func ControllerAxis(name string) (sdl.GameControllerAxis, bool) {
	axis := sdl.GameControllerGetAxisFromString(name)

	if axis == sdl.CONTROLLER_AXIS_INVALID {
		return axis, false
	}

	return axis, true
}
