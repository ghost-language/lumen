package modules

import (
	"path/filepath"

	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
)

var ImageMethods = map[string]*object.LibraryFunction{}
var ImageProperties = map[string]*object.LibraryProperty{}

// imageConstructor loads an image relative to the game's source directory:
// new Image('resources/player.png').
func imageConstructor(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arity("Image", tok, args, 1); err != nil {
		return err
	}

	path, err := text("Image", tok, args, 0)

	if err != nil {
		return err
	}

	resolved := resolvePath(path)

	image, loadErr := engine.NewImage(resolved)

	if loadErr != nil {
		return engine.AssetFailure("Image", tok, path, resolved, loadErr)
	}

	return image
}

// spritesheetConstructor slices an image into equally sized frames:
// new Spritesheet('sheet.png', 16) for square frames, or
// new Spritesheet('sheet.png', 16, 24) for tall ones. The first argument may
// also be an image that is already loaded, so several sheets can be cut from
// one file at different frame sizes.
func spritesheetConstructor(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("Spritesheet", tok, args, 2, 3); err != nil {
		return err
	}

	var source *engine.Image

	switch given := args[0].(type) {
	case *object.String:
		resolved := resolvePath(given.Value)

		loaded, loadErr := engine.NewImage(resolved)

		if loadErr != nil {
			return engine.AssetFailure("Spritesheet", tok, given.Value, resolved, loadErr)
		}

		source = loaded
	case *engine.Image:
		source = given
	default:
		return engine.Mistyped("Spritesheet", tok, 0, "a path or an image", args[0])
	}

	frameWidth, err := integer("Spritesheet", tok, args, 1)

	if err != nil {
		return err
	}

	frameHeight := frameWidth

	if len(args) == 3 {
		given, heightErr := integer("Spritesheet", tok, args, 2)

		if heightErr != nil {
			return heightErr
		}

		frameHeight = given
	}

	sheet, sheetErr := engine.NewSpritesheet(source, int32(frameWidth), int32(frameHeight))

	if sheetErr != nil {
		return engine.Value("Spritesheet", tok, "%s", sheetErr).
			WithHelp("frames are cut from the top left of the image, so the frame size has to fit inside it")
	}

	return sheet
}

// animationConstructor builds an animation over a spritesheet's frames:
// new Animation(sheet, [0, 1, 2], 0.1, 'loop').
func animationConstructor(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if err := arityRange("Animation", tok, args, 3, 4); err != nil {
		return err
	}

	sheet, err := engine.SpritesheetArgument("Animation", tok, args, 0)

	if err != nil {
		return err
	}

	animation, animErr := engine.ParseAnimationArguments("Animation", tok, sheet, args)

	if animErr != nil {
		return animErr
	}

	return animation
}

// resolvePath turns a game-relative asset path into an absolute one. Absolute
// paths are left alone so a game can load from anywhere it likes.
func resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(engine.Lumen.Ghost.GetDirectory(), path)
}
