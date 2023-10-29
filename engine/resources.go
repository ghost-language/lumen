package engine

import "github.com/veandco/go-sdl2/sdl"

// Register a resource to the resource manager.
func (engine *Engine) RegisterResource(name string, resource interface{}) {
	engine.Resources[name] = resource
}

// Get a resource from the resource manager.
func (engine *Engine) GetResource(name string) interface{} {
	return engine.Resources[name]
}

// Free all resources registered in the resource manager.
func (engine *Engine) FreeResources() {
	for _, res := range engine.Resources {
		switch resource := res.(type) {
		case *sdl.Texture:
			resource.Destroy()
		case *sdl.Surface:
			resource.Free()
		}
	}
}
