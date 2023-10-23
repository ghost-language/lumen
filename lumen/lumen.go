package lumen

import (
	"fmt"

	"ghostlang.org/x/ghost/ghost"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/lumen/keyboard"
	"ghostlang.org/x/lumen/renderer"
	"ghostlang.org/x/lumen/window"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	gameRunning bool
	err         error
	frameDelay  uint64
)

type Lumen struct {
	Title  string
	Width  int32
	Height int32
	Fps    uint64

	GraphicsMethods    map[string]*object.LibraryFunction
	GraphicsProperties map[string]*object.LibraryProperty
}

func New(title string) (lumen *Lumen) {
	lumen = new(Lumen)

	lumen.Title = title
	lumen.Fps = 60
	lumen.Width = 800
	lumen.Height = 600

	frameDelay = 1000 / lumen.Fps

	err = img.Init(img.INIT_JPG | img.INIT_PNG)

	if err != nil {
		panic(err)
	}

	err = sdl.Init(sdl.INIT_EVERYTHING)

	if err != nil {
		panic(err)
	}

	window.New(lumen.Title, lumen.Width, lumen.Height)

	if err != nil {
		panic(err)
	}

	renderer.New()

	if err != nil {
		panic(err)
	}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "2")

	return lumen
}

func (lumen *Lumen) Run(ghost *ghost.Ghost) {
	ghost.Call("load", nil)

	gameRunning = true

	keyboard.State = sdl.GetKeyboardState()
	keyboard.PreviousState = make([]uint8, len(keyboard.State))

	for gameRunning {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quitting Lumen...")
				gameRunning = false
			case *sdl.KeyboardEvent:
				keyboard.State = sdl.GetKeyboardState()
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

		keyboard.UpdateState()

		// fmt.Println("FPS:", 1000.0/float64(sdl.GetTicks64()-frameStart))
	}
}

func (lumen *Lumen) update(ghost *ghost.Ghost) {
	ghost.Call("update", nil)
}

func (lumen *Lumen) draw(ghost *ghost.Ghost) {
	renderer.Renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Renderer.Clear()

	ghost.Call("draw", nil)

	renderer.Renderer.Present()
}
