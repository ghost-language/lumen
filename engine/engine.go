package engine

import (
	"fmt"

	"ghostlang.org/x/ghost/ghost"
	"github.com/veandco/go-sdl2/sdl"
)

var Lumen *Engine
var frameDelay uint64

type Engine struct {
	Title         string
	Width         int32
	Height        int32
	TargetFps     uint64
	CurrentFps    uint64
	IsRunning     bool
	Resources     map[string]interface{}
	Ghost         *ghost.Ghost
	Window        *sdl.Window
	Renderer      *sdl.Renderer
	Texture       *sdl.Texture
	KeyboardState []uint8
	MouseState    []uint32
}

func New(title string) (engine *Engine) {
	Lumen = new(Engine)
	Lumen.Resources = make(map[string]interface{})

	Lumen.Title = title
	Lumen.TargetFps = 60
	Lumen.Width = 800
	Lumen.Height = 600

	frameDelay = 1000 / Lumen.TargetFps

	Lumen.initSDL()

	return Lumen
}

func (engine *Engine) Run() {
	engine.load()

	engine.IsRunning = true

	for engine.IsRunning {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				engine.handleQuitEvent(e)
			case *sdl.KeyboardEvent:
				engine.handleKeyboardEvent(e)
			}
		}

		frameStart := sdl.GetTicks64()

		engine.update()
		engine.draw()

		frameTime := sdl.GetTicks64() - frameStart

		// Give the CPU some time to run calculations
		if frameDelay > frameTime {
			sdl.Delay(uint32(frameDelay - frameTime))
		}

		engine.CurrentFps = uint64(1000 / float64(sdl.GetTicks64()-frameStart))
		engine.Window.SetTitle(fmt.Sprintf("%s (FPS: %d)", engine.Title, engine.CurrentFps))
	}
}

func (engine *Engine) Quit() {
	engine.IsRunning = false

	engine.FreeResources()
	engine.quitSDL()
}
