// Package resources holds the assets Lumen ships with. Embedding them means a
// built game is a single binary that can run from any working directory rather
// than one that has to find files relative to the executable.
package resources

import _ "embed"

// DefaultFont is the font Lumen draws text with until a game sets its own.
//
//go:embed silver.ttf
var DefaultFont []byte
