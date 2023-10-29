package window

import "github.com/veandco/go-sdl2/sdl"

var Window *window

type window struct {
	Window *sdl.Window
	Title  string
	Width  int32
	Height int32
	Fps    float64
}

func New(title string, width int32, height int32) *window {
	var err error

	Window = new(window)

	Window.Title = title
	Window.Width = width
	Window.Height = height

	Window.Window, err = sdl.CreateWindow(
		Window.Title,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		Window.Width,
		Window.Height,
		sdl.WINDOW_SHOWN,
	)

	if err != nil {
		panic(err)
	}

	return Window
}
