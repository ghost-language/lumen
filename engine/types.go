package engine

import "ghostlang.org/x/ghost/object"

// Ghost's object.Type is a compact integer enum owned by the interpreter. Host
// applications that introduce their own object types have to pick values that
// sit outside that enum, so Lumen's types start well above Ghost's highest
// constant. Ghost prints these as "UNKNOWN" in its own error messages, which is
// why every Lumen object gives String() a descriptive value instead.
const (
	IMAGE object.Type = 1000 + iota
	QUAD
	FONT
	COLOR
	TARGET
	SOURCE
	TRANSFORM
)

// TypeName returns the human readable name for a Lumen object type.
func TypeName(t object.Type) string {
	switch t {
	case IMAGE:
		return "Image"
	case QUAD:
		return "Quad"
	case FONT:
		return "Font"
	case COLOR:
		return "Color"
	case TARGET:
		return "Target"
	case SOURCE:
		return "Source"
	case TRANSFORM:
		return "Transform"
	}

	return t.String()
}
