package engine

import "math"

// Transform is a 2D affine transform stored as the two rows of a 2x3 matrix:
//
//	| A C TX |
//	| B D TY |
//
// Points map as (x, y) -> (A*x + C*y + TX, B*x + D*y + TY). Every coordinate a
// game passes to the canvas is pushed through the transform at the top of the
// stack before it reaches SDL, which is what makes cameras, zoom, and rotated
// scenes work uniformly across primitives, images, and text.
type Transform struct {
	A, B, C, D, TX, TY float64
}

// IdentityTransform returns a transform that leaves points untouched.
func IdentityTransform() Transform {
	return Transform{A: 1, B: 0, C: 0, D: 1, TX: 0, TY: 0}
}

// Multiply returns the composition t * other, meaning other is applied first.
func (t Transform) Multiply(other Transform) Transform {
	return Transform{
		A:  t.A*other.A + t.C*other.B,
		B:  t.B*other.A + t.D*other.B,
		C:  t.A*other.C + t.C*other.D,
		D:  t.B*other.C + t.D*other.D,
		TX: t.A*other.TX + t.C*other.TY + t.TX,
		TY: t.B*other.TX + t.D*other.TY + t.TY,
	}
}

// Translate returns the transform with a translation applied on the right.
func (t Transform) Translate(x, y float64) Transform {
	return t.Multiply(Transform{A: 1, D: 1, TX: x, TY: y})
}

// Scale returns the transform with a scale applied on the right.
func (t Transform) Scale(x, y float64) Transform {
	return t.Multiply(Transform{A: x, D: y})
}

// Rotate returns the transform with a rotation (in radians) applied on the right.
func (t Transform) Rotate(radians float64) Transform {
	sin, cos := math.Sincos(radians)

	return t.Multiply(Transform{A: cos, B: sin, C: -sin, D: cos})
}

// Shear returns the transform with a shear applied on the right.
func (t Transform) Shear(x, y float64) Transform {
	return t.Multiply(Transform{A: 1, B: y, C: x, D: 1})
}

// Apply maps a point through the transform.
func (t Transform) Apply(x, y float64) (float64, float64) {
	return t.A*x + t.C*y + t.TX, t.B*x + t.D*y + t.TY
}

// Determinant returns the determinant of the transform's linear part. A zero
// determinant means the transform collapses the plane and cannot be inverted.
func (t Transform) Determinant() float64 {
	return t.A*t.D - t.B*t.C
}

// Inverse returns the inverse transform. The second return value is false when
// the transform is singular, in which case the identity is returned.
func (t Transform) Inverse() (Transform, bool) {
	determinant := t.Determinant()

	if determinant == 0 {
		return IdentityTransform(), false
	}

	inverse := Transform{
		A: t.D / determinant,
		B: -t.B / determinant,
		C: -t.C / determinant,
		D: t.A / determinant,
	}

	inverse.TX = -(inverse.A*t.TX + inverse.C*t.TY)
	inverse.TY = -(inverse.B*t.TX + inverse.D*t.TY)

	return inverse, true
}

// ApproximateScale returns the average scale factor the transform applies. It is
// used to keep stroke widths and curve tessellation looking right under zoom.
func (t Transform) ApproximateScale() float64 {
	x := math.Hypot(t.A, t.B)
	y := math.Hypot(t.C, t.D)

	return (x + y) / 2
}
