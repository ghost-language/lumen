package renderer

import (
	"ghostlang.org/x/lumen/window"
	"github.com/veandco/go-sdl2/sdl"
)

var Renderer *sdl.Renderer

func New() *sdl.Renderer {
	var err error

	Renderer, err = sdl.CreateRenderer(window.Window.Window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "2")

	if err != nil {
		panic(err)
	}

	return Renderer
}
