package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/ghost/value"
	"ghostlang.org/x/lumen/engine"
	"github.com/veandco/go-sdl2/sdl"
)

var JoystickMethods = map[string]*object.LibraryFunction{}
var JoystickProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(JoystickMethods, "isDown", joystickIsDownMethod)
	modules.RegisterMethod(JoystickMethods, "isUp", joystickIsUpMethod)
	modules.RegisterMethod(JoystickMethods, "wasPressed", joystickWasPressedMethod)
	modules.RegisterMethod(JoystickMethods, "wasReleased", joystickWasReleasedMethod)
	modules.RegisterMethod(JoystickMethods, "getAxis", joystickGetAxisMethod)
	modules.RegisterMethod(JoystickMethods, "getName", joystickGetNameMethod)
	modules.RegisterMethod(JoystickMethods, "isConnected", joystickIsConnectedMethod)
	modules.RegisterMethod(JoystickMethods, "vibrate", joystickVibrateMethod)

	modules.RegisterProperty(JoystickProperties, "count", joystickCountProperty)
}

// Controllers are addressed by a 1-based index so Ghost code can talk about
// "player 1" without an off-by-one dance.

func joystickIsDownMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return buttonPredicate("joystick.isDown", tok, args, (*engine.Joystick).IsDown)
}

func joystickIsUpMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return buttonPredicate("joystick.isUp", tok, args, func(joystick *engine.Joystick, button sdl.GameControllerButton) bool {
		return !joystick.IsDown(button)
	})
}

func joystickWasPressedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return buttonPredicate("joystick.wasPressed", tok, args, (*engine.Joystick).WasPressed)
}

func joystickWasReleasedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return buttonPredicate("joystick.wasReleased", tok, args, (*engine.Joystick).WasReleased)
}

// joystickGetAxisMethod reads a stick or trigger. Sticks report -1 to 1 and
// triggers 0 to 1. An optional third argument sets a dead zone, below which the
// axis reads exactly zero so a worn stick does not drift the player around.
func joystickGetAxisMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("joystick.getAxis", tok, args, 2, 3); err != nil {
		return err
	}

	index, err := integer("joystick.getAxis", tok, args, 0)

	if err != nil {
		return err
	}

	name, err := text("joystick.getAxis", tok, args, 1)

	if err != nil {
		return err
	}

	deadZone := 0.15

	if len(args) == 3 {
		given, err := number("joystick.getAxis", tok, args, 2)

		if err != nil {
			return err
		}

		deadZone = given
	}

	// The axis name is validated before the controller is looked up, so a typo is
	// reported even on a machine with no controller plugged in.
	axis, ok := engine.ControllerAxis(name)

	if !ok {
		return engine.Choice("joystick.getAxis", tok, name, engine.ControllerAxisNames()...)
	}

	joystick := engine.Lumen.Joystick(int(index))

	if joystick == nil {
		return object.NewFloat(0)
	}

	position := joystick.Axis(axis)

	if position < 0 && position > -deadZone {
		return object.NewFloat(0)
	}

	if position >= 0 && position < deadZone {
		return object.NewFloat(0)
	}

	return object.NewFloat(position)
}

func joystickGetNameMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("joystick.getName", tok, args, 1); err != nil {
		return err
	}

	index, err := integer("joystick.getName", tok, args, 0)

	if err != nil {
		return err
	}

	joystick := engine.Lumen.Joystick(int(index))

	if joystick == nil {
		return value.NULL
	}

	return &object.String{Value: joystick.Name}
}

func joystickIsConnectedMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("joystick.isConnected", tok, args, 1); err != nil {
		return err
	}

	index, err := integer("joystick.isConnected", tok, args, 0)

	if err != nil {
		return err
	}

	return &object.Boolean{Value: engine.Lumen.Joystick(int(index)) != nil}
}

// joystickVibrateMethod rumbles a controller: strength values run 0 to 1 and the
// duration is in seconds.
func joystickVibrateMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("joystick.vibrate", tok, args, 2, 4); err != nil {
		return err
	}

	values, err := numbers("joystick.vibrate", tok, args)

	if err != nil {
		return err
	}

	joystick := engine.Lumen.Joystick(int(values[0]))

	if joystick == nil {
		return value.NULL
	}

	low := values[1]
	high := values[1]
	duration := 0.2

	if len(values) > 2 {
		high = values[2]
	}

	if len(values) > 3 {
		duration = values[3]
	}

	joystick.Controller.Rumble(
		uint16(clamp(low, 0, 1)*65535),
		uint16(clamp(high, 0, 1)*65535),
		uint32(duration*1000),
	)

	return value.NULL
}

func joystickCountProperty(scope *object.Scope, tok token.Token) object.Object {
	return object.NewInt(int64(len(engine.Lumen.Joysticks)))
}

// buttonPredicate resolves a controller and button name, then applies the given
// test. An absent controller reads as "not pressed" rather than an error, so a
// game does not have to guard every button check on whether a pad is plugged in.
func buttonPredicate(name string, tok token.Token, args []object.Object, predicate func(*engine.Joystick, sdl.GameControllerButton) bool) object.Object {
	if err := arity(name, tok, args, 2); err != nil {
		return err
	}

	index, err := integer(name, tok, args, 0)

	if err != nil {
		return err
	}

	buttonName, err := text(name, tok, args, 1)

	if err != nil {
		return err
	}

	button, ok := engine.ControllerButton(buttonName)

	if !ok {
		return engine.Unknown(name, tok, "button", buttonName, engine.ControllerButtonNames()...)
	}

	joystick := engine.Lumen.Joystick(int(index))

	if joystick == nil {
		return value.FALSE
	}

	return &object.Boolean{Value: predicate(joystick, button)}
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}

	if value > high {
		return high
	}

	return value
}
