package mvm

type TouchSnapshot struct {
	Local  Vec2
	Global Vec2
}

type Touch struct {
	TouchSnapshot
	Last    TouchSnapshot
	Touched map[string]Touching
}

func (t *Touch) BeginTouching(source string, f func(*Touch) Touching) {
	if t.Touched[source] != nil {
		return
	}
	touching := f(t)
	if touching == nil {
		return
	}
	t.Touched[source] = touching
}

func (t *Touch) EndTouching(source string) {
	touching, ok := t.Touched[source]
	if ok {
		touching.End(t)
		delete(t.Touched, source)
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

var Pointer = Touch{Touched: map[string]Touching{}}

func (t TouchSnapshot) FindBlueprintBelow() *Blueprint {
	left := margin
	right := left + buttonWidth
	top := margin
	p := t.Local
	for it := TheVM.active; it != nil; it = it.parent {
		bottom := top + buttonHeight
		if p.X > left && p.X < right && p.Y > top && p.Y < bottom {
			return it.typ.(*Blueprint)
		}
		top += buttonHeight + margin
	}
	return nil
}

func (t TouchSnapshot) FindFrameBelow() *Frame {
	for _, frame := range TheVM.active.typ.(*Blueprint).Frames() {
		if frame.ContentHitTest(t.Global) {
			return frame
		}
	}
	return nil
}

func (t TouchSnapshot) FindFrameTitleBelow() *Frame {
	for _, frame := range TheVM.active.typ.(*Blueprint).Frames() {
		if frame.TitleHitTest(t.Global) {
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
	for _, frame := range TheVM.active.typ.(*Blueprint).frames {
		for i, _ := range frame.params {
			frame_parameter := &frame.params[i]
			if frame_parameter.Target == nil {
				continue
			}
			link := &Link{frame, frame_parameter}
			current_dist := LinkDist(t.Global, link)
			if current_dist < best_dist {
				best, best_dist = link, current_dist
			}
		}
	}
	return
}
