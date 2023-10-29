package lumen

import "github.com/veandco/go-sdl2/sdl"

func (lumen *Lumen) initSDL() {
	var err error

	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	lumen.initSDLWindow()
	lumen.initSDLRenderer()
	lumen.initSDLTexture()

	// Set the scaling quality
	// TODO: Make this configurable
	// 1 = linear filtering
	// 2 = anisotropic filtering
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "2")
}

func (lumen *Lumen) initSDLWindow() {
	var err error

	// Create the window
	lumen.Window, err = sdl.CreateWindow(
		lumen.Title,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		lumen.Width,
		lumen.Height,
		sdl.WINDOW_SHOWN,
	)

	if err != nil {
		panic(err)
	}
}

func (lumen *Lumen) initSDLRenderer() {
	var err error

	// Create the renderer
	lumen.Renderer, err = sdl.CreateRenderer(lumen.Window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)

	if err != nil {
		panic(err)
	}
}

func (lumen *Lumen) initSDLTexture() {
	var err error

	// Create the texture
	lumen.Texture, err = lumen.Renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		lumen.Width,
		lumen.Height,
	)

	if err != nil {
		panic(err)
	}
}

func (lumen *Lumen) quitSDL() {
	lumen.Texture.Destroy()
	lumen.Renderer.Destroy()
	lumen.Window.Destroy()

	sdl.Quit()
}
