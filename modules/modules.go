package modules

import "ghostlang.org/x/ghost/ghost"

func Register() {
	ghost.RegisterModule("canvas", CanvasMethods, CanvasProperties)
	ghost.RegisterModule("color", ColorMethods, ColorProperties)
	ghost.RegisterModule("window", WindowMethods, WindowProperties)
}
