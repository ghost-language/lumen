package modules

import "ghostlang.org/x/ghost/ghost"

func Register() {
	ghost.RegisterModule("canvas", CanvasMethods, CanvasProperties)
	ghost.RegisterModule("color", ColorMethods, ColorProperties)
	ghost.RegisterModule("image", ImageMethods, ImageProperties)
	ghost.RegisterModule("keyboard", KeyboardMethods, KeyboardProperties)
	ghost.RegisterModule("mouse", MouseMethods, MouseProperties)
	ghost.RegisterModule("window", WindowMethods, WindowProperties)
}
