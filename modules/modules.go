package modules

import (
	ghostmodules "ghostlang.org/x/ghost/library/modules"

	"ghostlang.org/x/ghost/ghost"
)

// Scheme is the import scheme Lumen registers its own modules under, so a
// game reaches them the same way it reaches Ghost's standard library, just
// with its own prefix: `import "lumen:canvas"` rather than borrowing Ghost's
// own `import "ghost:math"`.
const Scheme = "lumen"

// Register makes every Lumen module and native class available to Ghost code.
// It runs once, before a game's source is executed.
func Register() {
	ghost.RegisterModuleForScheme(Scheme, "audio", AudioMethods, AudioProperties)
	ghost.RegisterModuleForScheme(Scheme, "canvas", CanvasMethods, CanvasProperties)
	ghost.RegisterModuleForScheme(Scheme, "color", ColorMethods, ColorProperties)
	ghost.RegisterModuleForScheme(Scheme, "filesystem", FilesystemMethods, FilesystemProperties)
	ghost.RegisterModuleForScheme(Scheme, "font", FontMethods, FontProperties)
	ghost.RegisterModuleForScheme(Scheme, "image", ImageMethods, ImageProperties)
	ghost.RegisterModuleForScheme(Scheme, "joystick", JoystickMethods, JoystickProperties)
	ghost.RegisterModuleForScheme(Scheme, "keyboard", KeyboardMethods, KeyboardProperties)
	ghost.RegisterModuleForScheme(Scheme, "lumen", LumenMethods, LumenProperties)
	ghost.RegisterModuleForScheme(Scheme, "mouse", MouseMethods, MouseProperties)
	ghost.RegisterModuleForScheme(Scheme, "system", SystemMethods, SystemProperties)
	ghost.RegisterModuleForScheme(Scheme, "timer", TimerMethods, TimerProperties)
	ghost.RegisterModuleForScheme(Scheme, "window", WindowMethods, WindowProperties)

	// Native classes: host resources built and driven entirely by Go, `new`-able
	// exactly like a Ghost-defined class once imported — `import { Image } from
	// "lumen:image"` then `new Image("player.png")`.
	ghost.RegisterClassForScheme(Scheme, "image", "Image", imageConstructor)
	ghost.RegisterClassForScheme(Scheme, "image", "Spritesheet", spritesheetConstructor)
	ghost.RegisterClassForScheme(Scheme, "image", "Animation", animationConstructor)
	ghost.RegisterClassForScheme(Scheme, "audio", "Source", sourceConstructor)
	ghost.RegisterClassForScheme(Scheme, "font", "Font", fontConstructor)
	ghost.RegisterClassForScheme(Scheme, "canvas", "Target", targetConstructor)
	ghost.RegisterClassForScheme(Scheme, "canvas", "Quad", quadConstructor)

	// Ghost seeds its generator from the clock. A game is better served by a
	// fixed default: the same seed every run means a procedural level looks the
	// same each time it is opened, which is what makes it something to iterate
	// on rather than something that changes underfoot. A game that wants a
	// different world each run says so with math.randomSeed().
	ghostmodules.SeedRandom(1)
}
