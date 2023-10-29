package lumen

import (
	"fmt"

	"ghostlang.org/x/ghost/ghost"
	"github.com/veandco/go-sdl2/sdl"
)

var frameDelay uint64

type Lumen struct {
	Title         string
	Width         int32
	Height        int32
	TargetFps     uint64
	IsRunning     bool
	Ghost         *ghost.Ghost
	Window        *sdl.Window
	Renderer      *sdl.Renderer
	Texture       *sdl.Texture
	KeyboardState []uint8
	MouseState    []uint32
}

func New(title string) (lumen *Lumen) {
	lumen = new(Lumen)

	lumen.Title = title
	lumen.TargetFps = 60
	lumen.Width = 800
	lumen.Height = 600

	frameDelay = 1000 / lumen.TargetFps

	lumen.initSDL()

	return lumen
}

func (lumen *Lumen) Run() {
	lumen.load()

	lumen.IsRunning = true

	for lumen.IsRunning {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				lumen.handleQuitEvent(e)
			case *sdl.KeyboardEvent:
				lumen.handleKeyboardEvent(e)
			}

			frameStart := sdl.GetTicks64()

			lumen.update()
			lumen.draw()

			frameTime := sdl.GetTicks64() - frameStart

			// Give the CPU some time to run calculations
			if frameDelay > frameTime {
				sdl.Delay(uint32(frameDelay - frameTime))
			}

			fps := 1000 / float64(sdl.GetTicks64()-frameStart)
			lumen.Window.SetTitle(fmt.Sprintf("%s (FPS: %d)", lumen.Title, int64(fps)))
		}
	}
}

func (lumen *Lumen) Quit() {
	lumen.IsRunning = false

	lumen.quitSDL()
}
