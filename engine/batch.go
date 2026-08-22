package engine

import "github.com/veandco/go-sdl2/sdl"

// batchVertexLimit caps how many vertices accumulate before the batch is handed
// to SDL. Drivers are happiest with draws in the low thousands of vertices, and
// the cap keeps the scratch buffers from growing without bound on a frame that
// draws an unusually large world.
const batchVertexLimit = 16384

// batch accumulates textured quads that share render state so SDL sees one
// RenderGeometry call instead of one per sprite.
//
// The batch is ordered, not sorted: quads are only merged with the ones
// immediately before them, and any change of texture, blend mode, or scissor
// box ends the run. Sorting by texture would collapse more draws into fewer
// calls, but it would also reorder overlapping sprites, and a tilemap drawn
// back-to-front depends on that order. Correct output first.
type batch struct {
	texture    *sdl.Texture
	blendMode  sdl.BlendMode
	scissor    sdl.Rect
	hasScissor bool
	active     bool
	vertices   []sdl.Vertex
	indices    []int32
}

// compatible reports whether a quad drawn with the given state can join the
// batch as it stands.
func (batch *batch) compatible(texture *sdl.Texture, blendMode sdl.BlendMode, scissor *sdl.Rect) bool {
	if !batch.active {
		return false
	}

	if batch.texture != texture || batch.blendMode != blendMode {
		return false
	}

	if (scissor != nil) != batch.hasScissor {
		return false
	}

	return scissor == nil || *scissor == batch.scissor
}

// begin points an empty batch at a new render state.
func (batch *batch) begin(texture *sdl.Texture, blendMode sdl.BlendMode, scissor *sdl.Rect) {
	batch.texture = texture
	batch.blendMode = blendMode
	batch.hasScissor = scissor != nil
	batch.active = true

	if scissor != nil {
		batch.scissor = *scissor
	}
}

// append adds one quad's four corners and the two triangles that cover them.
func (batch *batch) append(corners *[4]sdl.Vertex) {
	base := int32(len(batch.vertices))

	batch.vertices = append(batch.vertices, corners[0], corners[1], corners[2], corners[3])
	batch.indices = append(batch.indices, base, base+1, base+2, base, base+2, base+3)
}

// reset empties the batch while keeping the buffers it has already grown, which
// is what makes drawing a frame allocation-free after the first one.
func (batch *batch) reset() {
	batch.vertices = batch.vertices[:0]
	batch.indices = batch.indices[:0]
	batch.texture = nil
	batch.active = false
	batch.hasScissor = false
}

// =============================================================================
// Engine helpers

// batchQuad queues a textured quad, flushing first if it cannot join the run
// already in progress.
func (engine *Engine) batchQuad(texture *sdl.Texture, blendMode sdl.BlendMode, scissor *sdl.Rect, corners *[4]sdl.Vertex) {
	if !engine.batch.compatible(texture, blendMode, scissor) {
		engine.Flush()
		engine.batch.begin(texture, blendMode, scissor)
	}

	engine.batch.append(corners)

	if len(engine.batch.vertices) >= batchVertexLimit {
		engine.Flush()
	}
}

// Flush draws everything queued in the sprite batch and empties it.
//
// Anything that talks to the renderer directly has to call this first, or its
// output lands underneath sprites that were queued before it. That covers the
// obvious cases — presenting the frame, clearing, switching render target — and
// the less obvious ones: reading pixels back, and destroying a texture the
// pending batch still points at.
func (engine *Engine) Flush() {
	if len(engine.batch.indices) == 0 {
		engine.batch.reset()

		return
	}

	var scissor *sdl.Rect

	if engine.batch.hasScissor {
		scissor = &engine.batch.scissor
	}

	engine.Renderer.SetDrawBlendMode(engine.batch.blendMode)
	engine.Renderer.SetClipRect(scissor)

	if engine.batch.texture != nil {
		engine.batch.texture.SetBlendMode(engine.batch.blendMode)
	}

	engine.Renderer.RenderGeometry(engine.batch.texture, engine.batch.vertices, engine.batch.indices)

	engine.batch.reset()
}

// FlushTexture flushes the batch only when it is holding draws for the given
// texture. Callers about to destroy or replace a texture use it so freeing an
// image mid-frame cannot leave the batch pointing at memory SDL has released.
func (engine *Engine) FlushTexture(texture *sdl.Texture) {
	if texture == nil || engine.batch.texture != texture {
		return
	}

	engine.Flush()
}
