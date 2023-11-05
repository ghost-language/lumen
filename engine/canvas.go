package engine

func (engine *Engine) ApplyOffset(x, y int32) (int32, int32) {
	return x + engine.OffsetX, y + engine.OffsetY
}
