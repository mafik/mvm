package mvm

type TouchSnapshot struct {
	Local  Vec2
	Global Vec2
}

type Touch struct {
	TouchSnapshot
	Last    TouchSnapshot
	Touched Touching
	Source  string
}

func (t *Touch) BeginTouching(source string, f func(*Touch) Touching) {
	if t.Touched != nil {
		return
	}
	touched := f(t)
	if touched == nil {
		return
	}
	t.Touched = touched
	t.Source = source
}

func (t *Touch) EndTouching(source string) {
	if t.Source == source {
		t.Touched.End(t)
		t.Touched = nil
		t.Source = ""
	}
}

func (t *TouchSnapshot) UpdateGlobal() {
	t.Global = window.ToGlobal(t.Local)
}

func (t Touch) Delta() Vec2 {
	return Sub(t.Last.Global, t.TouchSnapshot.Global)
}

type Touching interface {
	Move(*Touch)
	End(*Touch)
}

type NoopTouching struct{}

func (noop NoopTouching) Move(*Touch) {}
func (noop NoopTouching) End(*Touch)  {}

// Container of touchable elements
type Container struct {
	elements []interface{}
}

var Pointer Touch

func (t TouchSnapshot) FindFrameBelow() *Frame {
	for _, frame := range TheVM.active.typ.(*Blueprint).Frames() {
		if frame.HitTest(t.Global) {
			return frame
		}
	}
	return nil
}

func (t TouchSnapshot) FindObjectBelow() *Object {
	machine := TheVM.active.priv.(*Machine)
	frame := t.FindFrameBelow()
	return machine.objects[frame]
}

func (t TouchSnapshot) FindParamBelow() (*Frame, string) {
	blueprint := TheVM.active.typ.(*Blueprint)
	for _, f := range blueprint.Frames() {
		for i, param := range f.Parameters() {
			if CircleClicked(f.ParamCenter(i), t.Global) {
				return f, param.Name()
			}
		}
	}
	return nil, ""
}

func CircleClicked(pos Vec2, touch Vec2) bool {
	return Dist(pos, touch) < param_r
}

func LinkDist(p Vec2, link *Link) float64 {
	start, end := link.StartPos(), link.EndPos()
	l := Dist(start, end)
	if l == 0 {
		return Dist(p, start)
	}
	l = l * l
	alpha := Clamp(0, 1, Dot(Sub(p, start), Sub(end, start))/l)
	proj := Add(start, Scale(Sub(end, start), alpha))
	return Dist(p, proj)
}

func (t TouchSnapshot) PointedLink() (best *Link) {
	best_dist := 8.0
	for frame, _ := range TheVM.active.typ.(*Blueprint).frames {
		for _, link_set := range frame.link_sets {
			for i, _ := range link_set.Targets {
				link := &Link{link_set.ParamName, frame, i}
				current_dist := LinkDist(t.Global, link)
				if current_dist < best_dist {
					best, best_dist = link, current_dist
				}
			}
		}
	}
	return
}
