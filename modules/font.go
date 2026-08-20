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
	modules.RegisterMethod(FontMethods, "load", fontLoadMethod)
	modules.RegisterMethod(FontMethods, "system", fontSystemMethod)
}

// fontLoadMethod loads a TrueType font at a pixel size. Called with a size
// alone it returns Lumen's built-in font, which mirrors how LOVE's newFont()
// falls back to its own font when given no file.
func fontLoadMethod(scope *object.Scope, tok token.Token, args ...object.Object) object.Object {
	if len(args) == 1 {
		return fontSystemMethod(scope, tok, args...)
	}

	if err := arity("font.load", tok, args, 2); err != nil {
		return err
	}

	path, err := text("font.load", tok, args, 0)

	if err != nil {
		return err
	}

	size, err := integer("font.load", tok, args, 1)

	if err != nil {
		return err
	}

	font, loadErr := engine.NewFont(resolvePath(path), int(size))

	if loadErr != nil {
		return object.NewError("%d:%d: runtime error: font.load() could not load %s: %s", tok.Line, tok.Column, path, loadErr)
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

	font, loadErr := engine.NewDefaultFont(int(size))

	if loadErr != nil {
		return object.NewError("%d:%d: runtime error: font.system() %s", tok.Line, tok.Column, loadErr)
	}

	return font
}
