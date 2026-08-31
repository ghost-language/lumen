package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"ghostlang.org/x/ghost/ghost"
	"github.com/veandco/go-sdl2/sdl"
)

// Lumen is the running engine. Ghost's module layer is a set of free functions
// with no receiver of its own, so the engine is reachable through this package
// level handle.
var Lumen *Engine

// Version is the engine version reported to games through lumen.version.
const Version = "0.2.1"

// maximumDelta caps the delta time handed to update(). Without it, a frame that
// stalls (a breakpoint, a window drag, a slow asset load) would hand the game a
// huge step and teleport everything through walls.
const maximumDelta = 0.25

type Engine struct {
	Title      string
	Width      int32
	Height     int32
	TargetFps  uint64
	CurrentFps uint64
	IsRunning  bool
	HasFocus   bool

	// LogicalWidth and LogicalHeight are the size of the coordinate space the
	// game draws in, which is not always the size of the window. Zero means the
	// game draws straight into window pixels. See viewport.go.
	LogicalWidth  int32
	LogicalHeight int32
	PixelPerfect  bool

	// WindowedWidth and WindowedHeight remember how big the window was before
	// it went fullscreen, so leaving fullscreen puts it back.
	WindowedWidth  int32
	WindowedHeight int32

	// OutputWidth and OutputHeight are the renderer's size in real pixels.
	OutputWidth  int32
	OutputHeight int32

	// Delta is the number of seconds the previous frame took. Every piece of
	// game logic that moves should scale by it so speeds stay the same
	// regardless of frame rate.
	Delta        float64
	AverageDelta float64
	ElapsedTime  float64
	FrameCount   uint64

	MasterVolume float64

	Resources []Releasable
	Ghost     *ghost.Ghost
	Window    *sdl.Window
	Renderer  *sdl.Renderer
	Graphics  *Graphics

	DefaultFont *Font
	CurrentFont *Font

	PreviousKeyboardState []uint8
	CurrentKeyboardState  []uint8
	PreviousMouseState    uint32
	CurrentMouseState     uint32
	MouseX                int32
	MouseY                int32
	WheelX                int32
	WheelY                int32
	Joysticks             []*Joystick

	batch        batch
	images       map[string]*Image
	saveIdentity string
	viewScale    float64
	viewport     sdl.Rect
	lastTicks    uint64
	errors       errorReporter

	// reportFont is the built-in font at heading size, loaded the first time an
	// error screen is drawn and not before: a game that never fails never pays
	// for a second copy of the font.
	reportFont *Font

	// report is the failure the game has stopped on, and nil while it is
	// running. See report.go: while one is set, no game code is called and the
	// window shows the error screen instead of a frame.
	report *Report
}

func New(title string) *Engine {
	Lumen = new(Engine)

	Lumen.Title = title
	Lumen.TargetFps = 60
	Lumen.Width = 800
	Lumen.Height = 600
	Lumen.WindowedWidth = 800
	Lumen.WindowedHeight = 600
	Lumen.HasFocus = true
	Lumen.MasterVolume = 1
	Lumen.Graphics = NewGraphics()
	Lumen.viewScale = 1
	Lumen.saveIdentity = "lumen"

	Lumen.initSDL()

	return Lumen
}

// Run drives the game loop: load once, then update and draw until the game quits.
func (engine *Engine) Run() {
	engine.IsRunning = true

	engine.SetInitialKeyboardState()
	engine.SetInitialMouseState()
	engine.OpenJoysticks()

	engine.load()

	engine.lastTicks = sdl.GetTicks64()

	for engine.IsRunning {
		frameStart := sdl.GetTicks64()

		// Input is handled in three steps, and the order matters. The previous
		// frame's snapshots are put aside first, then SDL's queue is drained,
		// then the new state is read. Reading before the pump would report
		// whatever was true last frame, leaving the pointer a frame behind the
		// cursor and every wasPressed() a frame late.
		engine.AgeKeyboardState()
		engine.AgeMouseState()
		engine.AgeJoystickState()

		engine.handleEvents()

		engine.SampleMouseState()
		engine.SampleJoystickState()

		if !engine.IsRunning {
			break
		}

		// A halted game draws its error screen and nothing else. Its clock is
		// not advanced either: whenever it is resumed, it resumes into a frame
		// that took as long as a frame, rather than one that took as long as
		// the reader spent reading.
		if engine.Halted() {
			// Unless there is nobody there. A game run headlessly — in CI, in a
			// build script, over ssh — has a window nobody can see and no way to
			// dismiss what is on it, so holding it open would hang the run
			// rather than report it. The console already has the report.
			if engine.headless() {
				break
			}

			engine.drawReport()
			engine.lastTicks = sdl.GetTicks64()
			engine.pace(frameStart)

			continue
		}

		engine.Delta = float64(frameStart-engine.lastTicks) / 1000.0
		engine.lastTicks = frameStart

		if engine.Delta > maximumDelta {
			engine.Delta = maximumDelta
		}

		engine.ElapsedTime += engine.Delta

		// A rolling average smooths out the single-frame spikes that make a raw
		// delta useless for anything a player would read.
		engine.AverageDelta = engine.AverageDelta*0.9 + engine.Delta*0.1

		engine.update()
		engine.draw()

		engine.FrameCount++

		if engine.FrameCount%60 == 0 {
			engine.PruneCaches()
		}

		engine.pace(frameStart)
	}

	// A game that was closed while an error was on screen did not finish, and
	// the process should not claim it did.
	if engine.Halted() {
		engine.errors.failed = true
	}

	engine.flushRepeats()
	engine.shutdown()
}

// headless reports whether SDL is drawing into nothing, which is what it does
// on a machine with no display and what a CI run asks it for by name.
func (engine *Engine) headless() bool {
	driver, err := sdl.GetCurrentVideoDriver()

	if err != nil {
		return false
	}

	switch driver {
	case "dummy", "offscreen":
		return true
	}

	return false
}

// pace hands time back to the CPU when a frame finished early, and records how
// long the frame took in the end.
func (engine *Engine) pace(frameStart uint64) {
	frameTime := sdl.GetTicks64() - frameStart

	if engine.TargetFps > 0 {
		budget := 1000.0 / float64(engine.TargetFps)

		if budget > float64(frameTime) {
			sdl.Delay(uint32(budget - float64(frameTime)))
		}
	}

	elapsed := sdl.GetTicks64() - frameStart

	if elapsed > 0 {
		engine.CurrentFps = 1000 / elapsed
	}
}

// Quit asks the game loop to stop at the end of the current frame.
func (engine *Engine) Quit() {
	engine.IsRunning = false
}

// shutdown releases everything the engine owns. It runs after the loop exits so
// a game is never torn down mid-frame.
func (engine *Engine) shutdown() {
	engine.CloseJoysticks()
	engine.FreeResources()
	engine.quitSDL()
}

// SetSaveIdentity names the folder inside the user's config directory that this
// game's saved data lives in.
func (engine *Engine) SetSaveIdentity(identity string) error {
	cleaned := filepath.Base(filepath.Clean(identity))

	if cleaned == "." || cleaned == string(filepath.Separator) || cleaned == "" {
		return fmt.Errorf("invalid save identity: %q", identity)
	}

	engine.saveIdentity = cleaned

	return nil
}

// SaveDirectory returns the directory this game's saved data lives in, creating
// it on first use.
func (engine *Engine) SaveDirectory() (string, error) {
	base, err := userDataDir()

	if err != nil {
		return "", err
	}

	directory := filepath.Join(base, "lumen", engine.saveIdentity)

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}

	return directory, nil
}

// userDataDir returns the per-user directory for application data, which is
// where a saved game belongs. Go's os.UserConfigDir is the right answer on
// macOS and Windows, but on Linux it resolves to ~/.config, which is for
// configuration. Saves are data, so Linux follows the XDG data directory:
// ~/.local/share.
func userDataDir() (string, error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "freebsd" && runtime.GOOS != "openbsd" && runtime.GOOS != "netbsd" {
		return os.UserConfigDir()
	}

	if data := os.Getenv("XDG_DATA_HOME"); data != "" && filepath.IsAbs(data) {
		return data, nil
	}

	home, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share"), nil
}
