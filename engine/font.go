package engine

import (
	"strings"

	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/value"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Font is a TrueType face at a fixed pixel size. Rendered strings are cached as
// textures because rasterising a glyph run every frame is far too slow for text
// that changes rarely, which is most of the text in a game.
type Font struct {
	Path       string
	Size       int
	LineHeight float64
	Family     *ttf.Font
	cache      map[textKey]*textTexture
}

// textKey identifies a cached run of rendered text.
type textKey struct {
	text  string
	wrap  int
	style int
}

// textTexture is one cached, rasterised string.
type textTexture struct {
	texture *sdl.Texture
	width   int32
	height  int32
	usedAt  uint64
}

const (
	// textCacheMaxAge is how many frames a cached string survives without being
	// drawn before it is evicted.
	textCacheMaxAge = 120

	// textCacheMaxEntries bounds the cache regardless of age. Text that changes
	// every frame — a clock, a score, a damage number — would otherwise add an
	// entry per frame and hold every one of them until it aged out.
	textCacheMaxEntries = 256
)

// defaultFontData holds the embedded font bytes, populated during SDL start-up.
var defaultFontData []byte

// NewFont loads a font from disk at the given pixel size.
func NewFont(path string, size int) (*Font, error) {
	family, err := ttf.OpenFont(path, size)

	if err != nil {
		return nil, err
	}

	font := newFontFromFamily(family, path, size)

	Lumen.RegisterResource(font)

	return font, nil
}

// NewFontFromMemory loads a font from an in-memory TrueType file, which is how
// the built-in default font is loaded.
func NewFontFromMemory(name string, data []byte, size int) (*Font, error) {
	source, err := sdl.RWFromMem(data)

	if err != nil {
		return nil, err
	}

	family, err := ttf.OpenFontRW(source, 1, size)

	if err != nil {
		return nil, err
	}

	font := newFontFromFamily(family, name, size)

	Lumen.RegisterResource(font)

	return font, nil
}

func newFontFromFamily(family *ttf.Font, path string, size int) *Font {
	return &Font{
		Path:       path,
		Size:       size,
		LineHeight: float64(family.LineSkip()),
		Family:     family,
		cache:      make(map[textKey]*textTexture),
	}
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
		return font.print(args)
	case "printf":
		return font.printf(args)
	case "getWidth":
		return font.getWidth(args)
	case "getHeight":
		return object.NewInt(int64(font.Family.Height())), true
	case "getLineHeight":
		return object.NewFloat(font.LineHeight), true
	case "setLineHeight":
		return font.setLineHeight(args)
	case "getAscent":
		return object.NewInt(int64(font.Family.Ascent())), true
	case "getDescent":
		return object.NewInt(int64(font.Family.Descent())), true
	case "getBaseline":
		return object.NewInt(int64(font.Family.Ascent())), true
	case "getWrap":
		return font.getWrap(args)
	case "getSize":
		return object.NewInt(int64(font.Size)), true
	case "toString":
		return &object.String{Value: font.String()}, true
	}

	return nil, false
}

// =============================================================================
// Object methods

// print draws a string at a position, taking the same trailing transform
// arguments as image.draw().
func (font *Font) print(args []object.Object) (object.Object, bool) {
	if len(args) < 3 {
		return object.NewError("font.print() expects at least a string, x, and y. got=%d", len(args)), true
	}

	arguments, err := ParseDrawArguments("font.print", args, 1)

	if err != nil {
		return err, true
	}

	font.drawText(args[0].String(), 0, arguments)

	return value.NULL, true
}

// printf draws wrapped text inside a given width, aligned left, center, right,
// or justified: font.printf(text, x, y, limit, align, rotation, sx, sy, ox, oy).
func (font *Font) printf(args []object.Object) (object.Object, bool) {
	if len(args) < 4 {
		return object.NewError("font.printf() expects at least a string, x, y, and wrap limit. got=%d", len(args)), true
	}

	limit, err := Float("font.printf", args, 3)

	if err != nil {
		return err, true
	}

	align := "left"
	offset := 4

	if len(args) > 4 {
		if name, ok := args[4].(*object.String); ok {
			align = name.Value
			offset = 5
		}
	}

	arguments := DrawArguments{ScaleX: 1, ScaleY: 1}

	x, err := Float("font.printf", args, 1)

	if err != nil {
		return err, true
	}

	y, err := Float("font.printf", args, 2)

	if err != nil {
		return err, true
	}

	arguments.X = x
	arguments.Y = y

	if len(args) > offset {
		trailing, err := ParseDrawArguments("font.printf", append([]object.Object{args[1], args[2]}, args[offset:]...), 0)

		if err != nil {
			return err, true
		}

		arguments = trailing
	}

	text := font.wrapText(args[0].String(), limit)
	lineHeight := font.LineHeight

	for index, line := range text {
		width, _, _ := font.Family.SizeUTF8(line)

		lineArguments := arguments
		lineArguments.OriginY = arguments.OriginY - float64(index)*lineHeight

		switch align {
		case "center":
			lineArguments.OriginX = arguments.OriginX - (limit-float64(width))/2
		case "right":
			lineArguments.OriginX = arguments.OriginX - (limit - float64(width))
		}

		font.drawText(line, 0, lineArguments)
	}

	return value.NULL, true
}

// getWidth returns the pixel width a string would occupy.
func (font *Font) getWidth(args []object.Object) (object.Object, bool) {
	if len(args) != 1 {
		return object.NewError("font.getWidth() expects 1 argument. got=%d", len(args)), true
	}

	width, _, err := font.Family.SizeUTF8(args[0].String())

	if err != nil {
		return object.NewError("font.getWidth() %s", err), true
	}

	return object.NewInt(int64(width)), true
}

// setLineHeight overrides the vertical distance between wrapped lines.
func (font *Font) setLineHeight(args []object.Object) (object.Object, bool) {
	height, err := Float("font.setLineHeight", args, 0)

	if err != nil {
		return err, true
	}

	font.LineHeight = height

	return value.NULL, true
}

// getWrap returns the width of the widest wrapped line and the list of lines a
// string breaks into at the given limit.
func (font *Font) getWrap(args []object.Object) (object.Object, bool) {
	if len(args) != 2 {
		return object.NewError("font.getWrap() expects 2 arguments. got=%d", len(args)), true
	}

	limit, err := Float("font.getWrap", args, 1)

	if err != nil {
		return err, true
	}

	lines := font.wrapText(args[0].String(), limit)
	elements := make([]object.Object, 0, len(lines))

	widest := 0

	for _, line := range lines {
		width, _, _ := font.Family.SizeUTF8(line)

		if width > widest {
			widest = width
		}

		elements = append(elements, &object.String{Value: line})
	}

	return &object.List{Elements: []object.Object{
		object.NewInt(int64(widest)),
		&object.List{Elements: elements},
	}}, true
}

// =============================================================================
// Helper methods

// wrapText breaks a string into lines no wider than limit, honouring any
// newlines already present in the text.
func (font *Font) wrapText(text string, limit float64) []string {
	lines := make([]string, 0, 4)

	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)

		if len(words) == 0 {
			lines = append(lines, "")

			continue
		}

		current := words[0]

		for _, word := range words[1:] {
			candidate := current + " " + word
			width, _, _ := font.Family.SizeUTF8(candidate)

			if float64(width) > limit {
				lines = append(lines, current)
				current = word

				continue
			}

			current = candidate
		}

		lines = append(lines, current)
	}

	return lines
}

// drawText renders one line through the current transform, tinted by the
// current draw color.
func (font *Font) drawText(text string, wrap int, arguments DrawArguments) {
	if text == "" {
		return
	}

	cached := font.texture(text, wrap)

	if cached == nil {
		return
	}

	Lumen.DrawTexture(cached.texture, nil, cached.width, cached.height,
		arguments.X, arguments.Y, arguments.Rotation,
		arguments.ScaleX, arguments.ScaleY, arguments.OriginX, arguments.OriginY)
}

// texture returns the cached texture for a string, rasterising it on first use.
// Text is always rendered white so the draw color can tint it, which means one
// cache entry serves every color the game draws that string in.
func (font *Font) texture(text string, wrap int) *textTexture {
	key := textKey{text: text, wrap: wrap, style: font.Family.GetStyle()}

	if cached, ok := font.cache[key]; ok {
		cached.usedAt = Lumen.FrameCount

		return cached
	}

	white := sdl.Color{R: 255, G: 255, B: 255, A: 255}

	var surface *sdl.Surface
	var err error

	if wrap > 0 {
		surface, err = font.Family.RenderUTF8BlendedWrapped(text, white, wrap)
	} else {
		surface, err = font.Family.RenderUTF8Blended(text, white)
	}

	if err != nil {
		return nil
	}

	defer surface.Free()

	texture, err := Lumen.Renderer.CreateTextureFromSurface(surface)

	if err != nil {
		return nil
	}

	if len(font.cache) >= textCacheMaxEntries {
		font.evict()
	}

	cached := &textTexture{texture: texture, width: surface.W, height: surface.H, usedAt: Lumen.FrameCount}
	font.cache[key] = cached

	return cached
}

// evict makes room in a full cache by dropping everything not drawn this frame.
// Strings still in use this frame are re-rasterised on their next draw at worst,
// so the cache cannot be filled past its bound by a single busy frame.
func (font *Font) evict() {
	for key, cached := range font.cache {
		if cached.usedAt == Lumen.FrameCount {
			continue
		}

		cached.texture.Destroy()

		delete(font.cache, key)
	}
}

// PruneCache drops rendered strings that have not been drawn recently. Without
// it, text that changes every frame (a clock, a score) would grow the cache
// without bound.
func (font *Font) PruneCache(frame uint64) {
	for key, cached := range font.cache {
		if frame-cached.usedAt > textCacheMaxAge {
			cached.texture.Destroy()

			delete(font.cache, key)
		}
	}
}

// Release frees the font and every string it has rasterised.
func (font *Font) Release() {
	for key, cached := range font.cache {
		cached.texture.Destroy()

		delete(font.cache, key)
	}

	if font.Family != nil {
		font.Family.Close()
		font.Family = nil
	}
}

// NewDefaultFont loads Lumen's built-in font at a given size, so a game can use
// it at more than one size without shipping a font file of its own.
func NewDefaultFont(size int) (*Font, error) {
	return NewFontFromMemory("lumen:silver.ttf", defaultFontData, size)
}
