package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
	"github.com/veandco/go-sdl2/img"
)

var ImageMethods = map[string]*object.LibraryFunction{}
var ImageProperties = map[string]*object.LibraryProperty{}

func init() {
	// Methods
	modules.RegisterMethod(ImageMethods, "load", imageLoadMethod)

	// Properties
	// modules.RegisterProperty(ImageProperties, "fps", windowFpsProperty)
}

func imageLoadMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1", len(args))
	}

	image := new(engine.Image)

	image.Path = engine.Lumen.Ghost.GetDirectory() + "/" + args[0].String()
	image.BlendMode = 1
	surface, err := img.Load(image.Path)

	if err != nil {
		return object.NewError("could not load image: %s", err.Error())
	}

	image.Surface = surface

	texture, err := engine.Lumen.Renderer.CreateTextureFromSurface(image.Surface)

	if err != nil {
		return object.NewError("could not create texture from surface: %s", err.Error())
	}

	image.Texture = texture
	image.Width = image.Surface.W
	image.Height = image.Surface.H

	// Add image reference to resource manager
	engine.Lumen.RegisterResource(image.Path, image)

	return image
}

// func windowFpsProperty(scope *object.Scope, tok token.Token) object.Object {
// 	return &object.Number{Value: decimal.NewFromInt(int64(engine.Lumen.CurrentFps))}
// }
