package modules

import "ghostlang.org/x/ghost/ghost"

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

	registerMathExtensions()
}
