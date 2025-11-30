package archmage

type Vec2[T any] struct {
	X, Y T
}

func MakeVec2[T any](x, y T) Vec2[T] {
	return Vec2[T]{X: x, Y: y}
}

type Vec3[T any] struct {
	X, Y, Z T
}

func MakeVec3[T any](x, y, z T) Vec3[T] {
	return Vec3[T]{X: x, Y: y, Z: z}
}

type Vec4[T any] struct {
	X, Y, Z, W T
}

func MakeVec4[T any](x, y, z, w T) Vec4[T] {
	return Vec4[T]{X: x, Y: y, Z: z, W: w}
}
