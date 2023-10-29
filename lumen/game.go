package lumen

func (lumen *Lumen) load() {
	lumen.Ghost.Call("load", nil)
}

func (lumen *Lumen) update() {
	lumen.Ghost.Call("update", nil)
}

func (lumen *Lumen) draw() {
	lumen.Renderer.SetDrawColor(0, 0, 0, 255)
	lumen.Renderer.Clear()

	lumen.Ghost.Call("draw", nil)

	lumen.Renderer.Present()
}
