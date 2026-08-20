package engine

import (
	"fmt"
	"os"

	"ghostlang.org/x/lumen/resources"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// audioChannels is how many sound effects can overlap. Games rarely need more,
// and each channel costs mixing time.
const audioChannels = 32

func (engine *Engine) initSDL() {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_GAMECONTROLLER | sdl.INIT_EVENTS); err != nil {
		panic(err)
	}

	if err := img.Init(img.INIT_JPG | img.INIT_PNG); err != nil {
		panic(err)
	}

	if err := ttf.Init(); err != nil {
		panic(err)
	}

	// Audio is optional: a machine with no sound device should still be able to
	// run the game, just silently.
	if err := mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 1024); err != nil {
		fmt.Fprintf(os.Stderr, "lumen: audio unavailable: %s\n", err)
	} else {
		mix.AllocateChannels(audioChannels)
	}

	// Nearest-pixel sampling by default, which is what pixel art needs. Games
	// can switch an individual image over with image.setFilter('linear').
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "0")

	engine.createSDLWindow()
	engine.createSDLRenderer()
	engine.loadAndSetDefaultFont()
}

func (engine *Engine) createSDLWindow() {
	window, err := sdl.CreateWindow(
		engine.Title,
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		engine.Width,
		engine.Height,
		sdl.WINDOW_SHOWN,
	)

	if err != nil {
		panic(err)
	}

	engine.Window = window
}

func (engine *Engine) createSDLRenderer() {
	renderer, err := sdl.CreateRenderer(engine.Window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)

	// Not every machine can give us an accelerated renderer: a headless CI box,
	// a remote desktop, or a driver-less VM will all refuse. Falling back to
	// software keeps the game running instead of failing at start-up.
	if err != nil {
		renderer, err = sdl.CreateRenderer(engine.Window, -1, sdl.RENDERER_SOFTWARE)
	}

	if err != nil {
		panic(err)
	}

	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	engine.Renderer = renderer
}

func (engine *Engine) quitSDL() {
	if engine.Renderer != nil {
		engine.Renderer.Destroy()
	}

	if engine.Window != nil {
		engine.Window.Destroy()
	}

	mix.CloseAudio()
	mix.Quit()
	ttf.Quit()
	img.Quit()
	sdl.Quit()
}

// loadAndSetDefaultFont loads the font Lumen embeds in its own binary, so text
// works before a game loads any assets and regardless of the working directory.
func (engine *Engine) loadAndSetDefaultFont() {
	defaultFontData = resources.DefaultFont

	font, err := NewDefaultFont(19)

	if err != nil {
		panic(err)
	}

	engine.DefaultFont = font
	engine.CurrentFont = font
}

// SetMode resizes the window and switches fullscreen on or off.
func (engine *Engine) SetMode(width, height int32, fullscreen bool) error {
	flags := uint32(0)

	if fullscreen {
		flags = sdl.WINDOW_FULLSCREEN_DESKTOP
	}

	if err := engine.Window.SetFullscreen(flags); err != nil {
		return err
	}

	if !fullscreen {
		engine.Window.SetSize(width, height)
		engine.Window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)
	}

	engine.Width, engine.Height = engine.Window.GetSize()

	return nil
}

// IsFullscreen reports whether the window currently covers the display.
func (engine *Engine) IsFullscreen() bool {
	flags := engine.Window.GetFlags()

	return flags&(sdl.WINDOW_FULLSCREEN|sdl.WINDOW_FULLSCREEN_DESKTOP) != 0
}

// loadSurface reads an image file into an SDL surface.
func loadSurface(path string) (*sdl.Surface, error) {
	return img.Load(path)
}
