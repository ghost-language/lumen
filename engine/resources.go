package engine

// Releasable is anything holding SDL memory that has to be handed back before
// the process exits. Every Lumen object that owns a texture, font, or decoded
// sound implements it.
type Releasable interface {
	Release()
}

// RegisterResource tracks a resource so it is freed when the engine shuts down.
func (engine *Engine) RegisterResource(resource Releasable) {
	engine.Resources = append(engine.Resources, resource)
}

// FreeResources releases every tracked resource. Resources are released in
// reverse order of registration so views are dropped before what they point at.
func (engine *Engine) FreeResources() {
	for index := len(engine.Resources) - 1; index >= 0; index-- {
		engine.Resources[index].Release()
	}

	engine.Resources = nil
}

// PruneCaches drops cached data that has gone unused, keeping memory flat for
// games that draw text which changes every frame.
func (engine *Engine) PruneCaches() {
	for _, resource := range engine.Resources {
		if font, ok := resource.(*Font); ok {
			font.PruneCache(engine.FrameCount)
		}
	}
}
