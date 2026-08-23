package engine

import "ghostlang.org/x/ghost/object"

// load runs the game's one-time setup. A failure here is reported by the
// callback itself, and stops the game before the first frame is drawn.
func (engine *Engine) load() {
	engine.callback("load")
}

// update advances the game by one frame, handing it the seconds elapsed since
// the previous frame. Games that ignore the argument still work; Ghost drops
// arguments a function does not declare.
func (engine *Engine) update() {
	engine.callback("update", object.NewFloat(engine.Delta))
}

// draw renders one frame. The transform stack is reset first so a game that
// forgets to pop a camera transform does not compound it every frame.
func (engine *Engine) draw() {
	engine.Graphics.Reset()

	engine.SetTarget(nil)
	engine.ClearScissor()

	background := engine.Graphics.BackgroundColor

	engine.Renderer.SetDrawColor(background.Red, background.Green, background.Blue, background.Alpha)
	engine.Renderer.Clear()

	engine.callback("draw")

	engine.SetTarget(nil)

	// Everything the game queued has to reach the frame before the letterbox
	// bars are painted over the edges of it and the frame is presented.
	engine.Flush()

	engine.Renderer.SetClipRect(nil)

	engine.drawLetterbox()

	engine.Renderer.Present()
}
