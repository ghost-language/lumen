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
	modules.RegisterMethod(ImageMethods, "newSpritesheet", imageNewSpritesheetMethod)
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

// imageNewSpritesheetMethod slices an image into equally sized frames:
// image.newSpritesheet('sheet.png', 16) for square frames, or
// image.newSpritesheet('sheet.png', 16, 24) for tall ones. The first argument
// may also be an image that is already loaded, so several sheets can be cut
// from one file at different frame sizes.
func imageNewSpritesheetMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("image.newSpritesheet", tok, args, 2, 3); err != nil {
		return err
	}

	var source *engine.Image

	switch given := args[0].(type) {
	case *object.String:
		loaded, loadErr := engine.NewImage(resolvePath(given.Value))

		if loadErr != nil {
			return object.NewError("%d:%d: runtime error: image.newSpritesheet() could not load %s: %s", tok.Line, tok.Column, given.Value, loadErr)
		}

		source = loaded
	case *engine.Image:
		source = given
	default:
		return object.NewError("%d:%d: runtime error: image.newSpritesheet() expects a path or an image. got=%s", tok.Line, tok.Column, engine.TypeName(args[0].Type()))
	}

	frameWidth, err := integer("image.newSpritesheet", tok, args, 1)

	if err != nil {
		return err
	}

	frameHeight := frameWidth

	if len(args) == 3 {
		given, heightErr := integer("image.newSpritesheet", tok, args, 2)

		if heightErr != nil {
			return heightErr
		}

		frameHeight = given
	}

	sheet, sheetErr := engine.NewSpritesheet(source, int32(frameWidth), int32(frameHeight))

	if sheetErr != nil {
		return object.NewError("%d:%d: runtime error: image.newSpritesheet() %s", tok.Line, tok.Column, sheetErr)
	}

	return sheet
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
