package mvm

// On-screen element that can be dragged around.
type Draggable interface {
	Drag(*Touch) Touching
}

func (c *LayerList) Drag(touch *Touch) Touching {
	for _, elem := range *c {
		if t := elem.Drag(touch); t != nil {
			return t
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
	f := b.AddFrame()
	b.FillWithNew(f, LinkTargetType)
	return f
}

func (ParamLayer) Drag(t *Touch) Touching {
	frame, param := t.FindParamBelow()
	if frame == nil {
		return nil
	}
	target := frame.blueprint.MakeLinkTarget()
	target.pos = t.Global
	return frame.AddLink(param, target)
}

func (LinkLayer) Drag(t *Touch) Touching {
	link := t.PointedLink()
	if link == nil {
		return nil
	}
	target := TheVM.active.typ.(*Blueprint).MakeLinkTarget()
	target.pos = t.Global
	link.SetTarget(target)
	return link
}

func (l *Link) Move(touch *Touch) {
	l.B().pos = touch.Global
}

func (l *Link) End(touch *Touch) {
	frame := touch.FindFrameBelow()
	if frame == nil {
		l.B().Delete()
	} else {
		fake_target := l.B()
		l.SetTarget(frame)
		fake_target.Delete()
	}
}

// Frames

func (FrameLayer) Drag(t *Touch) Touching {
	frame := t.FindFrameBelow()
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

func (f *Frame) Drag(t *Touch) Touching {
	return &FrameDragging{f, Vec2{0, 0}}
}

type FrameDragging struct {
	frame *Frame
	cell  Vec2
}

func (d *FrameDragging) Move(t *Touch) {
	delta := t.Delta()
	e, cell := d.frame, d.cell
	start := e.pos
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
	end := e.pos
	movement := Sub(end, start)
	e.PropagateStiff(func(f *Frame) {
		if f != e {
			f.pos.Add(movement)
		}
	})
}

func (*FrameDragging) End(*Touch) {}

func (BackgroundLayer) Drag(*Touch) Touching { return nil }
func (ParamNameLayer) Drag(*Touch) Touching  { return nil }
func (OverlayLayer) Drag(*Touch) Touching    { return nil }
func (ObjectLayer) Drag(*Touch) Touching     { return nil }
