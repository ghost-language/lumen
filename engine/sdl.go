package engine

import (
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

func (engine *Engine) initSDL() {
	var err error

	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	if err = img.Init(img.INIT_JPG | img.INIT_PNG); err != nil {
		panic(err)
	}

	engine.createSDLWindow()
	engine.createSDLRenderer()
	engine.createSDLTexture()

	// Set the scaling quality. This hint is checked when a texture is created
	// and it affects scaling when copying that texture.
	//
	// 0 = nearest pixel sampling
	// 1 = linear filtering (supported by OpenGL and Direct3D)
	// 2 = anisotropic filtering (supported by Direct3D)
	//
	// TODO: Make this configurable
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "0")
}

func (engine *Engine) createSDLWindow() {
	var err error

	// Create the window
	engine.Window, err = sdl.CreateWindow(
		engine.Title,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		engine.Width,
		engine.Height,
		sdl.WINDOW_SHOWN,
	)

	if err != nil {
		panic(err)
	}
}

func (engine *Engine) createSDLRenderer() {
	var err error

	// Create the renderer
	engine.Renderer, err = sdl.CreateRenderer(engine.Window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)

	if err != nil {
		panic(err)
	}
}

func (engine *Engine) createSDLTexture() {
	var err error

	// Create the texture
	engine.Texture, err = engine.Renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		engine.Width,
		engine.Height,
	)

	if err != nil {
		panic(err)
	}
}

func (engine *Engine) quitSDL() {
	engine.Texture.Destroy()
	engine.Renderer.Destroy()
	engine.Window.Destroy()

	sdl.Quit()
}
