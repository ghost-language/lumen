package engine

import (
	"fmt"
	"os"

	"ghostlang.org/x/ghost/color"
	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/lumen/resources"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// audioChannels is how many sound effects can overlap. Games rarely need more,
// and each channel costs mixing time.
const audioChannels = 32

// initSDL brings up everything a game needs before its first line runs.
//
// None of this can be reported into the window, because the failures here are
// the reasons there is no window. They are reported to the console in the same
// shape everything else is — a kind, a sentence, and a line about what to do —
// and then the process stops, because there is nothing to stop for.
func (engine *Engine) initSDL() {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_GAMECONTROLLER | sdl.INIT_EVENTS); err != nil {
		fatal(fault.New(fault.System, "SDL could not start: %s", err).
			WithHelp("this machine has no display Lumen can open a window on; SDL_VIDEODRIVER=dummy runs a game headless"))
	}

	if err := img.Init(img.INIT_JPG | img.INIT_PNG); err != nil {
		fatal(fault.New(fault.System, "image loading could not start: %s", err).
			WithHelp("SDL2_image is missing or too old; a game cannot load PNGs or JPEGs without it"))
	}

	if err := ttf.Init(); err != nil {
		fatal(fault.New(fault.System, "font rendering could not start: %s", err).
			WithHelp("SDL2_ttf is missing or too old; a game cannot draw text without it"))
	}

	// Audio is optional: a machine with no sound device should still be able to
	// run the game, just silently. It is worth a word, though — a game that is
	// silent for a reason nobody mentioned reads as a game with broken sound.
	if err := mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 1024); err != nil {
		warn("audio is unavailable, and the game will run silently: %s", err)
	} else {
		mix.AllocateChannels(audioChannels)
	}

	// Nearest-pixel sampling by default, which is what pixel art needs. Games
	// can switch an individual image over with image.setFilter('linear').
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "0")

	engine.createSDLWindow()
	engine.createSDLRenderer()
	engine.loadAndSetDefaultFont()

	engine.UpdateViewport()
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
		fatal(fault.New(fault.System, "the game window could not be opened: %s", err))
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
		fatal(fault.New(fault.System, "the game window could not be drawn into: %s", err).
			WithHelp("neither an accelerated nor a software renderer would start on this machine"))
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
		fatal(fault.New(fault.System, "Lumen's built-in font could not be loaded: %s", err).
			WithHelp("this is the font Lumen carries inside its own binary; a failure here is a bug in Lumen, please report it"))
	}

	engine.DefaultFont = font
	engine.CurrentFont = font
}

// SetMode resizes the window and switches fullscreen on or off. The size is
// remembered as the game's windowed size even when the call puts it fullscreen,
// so leaving fullscreen restores the window rather than leaving a desktop-sized
// window behind.
func (engine *Engine) SetMode(width, height int32, fullscreen bool) error {
	if width > 0 && height > 0 {
		engine.WindowedWidth = width
		engine.WindowedHeight = height
	}

	flags := uint32(0)

	if fullscreen {
		flags = sdl.WINDOW_FULLSCREEN_DESKTOP
	}

	if err := engine.Window.SetFullscreen(flags); err != nil {
		return err
	}

	if !fullscreen {
		engine.Window.SetSize(engine.WindowedWidth, engine.WindowedHeight)
		engine.Window.SetPosition(sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED)
	}

	engine.Width, engine.Height = engine.Window.GetSize()

	engine.UpdateViewport()

	return nil
}

// SetFullscreen switches fullscreen on or off at the size the window already
// has, which for a windowed game is the size it was given at start-up.
func (engine *Engine) SetFullscreen(fullscreen bool) error {
	return engine.SetMode(engine.WindowedWidth, engine.WindowedHeight, fullscreen)
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

// fatal reports a failure that happened before there was a window to report it
// in, and stops. Start-up is the one place in Lumen where a failure ends the
// process: everything after it has somewhere to put a report and someone
// looking at it.
func fatal(raised *fault.Fault) {
	fmt.Fprintln(os.Stderr, raised.Render(color.Detect(os.Stderr)))

	os.Exit(1)
}

// warn reports something a game's author should know about that is not going to
// stop the game.
func warn(format string, arguments ...interface{}) {
	profile := color.Detect(os.Stderr)

	fmt.Fprintf(os.Stderr, "%s %s\n", profile.Warning("warning:"), fmt.Sprintf(format, arguments...))
}
