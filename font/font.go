package font

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/ttf"
)

var DefaultFont *ttf.Font

func init() {
	var err error

	err = ttf.Init()

	if err != nil {
		fmt.Printf("Failed to initialize TTF: %s\n", err)
		os.Exit(1)
	}

	// Load default font
	fontPath := "../resources/fonts/silver.ttf"
	fontSize := 32

	if DefaultFont, err = ttf.OpenFont(fontPath, fontSize); err != nil {
		fmt.Printf("Failed to load font: %s\n", err)
		os.Exit(1)
	}

	// defer DefaultFont.Close()
}
