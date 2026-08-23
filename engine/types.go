package engine

import "ghostlang.org/x/ghost/object"

// Ghost's object.Type is a compact integer enum owned by the interpreter. Host
// applications that introduce their own object types have to pick values that
// sit outside that enum, so Lumen's types start well above Ghost's highest
// constant. Ghost has no name for anything up here and calls it "unknown",
// which is why every message Lumen writes names a value through TypeName in
// errors.go rather than through Ghost's, and why every Lumen object gives
// String() a descriptive value.
const (
	IMAGE object.Type = 1000 + iota
	QUAD
	FONT
	COLOR
	TARGET
	SOURCE
	TRANSFORM
	SPRITESHEET
	ANIMATION
)
