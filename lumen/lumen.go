package lumen

import (
	"fmt"

	"ghostlang.org/x/ghost/ghost"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	gameRunning bool
	err         error
	frameDelay  uint64
)

type Lumen struct {
	title  string
	width  int32
	height int32
	fps    uint64

	window   *sdl.Window
	renderer *sdl.Renderer
}

func New(title string) (lumen *Lumen) {
	lumen = new(Lumen)

	lumen.title = title
	lumen.fps = 60
	lumen.width = 800
	lumen.height = 600

	frameDelay = 1000 / lumen.fps

	return lumen
}

func (lumen *Lumen) initialize(ghost *ghost.Ghost) {
	err = sdl.Init(sdl.INIT_EVERYTHING)

	if err != nil {
		panic(err)
	}

	lumen.window, err = sdl.CreateWindow(
		lumen.title,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		lumen.width,
		lumen.height,
		sdl.WINDOW_SHOWN,
	)

	if err != nil {
		panic(err)
	}

	lumen.renderer, err = sdl.CreateRenderer(lumen.window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil {
		panic(err)
	}

	ghost.Call("load", nil)

	gameRunning = true
}

func (lumen *Lumen) Run(ghost *ghost.Ghost) {
	lumen.initialize(ghost)

	for gameRunning {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quitting Lumen...")
				gameRunning = false
			}
		}

		frameStart := sdl.GetTicks64()

		lumen.update(ghost)
		lumen.draw(ghost)

		frameTime := sdl.GetTicks64() - frameStart

		// Give the CPU some time to run calculations
		if frameDelay > frameTime {
			sdl.Delay(uint32(frameDelay - frameTime))
		}
	}
}

func (lumen *Lumen) update(ghost *ghost.Ghost) {
	ghost.Call("update", nil)
}

func (lumen *Lumen) draw(ghost *ghost.Ghost) {
	lumen.renderer.SetDrawColor(26, 32, 44, 255)
	lumen.renderer.Clear()

	ghost.Call("draw", nil)

	lumen.renderer.Present()
}
