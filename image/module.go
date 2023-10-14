package image

import (
	"fmt"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/environment"
	"ghostlang.org/x/lumen/renderer"
	"github.com/veandco/go-sdl2/img"
)

var ImageMethods = map[string]*object.LibraryFunction{}
var ImageProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(ImageMethods, "load", imageLoadFunction)
}

func imageLoadFunction(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		panic("wrong number of arguments. expected=1")
		// return object.NewError("wrong number of arguments. got=%d, expected=2", len(args))
	}

	// build image path using ghost.GetDirectory() to fetch the root and append imagePath
	imagePath := environment.Ghost.GetDirectory() + "/" + args[0].String()

	fmt.Println("imagePath:", imagePath)

	image := new(Image)

	image.blendmode = 1
	surface, err := img.Load(imagePath)

	if err != nil {
		return object.NewError("error loading image: %s", err)
	}

	image.surface = surface

	texture, err := renderer.Renderer.CreateTextureFromSurface(image.surface)

	fmt.Printf("renderer: %v\n", renderer.Renderer)

	if err != nil {
		return object.NewError("error creating texture: %s", err)
	}

	image.texture = texture
	image.width = image.surface.W
	image.height = image.surface.H

	return image
}
