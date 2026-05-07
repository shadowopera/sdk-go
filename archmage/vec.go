package archmage

// Vec2 represents a 2D vector with comparable values.
// It marshals to JSON as a two-element array.
type Vec2[T comparable] struct {
	X T `json:"x"`
	Y T `json:"y"`
}

// MakeVec2 creates a Vec2 from x and y.
func MakeVec2[T comparable](x, y T) Vec2[T] {
	return Vec2[T]{X: x, Y: y}
}

// Vec3 represents a 3D vector with comparable values.
// It marshals to JSON as a three-element array.
type Vec3[T comparable] struct {
	X T `json:"x"`
	Y T `json:"y"`
	Z T `json:"z"`
}

// MakeVec3 creates a Vec3 from x, y, and z.
func MakeVec3[T comparable](x, y, z T) Vec3[T] {
	return Vec3[T]{X: x, Y: y, Z: z}
}

// Vec4 represents a 4D vector with comparable values.
// It marshals to JSON as a four-element array.
type Vec4[T comparable] struct {
	X T `json:"x"`
	Y T `json:"y"`
	Z T `json:"z"`
	W T `json:"w"`
}

// MakeVec4 creates a Vec4 from x, y, z, and w.
func MakeVec4[T comparable](x, y, z, w T) Vec4[T] {
	return Vec4[T]{X: x, Y: y, Z: z, W: w}
}
