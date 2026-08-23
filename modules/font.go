package modules

import (
	"ghostlang.org/x/ghost/library/modules"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/token"
	"ghostlang.org/x/lumen/engine"
)

var FontMethods = map[string]*object.LibraryFunction{}
var FontProperties = map[string]*object.LibraryProperty{}

func init() {
	modules.RegisterMethod(FontMethods, "system", fontSystemMethod)
}

// fontConstructor loads a TrueType font at a pixel size: new Font('font.ttf', 16).
// Called with a size alone, new Font(16), it returns Lumen's built-in font, so a
// game that only wants readable text at a chosen size never has to ship a font
// file to get one.
func fontConstructor(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) == 1 {
		return fontSystemMethod(scope, tok, args...)
	}

	if err := arity("Font", tok, args, 2); err != nil {
		return err
	}

	path, err := text("Font", tok, args, 0)

	if err != nil {
		return err
	}

	size, err := integer("Font", tok, args, 1)

	if err != nil {
		return err
	}

	if size <= 0 {
		return engine.Value("Font", tok, "was asked for a font %d pixels tall", size).
			WithHelp("a font size is measured in pixels, so it has to be above zero")
	}

	resolved := resolvePath(path)

	font, loadErr := engine.NewFont(resolved, int(size))

	if loadErr != nil {
		return engine.AssetFailure("Font", tok, path, resolved, loadErr)
	}

	return font
}

// fontSystemMethod returns the font Lumen ships with, at the requested size.
// Called with no size it returns the font the canvas is already using.
func fontSystemMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) == 0 {
		return engine.Lumen.DefaultFont
	}

	size, err := integer("font.system", tok, args, 0)

	if err != nil {
		return err
	}

	if size <= 0 {
		return engine.Value("font.system", tok, "was asked for a font %d pixels tall", size).
			WithHelp("a font size is measured in pixels, so it has to be above zero")
	}

	font, loadErr := engine.NewDefaultFont(int(size))

	if loadErr != nil {
		return engine.SystemFailure("font.system", tok, loadErr)
	}

	return font
}
