package engine

import (
	"ghostlang.org/x/ghost/object"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const FONT = "Font"

// Font objects consist of a nil value.
type Font struct {
	Path   string
	Size   int
	Family *ttf.Font
}

func NewFont(path string, size int) *Font {
	font := new(Font)
	font.Path = path
	font.Size = size

	family, err := ttf.OpenFont(font.Path, font.Size)

	if err != nil {
		panic(err)
	}

	font.Family = family

	Lumen.RegisterResource(font.Path, font)

	return font
}

// String represents the font object's value as a string.
func (font *Font) String() string {
	return "Font: " + font.Path
}

// Type returns the font object type.
func (font *Font) Type() object.Type {
	return FONT
}

// Method defines the set of methods available on font objects.
func (font *Font) Method(method string, args []object.Object) (object.Object, bool) {
	switch method {
	case "print":
		font.print(args)
	}

	return nil, false
}

// =============================================================================
// Object methods

func (font *Font) print(args []object.Object) {
	text := args[0].String()
	x := int32(args[1].(*object.Number).Value.IntPart())
	y := int32(args[2].(*object.Number).Value.IntPart())

	surface, _ := Lumen.GetResource(font.Path).(*Font).Family.RenderUTF8Blended(text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	texture, _ := Lumen.Renderer.CreateTextureFromSurface(surface)

	rect := &sdl.Rect{X: x, Y: y, W: surface.W, H: surface.H}

	Lumen.Renderer.Copy(texture, nil, rect)

	surface.Free()
	texture.Destroy()
}
