package mvm

import (
	"math"

	. "github.com/mafik/mvm/vec2"
)

func Limit(a, limit float64) float64 {
	switch {
	case a < -limit:
		return -limit
	case a > limit:
		return limit
	default:
		return a
	}
}

func Ray(source, dest, box Vec2) Vec2 {
	ray := Sub(source, dest)
	ray.X = Limit(ray.X, box.X/2)
	ray.Y = Limit(ray.Y, box.Y/2)
	return Add(ray, dest)
}

func Dot(a, b Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}

func Clamp(min, max, val float64) float64 {
	switch {
	case val < min:
		return min
	case val > max:
		return max
	default:
		return val
	}
}

func Dist(a, b Vec2) float64 {
	d := Sub(a, b)
	return math.Hypot(d.X, d.Y)
}
