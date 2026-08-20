package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"ghostlang.org/x/ghost/ghost"
	"github.com/veandco/go-sdl2/sdl"
)

// Lumen is the running engine. Ghost's module layer is a set of free functions
// with no receiver of its own, so the engine is reachable through this package
// level handle.
var Lumen *Engine

// Version is the engine version reported to games through lumen.version.
const Version = "0.2.0"

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

	saveIdentity string
	lastTicks    uint64
	errors       errorReporter
	loadFailed   bool
}

func New(title string) *Engine {
	Lumen = new(Engine)

	Lumen.Title = title
	Lumen.TargetFps = 60
	Lumen.Width = 800
	Lumen.Height = 600
	Lumen.HasFocus = true
	Lumen.MasterVolume = 1
	Lumen.Graphics = NewGraphics()
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

	// A game whose load() failed is missing the state every later frame assumes,
	// so it would do nothing but repeat the same error until the player closed
	// it. Stopping here reports the real problem once, at the top of the output.
	if engine.loadFailed {
		engine.flushRepeats()
		engine.shutdown()

		os.Exit(1)
	}

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

		frameTime := sdl.GetTicks64() - frameStart

		// Hand time back to the CPU when the frame finished early.
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

	engine.flushRepeats()
	engine.shutdown()
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
	config, err := os.UserConfigDir()

	if err != nil {
		return "", err
	}

	directory := filepath.Join(config, "lumen", engine.saveIdentity)

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}

	return directory, nil
}
