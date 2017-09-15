package mvm

type Vec2 struct {
	X, Y float64
}

func Add(a, b Vec2) Vec2 {
	return Vec2{a.X + b.X, a.Y + b.Y}
}

func (a *Vec2) Add(b Vec2) *Vec2 {
	a.X += b.X
	a.Y += b.Y
	return a
}

func Neg(a Vec2) Vec2 {
	return Vec2{-a.X, -a.Y}
}

func Sub(a, b Vec2) Vec2 {
	return Vec2{a.X - b.X, a.Y - b.Y}
}

func (a *Vec2) Sub(b Vec2) *Vec2 {
	a.X -= b.X
	a.Y -= b.Y
	return a
}

func Div(a, b Vec2) Vec2 {
	return Vec2{a.X / b.X, a.Y / b.Y}
}

func (a *Vec2) Div(b Vec2) *Vec2 {
	a.X /= b.X
	a.Y /= b.Y
	return a
}

func Mul(a, b Vec2) Vec2 {
	return Vec2{a.X * b.X, a.Y * b.Y}
}

func (a *Vec2) Mul(b Vec2) *Vec2 {
	a.X *= b.X
	a.Y *= b.Y
	return a
}

func Scale(a Vec2, b float64) Vec2 {
	return Vec2{a.X * b, a.Y * b}
}

func (a *Vec2) Scale(b float64) *Vec2 {
	a.X *= b
	a.Y *= b
	return a
}
