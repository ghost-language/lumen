package modules

import (
	ghostmodules "ghostlang.org/x/ghost/library/modules"

	"ghostlang.org/x/ghost/ghost"
)

// Register makes every Lumen module available to Ghost code. It runs once,
// before a game's source is executed.
func Register() {
	ghost.RegisterModule("audio", AudioMethods, AudioProperties)
	ghost.RegisterModule("canvas", CanvasMethods, CanvasProperties)
	ghost.RegisterModule("color", ColorMethods, ColorProperties)
	ghost.RegisterModule("filesystem", FilesystemMethods, FilesystemProperties)
	ghost.RegisterModule("font", FontMethods, FontProperties)
	ghost.RegisterModule("image", ImageMethods, ImageProperties)
	ghost.RegisterModule("joystick", JoystickMethods, JoystickProperties)
	ghost.RegisterModule("keyboard", KeyboardMethods, KeyboardProperties)
	ghost.RegisterModule("lumen", LumenMethods, LumenProperties)
	ghost.RegisterModule("mouse", MouseMethods, MouseProperties)
	ghost.RegisterModule("system", SystemMethods, SystemProperties)
	ghost.RegisterModule("timer", TimerMethods, TimerProperties)
	ghost.RegisterModule("window", WindowMethods, WindowProperties)

	// Ghost seeds its generator from the clock. A game is better served by a
	// fixed default: the same seed every run means a procedural level looks the
	// same each time it is opened, which is what makes it something to iterate
	// on rather than something that changes underfoot. A game that wants a
	// different world each run says so with math.randomSeed().
	ghostmodules.SeedRandom(1)
}
