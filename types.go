package mvm

import (
	"fmt"
	"math"

	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/ui"
	"github.com/mafik/mvm/vec2"
)

type VM struct {
	active *Object
}

type Blueprint struct {
	name      string
	frames    []*Frame
	instances map[*Object]bool
	transform matrix.Matrix
}

func (b *Blueprint) Name() string {
	return b.name
}

func (b *Blueprint) Frames() (out []*Frame) {
	for _, frame := range b.frames {
		if !frame.Hidden {
			out = append(out, frame)
		}
	}
	return out
}

func (b *Blueprint) Parameters() (params []Parameter) {
	for _, frame := range b.frames {
		if frame.param {
			params = append(params, frame)
		}
	}
	return
}

func (b *Blueprint) Instantiate(o *Object) {
	o.priv = &Machine{
		objects: make(map[*Frame]*Object),
	}
	b.instances[o] = true
}

func (b *Blueprint) Copy(from, to *Object) {
	b.Instantiate(to)
	for frame, proto := range from.priv.(*Machine).objects {
		new := MakeObject(proto.typ, frame, to)
		proto.typ.Copy(proto, new)
	}
}

func (b *Blueprint) Run(args Args) {
	self := args["self"]
	m := self.priv.(*Machine)
	for frame, object := range m.objects {
		if frame.name == "run" {
			object.MarkForExecution()
			return
		}
	}
	fmt.Println("Warning, couldn't find a \"run\" frame")
}

func (b *Blueprint) String(interface{}) string {
	return fmt.Sprintf("%d frames", len(b.Frames()))
}

func (b *Blueprint) MakeWidget(o *Object) ui.Widget { return BlueprintWidget{b, o} }

type BlueprintWidget struct {
	Blueprint *Blueprint
	Object    *Object
}

func (w BlueprintWidget) PostDraw(ctx *ui.Context2D) {
	b := w.Blueprint
	for _, frame := range b.frames {
		for i, _ := range frame.params {
			frame_parameter := &frame.params[i]
			if frame_parameter.Target == nil {
				continue
			}
			start := vec2.Add(frame.ParamCenter(i), frame.pos)
			end := frame_parameter.Target.pos
			delta := vec2.Sub(end, start)
			length := vec2.Len(delta)

			ctx.Save()
			ctx.Translate2(start)
			ctx.Rotate(math.Atan2(delta.Y, delta.X))

			if frame_parameter.Stiff && start != end {
				// white line outline
				ctx.StrokeStyle("#fff")
				ctx.BeginPath()
				ctx.MoveTo(0, 0)
				ctx.LineTo(length-4, 0)
				ctx.LineWidth(6.0)
				ctx.Stroke()

				// white arrow outline
				ctx.Save()
				ctx.FillStyle("#fff")
				ctx.Translate(length+4, 0)
				ctx.Arrow(13 + 6)
				ctx.Fill()
				ctx.Restore()
			}
			// line
			ctx.StrokeStyle("#000")
			ctx.LineWidth(2)
			ctx.BeginPath()
			ctx.MoveTo(0, 0)
			ctx.LineTo(length-5, 0)
			ctx.Stroke()

			// black circle
			ctx.FillStyle("#000")
			ctx.BeginPath()
			ctx.Circle(vec2.Vec2{0, 0}, param_r/4)
			ctx.Fill()

			// black arrow
			ctx.Translate(length, 0)
			ctx.Arrow(13)
			ctx.Fill()

			ctx.Restore()
		}
	}
}

func (w BlueprintWidget) Options(pos vec2.Vec2) []ui.Option {
	return []ui.Option{Navigate{w.Blueprint}, MakeFrame{w.Blueprint}, Zoom{w.Blueprint}}
}

func (w BlueprintWidget) Transform(ui.TextMeasurer) matrix.Matrix {
	return w.Blueprint.transform
}

func (w BlueprintWidget) Children() []interface{} {
	frames := w.Blueprint.frames
	children := make([]interface{}, 0, len(frames))
	for _, frame := range frames {
		if frame.Hidden {
			continue
		}
		children = append(children, FramedObject{frame, frame.Object(w.Object), w.Object})
	}
	return children
}

/*
func (b *Blueprint) Draw(o *Object, ctx *ui.Context2D) {
	for _, t := range Pointer.Touched {
		if fd, ok := t.(*FrameDragging); ok {
			f := fd.frame
			ctx.FillStyle("#ccc")
			ctx.BeginPath()
			f.PropagateStiff(func(f *Frame) {
				ctx.Rect2(Add(f.pos, shadowOffset), f.size)
				for i, _ := range f.Parameters(o) {
					pos := Add(f.ParamCenter(i), shadowOffset)
					ctx.Circle(pos, param_r)
					ctx.ClosePath()
				}
			})
			ctx.Fill()
		}
	}
}
*/

type Navigate struct{ blueprint *Blueprint }

func (d Navigate) Activate(ui.TouchContext) ui.Action { return d }
func (d Navigate) Keycode() string                    { return "KeyD" }
func (d Navigate) Name() string                       { return "Navigate" }
func (d Navigate) Move(ctx ui.TouchContext) ui.Action { return d }
func (d Navigate) End(ui.TouchContext)                {}
func (d Navigate) PreMove(ctx ui.TouchContext) {
	delta := ctx.Delta()
	ctx.Touch.Last = ctx.Touch.Curr
	d.blueprint.transform = matrix.Multiply(matrix.Translate(delta), d.blueprint.transform)
}

type Zoom struct{ blueprint *Blueprint }

func (z Zoom) Name() string    { return "Zoom" }
func (z Zoom) Keycode() string { return "Wheel" }
func (z Zoom) Activate(ctx ui.TouchContext) ui.Action {
	transform := &z.blueprint.transform
	a := ctx.Position()
	alpha := math.Exp(-ctx.Touch.Wheel / 200)
	transform.Scale(alpha)
	b := ctx.Position()
	fix := vec2.Sub(b, a)
	*transform = matrix.Multiply(matrix.Translate(fix), *transform) // apply translation before scaling
	return nil
}

type MakeFrame struct {
	Blueprint *Blueprint
}

func (mf MakeFrame) Name() string    { return "New frame" }
func (mf MakeFrame) Keycode() string { return "KeyX" }
func (mf MakeFrame) Activate(ctx ui.TouchContext) ui.Action {
	f := mf.Blueprint.AddFrame()
	f.pos = ctx.Position()
	return FrameDragging{f, vec2.Vec2{0, 0}}
}

/*
func (b *Blueprint) LinkInput(t *Touch, e Event) Touching {
	l := PointedLink(b, t.Global())
	if l == nil {
		return nil
	}
	switch e.Code {
	case "KeyF":
		target := b.MakeLinkTarget()
		target.pos = t.Global()
		l.SetTarget(target)
		return l
	case "KeyZ":
		l.param.Stiff = !l.param.Stiff
		if l.param.Stiff {
			return l.source.Drag(t)
		} else {
			return NoopTouching{}
		}
	case "KeyQ:":
		l.Delete()
		return NoopTouching{}
	}
	return nil
}

func (b *Blueprint) ParamInput(o *Object, t *Touch, e Event) Touching {
	frame, name := FindParamBelow(o, t.Global())
	if frame == nil {
		return nil
	}
	frameParam := frame.ForceGetLinkSet(name)
	result := HandleEdit(e, frameParam, &frameParam.Name, false)
	if result != nil {
		return result
	}
	switch e.Code {
	case "KeyF":
		target := frame.blueprint.MakeLinkTarget()
		target.pos = t.Global()
		return frame.AddLink(name, target)
	case "KeyQ":
		index, _ := GetParam(frame.LocalParameters(), name)
		if index != -1 {
			frame.params = append(frame.params[:index], frame.params[index+1:]...)
			return NoopTouching{}
		}
	}
	return nil
}

func (b *Blueprint) ObjectInput(parent *Object, t *Touch, e Event) Touching {
	o := FindObjectBelow(parent, t.Global())
	if o == nil {
		return nil
	}
	if _, ok := o.typ.(*Blueprint); e.Code == "KeyE" && ok {
		TheVM.active = o
		return NoopTouching{}
	} else if h := o.typ.Input(o, t, e); h != nil {
		return h
	} else if e.Code == "Space" {
		o.MarkForExecution()
		return NoopTouching{}
	}
	return nil
}

func (b *Blueprint) FrameInput(parent *Object, t *Touch, e Event) Touching {
	f := FindFrameTitleBelow(b, t.Global())
	if f == nil {
		f = FindFrameBelow(b, t.Global())
	}
	if f == nil {
		return nil
	}
	initial_name := f.name
	result := HandleEdit(e, f, &f.name, false)
	if result != nil {
		if f.param {
			b := f.blueprint
			for instance, _ := range b.instances {
				if instance.frame == nil {
					continue
				}
				for i, _ := range instance.frame.params {
					ls := &instance.frame.params[i]
					if ls.Name == initial_name {
						ls.Name = f.name
					}
				}
			}
		}
		return result
	}
	switch e.Code {
	case "KeyF":
		return f.StartDrag(t)
	case "KeyS":
		o := f.Object(parent)
		b := f.blueprint
		f2 := b.AddFrame()
		f2.pos = o.frame.pos
		f2.size = o.frame.size
		b.FillWithCopy(f2, o)
		return f2.Drag(t)
	case "KeyQ":
		f.Delete()
		return NoopTouching{}
	case "KeyR":
		f.params = append(f.params, FrameParameter{"", nil, false})
		return NoopTouching{}
	case "KeyE":
		f.param = !f.param
		return NoopTouching{}
	case "KeyT":
		new_bp := MakeBlueprint("new blueprint")
		b.FillWithNew(f, new_bp)
		return NoopTouching{}
	case "KeyX":
		o := f.Object(parent)
		clip := e.Client.Focus()
		b := clip.typ.(*Blueprint)
		f2 := b.AddFrame()
		f2.pos = f.pos
		f2.size = f.size
		b.FillWithCopy(f2, o)
		return f2.Drag(t)
	}
	return nil
}

func (b *Blueprint) BackgroundInput(o *Object, t *Touch, e Event) Touching {
	switch e.Code {
	case "KeyW":
		parent := TheVM.active.parent
		if parent != nil {
			TheVM.active = parent
		}
		return NoopTouching{}
	case "KeyD":
		return Navigate()
	case "KeyS":
		f := b.AddFrame()
		f.pos = Pointer.Vec2
		return &FrameDragging{f, Vec2{1, 1}}
	}
	return nil

}
*/

type BlueprintParameter struct {
}

func (p *BlueprintParameter) Instantiate(*Object) {}

func (p *BlueprintParameter) Name() string {
	return "param"
}

func (p *BlueprintParameter) Parameters() []Parameter {
	return nil
}

func (p *BlueprintParameter) Run(Args) {}

func (p *BlueprintParameter) String(interface{}) string { return "" }

type Args map[string]*Object

type Parameter interface {
	Name() string
	Typ() Type
}

type Type interface {
	Name() string
	Parameters() []Parameter
	Instantiate(*Object) // TODO: remove
	Copy(from, to *Object)
	Run(Args)
	String(interface{}) string
	MakeWidget(*Object) ui.Widget
}

type Machine struct {
	objects map[*Frame]*Object
}

func (vm VM) AvailableBlueprints() (bl []*Blueprint, active int) {
	bl = append(bl, vm.active.typ.(*Blueprint))
	active = 0
	return
}

func MakeParameter() *BlueprintParameter {
	return &BlueprintParameter{}
}

func MakeBlueprint(name string) *Blueprint {
	return &Blueprint{
		name:      name,
		frames:    nil,
		instances: make(map[*Object]bool),
		transform: matrix.Identity(),
	}
}

func (bp *Blueprint) AddFrame() *Frame {
	frame := &Frame{
		blueprint: bp,
		pos:       vec2.Vec2{0, 0},
		size:      vec2.Vec2{100, 100},
	}
	bp.frames = append(bp.frames, frame)
	return frame
}

func (bp *Blueprint) FillWithNew(frame *Frame, typ Type) {
	for instance, _ := range bp.instances {
		o := MakeObject(typ, frame, instance)
		typ.Instantiate(o)
	}
}

func (bp *Blueprint) FillWithCopy(frame *Frame, proto *Object) {
	for instance, _ := range bp.instances {
		o := MakeObject(proto.typ, frame, instance)
		proto.typ.Copy(proto, o)
	}
}

func (f *Frame) Object(blueprint_instance *Object) *Object {
	return blueprint_instance.priv.(*Machine).objects[f]
}

func (f *Frame) Type(blueprintInstance *Object) Type {
	o := f.Object(blueprintInstance)
	if o == nil {
		return nil
	}
	return o.typ
}

func (f *Frame) Name() string {
	return f.name
}

func (f *Frame) Typ() Type {
	return nil // TODO: implement
}

func (f *Frame) Title() string {
	if f.param {
		return "¶" + f.name
	}
	return f.name
}

func (f *Frame) Delete() {
	b := f.blueprint
	X := 0
	for x, frame := range b.frames {
		if frame == f {
			X = x
		}
		for s, _ := range frame.params {
			if f == frame.params[s].Target {
				frame.params[s].Target = nil
			}
		}
	}
	b.frames = append(b.frames[:X], b.frames[X+1:]...)
	for m, _ := range b.instances {
		delete(m.priv.(*Machine).objects, f)
	}
}

func (e *Frame) PropagateStiff(cb func(*Frame)) {
	visited := map[*Frame]bool{}
	for f, q := e, []*Frame{e}; len(q) > 0; f, q = q[len(q)-1], q[:len(q)-1] {
		if _, ok := visited[f]; ok {
			continue
		}
		visited[f] = true
		cb(f)
		for _, param := range f.params {
			if param.Stiff && param.Target != nil {
				q = append(q, param.Target)
			}
		}
	}
}

func GetParam(params []Parameter, name string) (int, Parameter) {
	for i, param := range params {
		if param.Name() == name {
			return i, param
		}
	}
	return -1, nil
}
