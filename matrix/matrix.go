package matrix

import (
	"math"

	"github.com/mafik/mvm/vec2"
)

// 3x3 matrix with an implicit 0,0,1 row at the bottom
// [ #0 #2 #4 ]
// [ #1 #3 #5 ]
// [  0  0  1 ]
type Matrix [6]float64

func Identity() Matrix {
	return Matrix{1, 0, 0, 1, 0, 0}
}

func Translate(t vec2.Vec2) Matrix {
	return Matrix{1, 0, 0, 1, t.X, t.Y}
}

func (m *Matrix) Translate(t vec2.Vec2) *Matrix {
	*m = Multiply(*m, Translate(t))
	return m
}

func Scale(s float64) Matrix {
	return Matrix{s, 0, 0, s, 0, 0}
}

func (m *Matrix) Scale(s float64) *Matrix {
	*m = Multiply(Scale(s), *m)
	return m
}

func Rotate(a float64) Matrix {
	s, c := math.Sin(a), math.Cos(a)
	return Matrix{c, s, -s, c, 0, 0}
}

func Determinant(m Matrix) float64 {
	return m[0]*m[3] - m[1]*m[2]
}

func Invert(m Matrix) Matrix {
	d := Determinant(m)
	return Matrix{
		m[3] / d, -m[1] / d,
		-m[2] / d, m[0] / d,
		(m[2]*m[5] - m[3]*m[4]) / d,
		(m[1]*m[4] - m[0]*m[5]) / d,
	}

}

func Multiply(a, b Matrix) Matrix {
	return Matrix{
		a[0]*b[0] + a[1]*b[2],
		a[1]*b[3] + a[0]*b[1],
		a[2]*b[0] + a[3]*b[2],
		a[3]*b[3] + a[2]*b[1],
		a[4]*b[0] + a[5]*b[2] + b[4],
		a[5]*b[3] + a[4]*b[1] + b[5],
	}
}

func Apply(m Matrix, p vec2.Vec2) vec2.Vec2 {
	return vec2.Vec2{
		p.X*m[0] + p.Y*m[2] + m[4],
		p.X*m[1] + p.Y*m[3] + m[5],
	}
}
