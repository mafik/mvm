package mvm

import (
	"github.com/mafik/mvm/ui"
	"github.com/mafik/mvm/vec2"
)

// Raise

type Raise struct {
	Frame *Frame
}

func (Raise) Name() string    { return "Raise" }
func (Raise) Keycode() string { return "KeyG" }
func (r Raise) Activate(ctx ui.TouchContext) ui.Action {
	var newB *Blueprint
	var newPos vec2.Vec2
	f := r.Frame
	oldB := f.blueprint
	ctx.AtTopBlueprint().Query(func(path ui.TreePath, pos vec2.Vec2) ui.WalkAction {
		last := path[len(path)-1]
		bw, ok := last.(BlueprintWidget)
		if !ok {
			return ui.Explore
		}
		if bw.Blueprint != oldB {
			newB = bw.Blueprint
			newPos = pos
			return ui.Explore
		}
		if newB == nil {
			return ui.Return
		}

		oldF := bw.Object.frame
		newB.frames = append(newB.frames, f)
		f.blueprint = newB
		f.pos = newPos

		X := 0
		for x, frame := range oldB.frames {
			if frame == f {
				X = x
			}
			// TODO: create parameter object and point it to the new frame
			for s, _ := range frame.elems {
				if f == frame.elems[s].Target {
					frame.elems[s].Target = nil
				}
			}
		}
		oldB.frames = append(oldB.frames[:X], oldB.frames[X+1:]...)

		for newO, _ := range newB.instances {
			newM := newO.priv.(*Machine)
			oldO, ok := newM.objects[oldF]
			if !ok {
				continue
			}
			if oldO.typ != oldB {
				continue
			}
			oldM := oldO.priv.(*Machine)
			o, ok := oldM.objects[f]
			if !ok {
				continue
			}
			delete(oldM.objects, f)
			newM.objects[f] = o
			o.parent = newO
		}

		return ui.Return
	})
	return nil
}

// Lower

type Lower struct {
	Frame           *Frame
	BlueprintObject *Object
}

func (Lower) Name() string    { return "Lower" }
func (Lower) Keycode() string { return "KeyB" }
func (l Lower) Activate(ctx ui.TouchContext) ui.Action {
	var top *BlueprintWidget
	mine := false
	ctx.AtTopBlueprint().Query(func(path ui.TreePath, pos vec2.Vec2) ui.WalkAction {
		last := path[len(path)-1]
		if bw, ok := last.(BlueprintWidget); ok {
			//fmt.Println("Detected blueprint widget!")
			if top == nil {
				top = &bw
			}
			if mine {
				f := l.Frame
				oldB := f.blueprint
				newB := bw.Blueprint
				newF := bw.Object.frame
				newB.frames = append(newB.frames, f)
				f.blueprint = newB
				f.pos = pos

				X := 0
				for x, frame := range oldB.frames {
					if frame == f {
						X = x
					}
					// TODO: make the frame public and create parameters to maintain connections
					for s, _ := range frame.elems {
						if f == frame.elems[s].Target {
							frame.elems[s].Target = nil
						}
					}
				}
				oldB.frames = append(oldB.frames[:X], oldB.frames[X+1:]...)

				for oldO, _ := range oldB.instances {
					oldM := oldO.priv.(*Machine)
					newO, ok := oldM.objects[newF]
					if !ok {
						continue
					}
					if newO.typ != newB {
						continue
					}
					o, ok := oldM.objects[f]
					if !ok {
						continue
					}
					delete(oldM.objects, f)
					newM := newO.priv.(*Machine)
					newM.objects[f] = o
					o.parent = newO

				}
				return ui.Return
			}
			if bw.Object == l.BlueprintObject {
				mine = true
				//fmt.Println("That's my blueprint!")
			}
		}
		return ui.Explore
	})
	return nil
}

// Schedule

type Schedule struct {
	Frame  *Frame
	Object *Object
}

func (s Schedule) Name() string    { return "Schedule" }
func (s Schedule) Keycode() string { return "Space" }
func (s Schedule) Activate(ui.TouchContext) ui.Action {
	s.Object.MarkForExecution()
	return nil
}

// Enter

type Enter struct {
	Object *Object
}

func (e Enter) Name() string    { return "Enter" }
func (e Enter) Keycode() string { return "KeyE" }
func (e Enter) Activate(ctx ui.TouchContext) ui.Action {
	clientUI := ctx.Path[0].(*ClientUI)
	clientUI.focus = e.Object
	return nil
}

// New Blueprint

type NewBlueprint struct {
	Frame           *Frame
	BlueprintObject *Object
}

func (nb NewBlueprint) Name() string    { return "New blueprint" }
func (nb NewBlueprint) Keycode() string { return "KeyZ" }
func (nb NewBlueprint) Activate(ctx ui.TouchContext) ui.Action {
	b := MakeBlueprint("New blueprint")
	nb.Frame.blueprint.FillWithNew(nb.Frame, b)
	return nil
}

// Clear Frame

type ClearFrame struct {
	Frame  *Frame
	Object *Object
}

func (cf ClearFrame) Name() string    { return "Clear frame" }
func (cf ClearFrame) Keycode() string { return "KeyZ" }
func (cf ClearFrame) Activate(ctx ui.TouchContext) ui.Action {
	m := cf.Object.parent.priv.(*Machine)
	delete(m.objects, cf.Frame)
	return nil
}

// Copy Frame

type CopyFrame struct {
	Frame  *Frame
	Object *Object
}

func (cf CopyFrame) Name() string    { return "Copy frame" }
func (cf CopyFrame) Keycode() string { return "KeyC" }
func (cf CopyFrame) Activate(ctx ui.TouchContext) ui.Action {
	f := cf.Frame.blueprint.AddFrame()
	f.pos = ctx.AtTopBlueprint().Position()
	f.size = cf.Frame.size
	if cf.Object != nil {
		cf.Frame.blueprint.FillWithNew(f, cf.Object.typ)
	}
	return FrameDragging{f, vec2.Vec2{0, 0}}
}

// Frame clone

type CloneFrame struct {
	Frame  *Frame
	Object *Object
}

func (cf CloneFrame) Name() string    { return "Clone frame" }
func (cf CloneFrame) Keycode() string { return "KeyV" }
func (cf CloneFrame) Activate(ctx ui.TouchContext) ui.Action {
	f := cf.Frame.blueprint.AddFrame()
	f.pos = ctx.AtTopBlueprint().Position()
	f.size = cf.Frame.size
	if cf.Object != nil {
		cf.Frame.blueprint.FillWithCopy(f, cf.Object)
	}
	return FrameDragging{f, vec2.Vec2{0, 0}}
}

// Frame deletion

type DeleteFrame struct {
	Frame *Frame
}

func (df DeleteFrame) Name() string    { return "Delete frame" }
func (df DeleteFrame) Keycode() string { return "KeyQ" }
func (df DeleteFrame) Activate(ui.TouchContext) ui.Action {
	df.Frame.Delete()
	return nil
}

// Toggle parameter (frame)

type ToggleParameter struct {
	*Frame
}

func (ToggleParameter) Name() string    { return "Toggle parameter" }
func (ToggleParameter) Keycode() string { return "KeyT" }
func (tp ToggleParameter) Activate(ui.TouchContext) ui.Action {
	tp.param = !tp.param
	return nil
}

// Toggle public

type TogglePublic struct {
	*Frame
}

func (TogglePublic) Name() string    { return "Toggle public" }
func (TogglePublic) Keycode() string { return "KeyY" }
func (tp TogglePublic) Activate(ui.TouchContext) ui.Action {
	tp.public = !tp.public
	return nil
}

// Toggle show window

type ToggleShowWindow struct {
	*Frame
}

func (ToggleShowWindow) Name() string    { return "Toggle window" }
func (ToggleShowWindow) Keycode() string { return "KeyH" }
func (tsw ToggleShowWindow) Activate(ui.TouchContext) ui.Action {
	tsw.ShowWindow = !tsw.ShowWindow
	return nil
}

// Add Parameter (frame / object parameter)

type AddParameter struct {
	*Frame
}

func (AddParameter) Name() string    { return "Add parameter" }
func (AddParameter) Keycode() string { return "KeyR" }
func (ap AddParameter) Activate(ui.TouchContext) ui.Action {
	ap.GetElement("")
	return nil
}

// Delete Parameter

type DeleteParameter struct {
	*Frame
	ParameterIndex int
}

func (DeleteParameter) Name() string    { return "Delete parameter" }
func (DeleteParameter) Keycode() string { return "KeyQ" }
func (dp DeleteParameter) Activate(ui.TouchContext) ui.Action {
	dp.elems = append(dp.elems[:dp.ParameterIndex], dp.elems[dp.ParameterIndex+1:]...)
	return nil
}

// Frame Dragging

type FrameDragging struct {
	Frame *Frame
	Cell  vec2.Vec2
}

func StartFrameDragging(p vec2.Vec2, f *Frame) FrameDragging {
	threshold := 1. / 3.
	frac := vec2.Div(p, f.size)
	var cell vec2.Vec2
	switch {
	case frac.X < threshold:
		cell.X = -1
	case frac.X < 2*threshold:
		cell.X = 0
	default:
		cell.X = 1
	}
	switch {
	case frac.Y < threshold:
		cell.Y = -1
	case frac.Y < 2*threshold:
		cell.Y = 0
	default:
		cell.Y = 1
	}
	return FrameDragging{f, cell}
}
func (d FrameDragging) Name() string    { return "Move" }
func (d FrameDragging) Keycode() string { return "KeyF" }
func (d FrameDragging) Activate(ui.TouchContext) ui.Action {
	b := d.Frame.blueprint
	X := 0
	for i, frame := range b.frames {
		if frame == d.Frame {
			X = i
			break
		}
	}
	l := len(b.frames)
	if X < l {
		b.frames[X], b.frames[l-1] = b.frames[l-1], b.frames[X]
	}
	return d
}
func (d FrameDragging) End(ui.TouchContext) {}
func (d FrameDragging) Move(ctx ui.TouchContext) ui.Action {
	delta := ctx.Delta()
	e, cell := d.Frame, d.Cell
	start := e.pos
	e.size.Add(vec2.Mul(delta, cell))
	if cell.X == -1 {
		e.pos.X += delta.X
	}
	if cell.Y == -1 {
		e.pos.Y += delta.Y
	}
	if cell.X == 0 && cell.Y == 0 {
		e.pos.Add(delta)
	}
	if e.size.X < 30 {
		e.size.X = 30
	}
	if e.size.Y < 30 {
		e.size.Y = 30
	}
	end := e.pos
	movement := vec2.Sub(end, start)
	e.PropagateStiff(func(f *Frame) {
		if f != e {
			f.pos.Add(movement)
		}
	})
	return d
}

// Parameter dragging

func IsBlueprintWidget(i interface{}) bool {
	_, ok := i.(BlueprintWidget)
	return ok
}

type ParameterDragging struct {
	Frame     *Frame
	ParamName string
}

func (d ParameterDragging) Param() *FrameElement {
	return d.Frame.GetElement(d.ParamName)
}
func (d ParameterDragging) Name() string    { return "Connect" }
func (d ParameterDragging) Keycode() string { return "KeyF" }
func (d ParameterDragging) Activate(t ui.TouchContext) ui.Action {
	dummyTarget := d.Frame.blueprint.MakeLinkTarget()
	dummyTarget.pos = t.At(IsBlueprintWidget).Position()
	d.Param().Target = dummyTarget
	return d
}
func (d ParameterDragging) Move(t ui.TouchContext) ui.Action {
	d.Param().Target.Frame().pos.Add(t.Delta())
	return d
}
func (d ParameterDragging) End(ctx ui.TouchContext) {
	dummy := d.Param().Target.Frame()
	ctx.At(IsBlueprintWidget).Query(func(path ui.TreePath, p vec2.Vec2) ui.WalkAction {
		elem := path[len(path)-1]
		elementWidget, ok := elem.(FrameElementCircle)
		if ok && elementWidget.IsMember {
			d.Param().Target = elementWidget.Frame.GetElement(elementWidget.Name)
		}
		frameElement, ok := elem.(FramePart)
		if ok {
			d.Param().Target = frameElement.MyFrame()
			return ui.Return
		}
		return ui.Explore
	})
	dummy.Delete()
}
