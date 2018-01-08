package mvm

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/ui"
	"github.com/mafik/mvm/vec2"
)

type Blueprint struct {
	name      string
	frames    []*Frame
	instances map[*Object]bool
	transform matrix.Matrix
	color     int
}

type Machine struct {
	objects map[*Frame]*Object
}

func MakeBlueprint(name string) *Blueprint {
	return &Blueprint{
		name:      name,
		frames:    nil,
		instances: make(map[*Object]bool),
		transform: matrix.Identity(),
		color:     rand.Intn(360),
	}
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

func (b *Blueprint) Members() (members []string) {
	for _, frame := range b.frames {
		if frame.public {
			members = append(members, frame.name)
		}
	}
	return
}

func (b *Blueprint) GetMember(self *Object, name string) *Object {
	for _, frame := range b.frames {
		if frame.name == name {
			return frame.Object(self)
		}
	}
	return nil
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
func (b *Blueprint) BrightColor() string            { return fmt.Sprintf("hsl(%d, 50%%, 85%%)", b.color) }
func (b *Blueprint) DarkColor() string              { return fmt.Sprintf("hsl(%d, 60%%, 20%%)", b.color) }

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

type BlueprintWidget struct {
	Blueprint *Blueprint
	Object    *Object
}

func (w BlueprintWidget) Draw(ctx *ui.Context2D) {
	ctx.BeginPath()
	ctx.FillStyle(w.Blueprint.BrightColor())
	ctx.Rect(-10000, -10000, 20000, 20000)
	ctx.Fill()
}

func (w BlueprintWidget) PostDraw(ctx *ui.Context2D) {
	b := w.Blueprint
	for _, frame := range b.frames {
		for i, _ := range frame.elems {
			frame_parameter := &frame.elems[i]
			if frame_parameter.Target == nil {
				continue
			}
			start := vec2.Add(frame.ParamCenter(i), frame.pos)
			end := frame_parameter.Target.pos
			if frame_parameter.TargetMember != "" {
				targetElement := frame_parameter.Target.GetElement(frame_parameter.TargetMember)
				targetI := 0
				for i, elem := range frame_parameter.Target.elems {
					if &elem == targetElement {
						targetI = i
						break
					}
				}
				end = vec2.Add(end, frame_parameter.Target.ParamCenter(targetI))
			}
			delta := vec2.Sub(end, start)
			length := vec2.Len(delta)

			if frame_parameter.TargetMember != "" {
				length -= param_r
			}

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
*/
