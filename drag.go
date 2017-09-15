package mvm

import "fmt"

// On-screen element that can be dragged around.
type Draggable interface {
	Drag(*Touch) Touching
}

func (c *Container) Drag(touch *Touch) Touching {
	for _, elem := range c.elements {
		if drag, ok := elem.(Draggable); ok {
			if t := drag.Drag(touch); t != nil {
				return t
			}
		} else {
			fmt.Println("Warning: found non-Draggable layer")
		}
	}
	return nil
}

// Links & Params

var LinkTargetType Type = &PrimitiveType{
	name: "",
	instantiate: func(me *Object) {
		me.frame.size = Vec2{0, 0}
	},
}

func (b *Blueprint) MakeLinkTarget() *Frame {
	return b.Add(LinkTargetType)
}

func (ParamLayer) Drag(t *Touch) Touching {
	frame, i := t.PointedParam()
	if frame == nil {
		return nil
	}
	blue := frame.blueprint
	link := &Link{blue, frame, blue.MakeLinkTarget(), i, NextOrder(frame, i)}
	link.B.pos = t.Global
	blue.links[link] = true
	return link
}

func (LinkLayer) Drag(t *Touch) Touching {
	link := t.PointedLink()
	if link == nil {
		return nil
	}
	blue := link.Blueprint
	link.B = blue.MakeLinkTarget()
	link.B.pos = t.Global
	return link
}

func (l *Link) Move(touch *Touch) {
	l.B.pos = touch.Global
}

/* // TODO: Convert to Drawable
func (drag *LinkDrag) DragHighlight(widgets *Widgets) {
	link := drag.link
	param := link.A.typ.Parameters[link.Param]
	for frame, _ := range link.Blueprint.frames {
		ok := param.Typ == nil || frame.typ == param.Typ
		if param.Runnable && frame.typ.Run == nil {
			ok = false
		}
		color := "rgba(0,255,0,0.2)"
		if !ok {
			color = "rgba(255,0,0,0.2)"
		}
		widgets.Rect(frame.pos, frame.size, color)
	}
}
*/

func (l *Link) End(touch *Touch) {
	frame := touch.PointedFrame()
	if frame == nil {
		l.B.Delete()
	} else {
		fake_target := l.B
		l.B = frame
		fake_target.Delete()
	}
}

// Frames

func (FrameLayer) Drag(t *Touch) Touching {
	frame := t.PointedFrame()
	if frame == nil {
		return nil
	}
	low, high := -1./6., 1./6.
	frac := Div(Sub(t.Global, frame.pos), frame.size)
	var cell Vec2
	switch {
	case frac.X < low:
		cell.X = -1
	case frac.X < high:
		cell.X = 0
	case frac.X > high:
		cell.X = 1
	}
	switch {
	case frac.Y < low:
		cell.Y = -1
	case frac.Y < high:
		cell.Y = 0
	case frac.Y > high:
		cell.Y = 1
	}
	return &FrameDragging{frame, cell}
}

type FrameDragging struct {
	frame *Frame
	cell  Vec2
}

func (d *FrameDragging) Move(t *Touch) {
	delta := t.Delta()
	e, cell := d.frame, d.cell
	e.size.Sub(Mul(delta, cell))
	e.pos.Sub(Mul(Scale(delta, 0.5), Mul(cell, cell)))
	if cell.X == 0 && cell.Y == 0 {
		e.pos.Sub(delta)
	}
	if e.size.X < 30 {
		e.size.X = 30
	}
	if e.size.Y < 30 {
		e.size.Y = 30
	}
}

/* TODO: Convert to Drawable
func (d *FrameDragging) DragHighlight(widgets *Widgets) {
	e := d.frame
	s3 := Scale(e.size, 1/3.0)
	widgets.Rect(Add(e.pos, Mul(s3, d.cell)), s3, "rgba(0,0,0,0.2)")
}
*/

func (*FrameDragging) End(*Touch) {}
