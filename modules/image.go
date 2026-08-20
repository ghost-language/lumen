package modules

import (
	"path/filepath"

	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
)

var ImageMethods = map[string]*object.LibraryFunction{}
var ImageProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(ImageMethods, "load", imageLoadMethod)
	modules.RegisterMethod(ImageMethods, "newQuad", imageNewQuadMethod)
}

// imageLoadMethod loads an image relative to the game's source directory.
func imageLoadMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("image.load", tok, args, 1); err != nil {
		return err
	}

	path, err := text("image.load", tok, args, 0)

	if err != nil {
		return err
	}

	image, loadErr := engine.NewImage(resolvePath(path))

	if loadErr != nil {
		return object.NewError("%d:%d: runtime error: image.load() could not load %s: %s", tok.Line, tok.Column, path, loadErr)
	}

	return image
}

// imageNewQuadMethod is an alias for canvas.newQuad(), kept here so spritesheet
// code that already reaches for the image module does not have to switch.
func imageNewQuadMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	return canvasNewQuadMethod(scope, tok, args...)
}

// resolvePath turns a game-relative asset path into an absolute one. Absolute
// paths are left alone so a game can load from anywhere it likes.
func resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(engine.Lumen.Ghost.GetDirectory(), path)
}
