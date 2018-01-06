package ui

import (
	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/vec2"
)

type Drawable interface {
	Draw(*Context2D)
}

type PostDrawable interface {
	PostDraw(*Context2D)
}

type Box struct {
	Top, Right, Bottom, Left float64
}

func (box Box) Contains(v vec2.Vec2) bool {
	return (v.Y > box.Top) &&
		(v.X > box.Left) &&
		(v.Y < box.Bottom) &&
		(v.X < box.Right)
}

func (box Box) Width() float64  { return box.Right - box.Left }
func (box Box) Height() float64 { return box.Bottom - box.Top }
func (box Box) Grow(x float64) Box {
	return Box{box.Top - x, box.Right + x, box.Bottom + x, box.Left - x}
}

type Sized interface {
	Size(TextMeasurer) Box
}

type Transformed interface {
	Transform(TextMeasurer) matrix.Matrix
}

func WalkAtPoint(m TextMeasurer, start interface{}, point vec2.Vec2, callback func(TreePath, vec2.Vec2) WalkAction) {
	Walk(start, func(path TreePath) WalkAction {
		localPoint := ToLocal(m, path, point)
		last := path[len(path)-1]
		if sized, ok := last.(Sized); ok {
			if !sized.Size(m).Contains(localPoint) {
				return Avoid
			}
		}
		return callback(path, localPoint)
	}, nil)
}

type TraversalContext interface {
	TextMeasurer
	EditStatus
}

func Draw(travCtx TraversalContext, root interface{}, ctx *Context2D) {
	Walk(root, func(path TreePath) WalkAction {
		//fmt.Println("Visiting", path[len(path)-1], "at depth", len(path))
		ctx.Save()
		if t, ok := path[len(path)-1].(Transformed); ok {
			ctx.Transform(t.Transform(travCtx))
		}
		if d, ok := path[len(path)-1].(Drawable); ok {
			d.Draw(ctx)
		}
		return Explore
	}, func(path TreePath) {
		last := path[len(path)-1]
		if d, ok := last.(PostDrawable); ok {
			d.PostDraw(ctx)
		}
		if editable, ok := last.(Editable); ok && travCtx.IsEditing(editable) {
			ctx.LineWidth(2)
			ctx.StrokeStyle("#f00")
			ctx.FillStyle("rgba(255,128,128,0.2)")
			ctx.BeginPath()
			box := editable.Size(travCtx)
			box.Left += 1
			box.Right -= 1
			box.Top += 1
			box.Bottom -= 1
			ctx.Rect2(box)
			ctx.Stroke()
			ctx.Fill()
		}
		ctx.Restore()
	})
}

func ToLocal(m TextMeasurer, path TreePath, v vec2.Vec2) vec2.Vec2 {
	return matrix.Apply(matrix.Invert(combineTransforms(m, path)), v)
}

func combineTransforms(m TextMeasurer, path TreePath) matrix.Matrix {
	combined := matrix.Identity()
	for i := 0; i < len(path); i++ {
		w := path[i]
		if t, ok := w.(Transformed); ok {
			combined = matrix.Multiply(t.Transform(m), combined)
		}
	}
	return combined
}

func IsVecOutside(vec, size vec2.Vec2) bool {
	return (vec.Y < -size.Y/2) ||
		(vec.X < -size.X/2) ||
		(vec.Y > size.Y/2) ||
		(vec.X > size.X/2)
}
