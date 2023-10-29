package engine

func (engine *Engine) load() {
	engine.Ghost.Call("load", nil)
}

func (engine *Engine) update() {
	engine.Ghost.Call("update", nil)
}

func (engine *Engine) draw() {
	engine.Renderer.SetDrawColor(0, 0, 0, 255)
	engine.Renderer.Clear()

	engine.Ghost.Call("draw", nil)

	engine.Renderer.Present()
}
