package engine

import (
	"fmt"
	"strings"

	"ghostlang.org/x/ghost/fault"
	"ghostlang.org/x/ghost/source"
	"github.com/veandco/go-sdl2/sdl"
)

// The error screen is the console report, drawn where the reader is looking.
//
// It says the same things in the same order — what kind of failure it is, what
// went wrong, where, the line it happened on with the offending part marked,
// what was in flight at the time, and what to do about it — because a reader
// who has learned to read one of them has learned to read both. What it does
// not do is imitate the console's typography: a caret row lines up under a
// character only in a monospaced font, and Lumen's built-in font is not one, so
// the offending lexeme is marked by measuring it and drawing a bar under it.
//
// Nothing here goes through the drawing stack a game uses. The transform, the
// batch, the scissor, and whatever render target was set are all part of the
// state that has just been shown not to be trustworthy, so the screen is drawn
// straight onto the renderer with plain copies.

// errorScreen holds the colors the report is laid out with. They are
// deliberately not the game's: an error screen that could be mistaken for a
// frame of the game is an error screen that gets missed.
var errorScreen = struct {
	background sdl.Color
	panel      sdl.Color
	heading    sdl.Color
	message    sdl.Color
	location   sdl.Color
	gutter     sdl.Color
	snippet    sdl.Color
	marker     sdl.Color
	note       sdl.Color
	help       sdl.Color
	footer     sdl.Color
	rule       sdl.Color
}{
	background: sdl.Color{R: 24, G: 22, B: 37, A: 255},
	panel:      sdl.Color{R: 34, G: 31, B: 52, A: 255},
	heading:    sdl.Color{R: 255, G: 106, B: 106, A: 255},
	message:    sdl.Color{R: 238, G: 238, B: 245, A: 255},
	location:   sdl.Color{R: 130, G: 170, B: 255, A: 255},
	gutter:     sdl.Color{R: 110, G: 110, B: 140, A: 255},
	snippet:    sdl.Color{R: 220, G: 220, B: 232, A: 255},
	marker:     sdl.Color{R: 255, G: 106, B: 106, A: 255},
	note:       sdl.Color{R: 150, G: 150, B: 175, A: 255},
	help:       sdl.Color{R: 120, G: 220, B: 210, A: 255},
	footer:     sdl.Color{R: 140, G: 140, B: 165, A: 255},
	rule:       sdl.Color{R: 70, G: 66, B: 96, A: 255},
}

// headingSize is how much taller the heading is than the body. A report that is
// all one size is a wall of text, and the first thing a reader needs from it is
// the two words saying what sort of failure this is.
const headingSize = 1.6

// drawReport paints the error screen for the failure the game stopped on.
func (engine *Engine) drawReport() {
	if engine.report == nil {
		return
	}

	font := engine.DefaultFont

	if font == nil || font.Family == nil {
		// Without a font there is nothing to say on screen, and the console has
		// the report either way. A flat red frame is still worth painting: it
		// says the game is not running, which a frozen frame does not.
		engine.Renderer.SetDrawColor(errorScreen.heading.R, errorScreen.heading.G, errorScreen.heading.B, 255)
		engine.Renderer.Clear()
		engine.Renderer.Present()

		return
	}

	width, height := engine.OutputSize()
	raised := engine.report.Fault

	// A game drawing into a 320x240 window gets the same report as one filling
	// a monitor, so the spacing is measured from the window rather than fixed.
	margin := clampInt32(width/24, 10, 32)
	line := int32(font.Family.Height())
	spacing := line / 3

	// The screen is drawn from a known state rather than from whatever the game
	// had set when it stopped: half a frame of a game that is not running any
	// more, and a target, scissor, or transform that are part of the state just
	// shown not to be trustworthy.
	engine.Graphics.Reset()
	engine.SetTarget(nil)
	engine.ClearScissor()

	engine.Renderer.SetClipRect(nil)
	engine.Renderer.SetDrawColor(errorScreen.background.R, errorScreen.background.G, errorScreen.background.B, 255)
	engine.Renderer.Clear()

	cursor := engine.drawReportHeading(font, raised, width, margin)
	footer := height - margin - line

	limit := float64(width - margin*2)

	// Every write below stops at the footer, so a long message runs out of room
	// rather than running over the line that says how to get out of here.
	write := func(text string, color sdl.Color) {
		if cursor+line > footer {
			return
		}

		engine.drawReportText(font, text, margin, cursor, color)

		cursor += line
	}

	wrap := func(text string, color sdl.Color) {
		for _, wrapped := range font.wrapText(text, limit) {
			write(wrapped, color)
		}
	}

	wrap(raised.Message, errorScreen.message)

	if raised.Position.Known() {
		cursor += spacing

		write(raised.Position.String(), errorScreen.location)

		engine.drawSnippet(font, raised, margin, &cursor, footer)
	}

	cursor += spacing

	for _, frame := range raised.Trace {
		wrap(describeFrame(frame), errorScreen.note)
	}

	if raised.Hidden > 0 {
		wrap(fmt.Sprintf("and %d more calls", raised.Hidden), errorScreen.note)
	}

	if raised.Help != "" {
		cursor += spacing

		wrap("help: "+firstParagraph(raised.Help), errorScreen.help)
	}

	engine.drawReportFooter(font, width, height, margin)

	engine.Renderer.Present()
}

// drawReportHeading paints the bar across the top and returns the line the
// report proper starts on.
func (engine *Engine) drawReportHeading(font *Font, raised *fault.Fault, width int32, margin int32) int32 {
	heading := engine.headingFont()
	line := int32(heading.Family.Height())
	bar := line + margin*2

	engine.fillReportRect(0, 0, width, bar, errorScreen.panel)
	engine.fillReportRect(0, bar-2, width, 2, errorScreen.heading)

	engine.drawReportText(heading, raised.Kind.String(), margin, margin, errorScreen.heading)

	// The engine's name sits opposite the heading, because a window that has
	// stopped saying anything about the game should still say what stopped it.
	if stamp := font.texture("lumen", 0); stamp != nil && width-margin-stamp.width > margin*8 {
		engine.drawReportText(font, "lumen", width-margin-stamp.width, margin+line-int32(font.Family.Height()), errorScreen.gutter)
	}

	return bar + margin
}

// headingFont is the built-in font at heading size, loaded the first time a
// report is drawn. A game that never fails never pays for it.
func (engine *Engine) headingFont() *Font {
	if engine.reportFont != nil {
		return engine.reportFont
	}

	font, err := NewDefaultFont(int(float64(engine.DefaultFont.Size) * headingSize))

	if err != nil {
		return engine.DefaultFont
	}

	engine.reportFont = font

	return font
}

// drawSnippet quotes the line the failure happened on and marks the part of it
// that failed, with a bar under the lexeme rather than a row of carets: the bar
// is measured in the same font the line is drawn in, so it lands under the
// right characters whatever their widths.
func (engine *Engine) drawSnippet(font *Font, raised *fault.Fault, margin int32, cursor *int32, footer int32) {
	text, ok := source.Line(raised.Position.File, raised.Position.Line)

	if !ok {
		return
	}

	line := int32(font.Family.Height())
	spacing := line / 3

	if *cursor+line*2+spacing > footer {
		return
	}

	*cursor += spacing

	number := fmt.Sprintf("%d", raised.Position.Line)
	gutter, _, err := font.Family.SizeUTF8(number + "   ")

	if err != nil {
		return
	}

	// Tabs are expanded before anything is measured. A tab drawn as a glyph is
	// a box of unknown width, and everything to the right of it — including the
	// bar — would be placed against a line the reader is not seeing.
	text = strings.ReplaceAll(text, "\t", "    ")

	left := margin + int32(gutter)

	// A rule down the left of the quoted line, where the console report has its
	// gutter pipe. It is what makes the line read as quoted source rather than
	// as another sentence of the message.
	engine.fillReportRect(margin-spacing, *cursor, 2, line, errorScreen.rule)

	engine.drawReportText(font, number, margin+spacing, *cursor, errorScreen.gutter)
	engine.drawReportText(font, text, left, *cursor, errorScreen.snippet)

	start, end := markedSpan(text, raised.Position)

	before, _, beforeErr := font.Family.SizeUTF8(text[:start])
	marked, _, markedErr := font.Family.SizeUTF8(text[start:end])

	if beforeErr == nil && markedErr == nil && marked > 0 {
		engine.fillReportRect(
			left+int32(before),
			*cursor+line-3,
			int32(marked),
			2,
			errorScreen.marker,
		)
	}

	*cursor += line + spacing
}

// markedSpan converts a position's column and width, which are counted in
// characters, into the byte offsets the text has to be sliced at to measure it.
func markedSpan(text string, position fault.Position) (int, int) {
	runes := []rune(text)

	start := position.Column - 1

	if start < 0 {
		start = 0
	}

	if start > len(runes) {
		start = len(runes)
	}

	length := position.Length

	if length < 1 {
		length = 1
	}

	end := start + length

	if end > len(runes) {
		end = len(runes)
	}

	return len(string(runes[:start])), len(string(runes[:end]))
}

// drawReportFooter draws the bar saying what the reader can do about this.
func (engine *Engine) drawReportFooter(font *Font, width int32, height int32, margin int32) {
	controls := []string{"esc  quit"}

	if engine.report.Resumable {
		controls = append(controls, "enter  continue")
	}

	controls = append(controls, "c  copy")

	line := int32(font.Family.Height())
	bar := line + margin

	engine.fillReportRect(0, height-bar, width, bar, errorScreen.panel)
	engine.fillReportRect(0, height-bar, width, 1, errorScreen.rule)
	engine.drawReportText(font, strings.Join(controls, "     ·     "), margin, height-bar+margin/2, errorScreen.footer)

	// Which callback this came out of belongs on this bar rather than in the
	// report: it is the one part of a report that is about Lumen's own doing
	// rather than about the game's code.
	if engine.report.Where == "" {
		return
	}

	label := "in " + engine.report.Where

	if stamp := font.texture(label, 0); stamp != nil && width-margin-stamp.width > width/2 {
		engine.drawReportText(font, label, width-margin-stamp.width, height-bar+margin/2, errorScreen.gutter)
	}
}

// drawReportText draws one line of the report at a window position.
//
// The font's cache renders text white so that a game can tint it; the same
// applies here, and the tint is put back to white afterwards so a cached string
// the game also draws is not left stained by the error screen.
func (engine *Engine) drawReportText(font *Font, text string, x int32, y int32, color sdl.Color) {
	if text == "" {
		return
	}

	cached := font.texture(text, 0)

	if cached == nil {
		return
	}

	cached.texture.SetColorMod(color.R, color.G, color.B)
	cached.texture.SetAlphaMod(color.A)

	engine.Renderer.Copy(cached.texture, nil, &sdl.Rect{X: x, Y: y, W: cached.width, H: cached.height})

	cached.texture.SetColorMod(255, 255, 255)
	cached.texture.SetAlphaMod(255)
}

// fillReportRect fills a rectangle in window pixels.
func (engine *Engine) fillReportRect(x int32, y int32, width int32, height int32, color sdl.Color) {
	engine.Renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	engine.Renderer.FillRect(&sdl.Rect{X: x, Y: y, W: width, H: height})
}

// describeFrame renders a call frame the way the console renders it, so the two
// reports read the same.
func describeFrame(frame fault.Frame) string {
	name := frame.Name

	if name == "" {
		name = "an anonymous function"
	}

	if !frame.Position.Known() {
		return "in " + name
	}

	return fmt.Sprintf("in %s, called at %s", name, frame.Position)
}

// firstParagraph trims a help line down to what fits on a screen. The only help
// that runs long is the internal one carrying a Go stack, which belongs in the
// console it was written to and not over the top of the report.
func firstParagraph(help string) string {
	if index := strings.Index(help, "\n"); index >= 0 {
		return strings.TrimSpace(help[:index])
	}

	return help
}

// handleReportKey answers the keys the error screen offers. It is reached only
// while the game is halted, which is what keeps these from being swallowed by a
// game that has bound them.
func (engine *Engine) handleReportKey(event *sdl.KeyboardEvent) {
	if event.Type != sdl.KEYDOWN {
		return
	}

	switch event.Keysym.Sym {
	case sdl.K_ESCAPE, sdl.K_q:
		engine.Quit()
	case sdl.K_RETURN, sdl.K_KP_ENTER, sdl.K_SPACE:
		engine.Resume()
	case sdl.K_c:
		engine.Copy()
	}
}

// clampInt32 keeps a measured size inside sensible bounds.
func clampInt32(value int32, low int32, high int32) int32 {
	if value < low {
		return low
	}

	if value > high {
		return high
	}

	return value
}
