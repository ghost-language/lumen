package engine

import (
	"ghostlang.org/x/ghost/ghost"
	"github.com/veandco/go-sdl2/sdl"
)

var Lumen *Engine

type Engine struct {
	Title                 string
	Width                 int32
	Height                int32
	TargetFps             uint64
	CurrentFps            uint64
	IsRunning             bool
	Resources             map[string]interface{}
	Ghost                 *ghost.Ghost
	Window                *sdl.Window
	Renderer              *sdl.Renderer
	Texture               *sdl.Texture
	DefaultFont           *Font
	CurrentFont           *Font
	PreviousKeyboardState []uint8
	CurrentKeyboardState  []uint8
	OffsetX               int32
	OffsetY               int32
}

func New(title string) (engine *Engine) {
	Lumen = new(Engine)
	Lumen.Resources = make(map[string]interface{})

	Lumen.Title = title
	Lumen.TargetFps = 60
	Lumen.Width = 800
	Lumen.Height = 600

	Lumen.initSDL()

	return Lumen
}

func (engine *Engine) Run() {
	engine.load()

	engine.IsRunning = true

	engine.UpdateKeyboardState()

	for engine.IsRunning {
		frameStart := sdl.GetTicks64()

		engine.handleEvents()
		engine.update()
		engine.draw()

		frameTime := sdl.GetTicks64() - frameStart

		// Give the CPU some time to run calculations
		if 1000.0/float64(engine.TargetFps) > float64(frameTime) {
			sdl.Delay(uint32(1000.0/float64(engine.TargetFps) - float64(frameTime)))
		}

		engine.CurrentFps = uint64(1000 / float64(sdl.GetTicks64()-frameStart))
	}
}

func (engine *Engine) Quit() {
	engine.IsRunning = false

	engine.FreeResources()
	engine.quitSDL()
}
