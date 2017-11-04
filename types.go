package mvm

import (
	"fmt"
	"math"
	"strings"

	. "github.com/mafik/mvm/vec2"
)

type VM struct {
	active *Object
}

type Blueprint struct {
	name      string
	frames    []*Frame
	instances map[*Object]bool
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

func (b *Blueprint) Draw(o *Object, ctx *Context2D) {
	// Frames
	for _, frame := range b.Frames() {
		obj := frame.Object(o)
		left := frame.pos.X - frame.size.X/2
		top := frame.pos.Y - frame.size.Y/2

		title := frame.Title()
		titleWidth := ctx.MeasureText(title)
		frameTitleWidth := ButtonSize(titleWidth)

		if obj != nil {
			typeName := obj.typ.Name()

			// White background
			ctx.FillStyle("#fff")
			ctx.BeginPath()
			ctx.Rect2(frame.pos, Sub(frame.size, Vec2{2, 2}))
			ctx.Fill()
			ctx.FillRect(frame.ObjectLeft(ctx), frame.ObjectTop(), frame.ObjectWidth(obj, ctx), buttonHeight)
			typ := obj.typ
			text := typ.String(obj.priv)
			lines := strings.Split(text, "\n")
			ctx.FillStyle("#000")
			ctx.TextAlign("left")

			ctx.FillText(typeName, left+margin+frameTitleWidth, top-textMargin)

			for i, line := range lines {
				ctx.FillText(line, left+margin, top+margin+float64(i+1)*25)
			}

			if typ == TextType {
				width := ctx.MeasureText(lines[len(lines)-1])
				ctx.FillRect(left+margin+width, top+margin+float64(len(lines)-1)*25, 2, 30)
			}
			if obj.execute {
				ctx.FillStyle("#f00")
				ctx.BeginPath()
				ctx.Rect2(frame.pos, Add(frame.size, Vec2{10, 10}))
				ctx.Fill()
			}
			if obj.running {
				ctx.Save()
				ctx.Translate2(Add(frame.pos, Vec2{frame.size.X / 2, -frame.size.Y / 2}))
				ctx.Hourglass("#f00")
				ctx.Restore()
			}
		}
		// Black outline
		ctx.BeginPath()
		ctx.Rect2(frame.pos, Sub(frame.size, Vec2{2, 2}))
		if ctx.client.Editing(obj) {
			DrawEditingOverlay(ctx)
		} else {
			ctx.LineWidth(2)
			ctx.StrokeStyle("#000")
			ctx.Stroke()
		}

		ctx.FillStyle("#000")
		ctx.BeginPath()
		ctx.Rect(left, top, frameTitleWidth, -ButtonSize(0))
		ctx.Fill()
		if ctx.client.Editing(frame) {
			DrawEditingOverlay(ctx)
		}
		ctx.FillStyle("#fff")
		ctx.TextAlign("left")
		ctx.FillText(title, left+margin, top-textMargin)
	}

	// Parameters
	ctx.LineWidth(2)
	for _, frame := range b.Frames() {
		local_params := frame.LocalParameters()
		type_params := frame.TypeParameters(o)
		params := frame.Parameters(o)
		n := len(params)
		if n > 0 {
			ctx.BeginPath()
			ctx.MoveTo2(Sub(frame.ParamCenter(0), Vec2{0, param_r + margin}))
			ctx.LineTo2(frame.ParamCenter(n - 1))
			ctx.Stroke()
		}

		for i, param := range params {
			pos := frame.ParamCenter(i)
			ctx.BeginPath()
			ctx.Circle(pos, param_r)
			if idx, _ := GetParam(type_params, param.Name()); idx >= 0 {
				ctx.FillStyle("#fff")
				ctx.Fill()
			}
			if idx, _ := GetParam(local_params, param.Name()); idx >= 0 {
				frameParameter := frame.ForceGetLinkSet(param.Name())
				if ctx.client.Editing(frameParameter) {
					DrawEditingOverlay(ctx)
				} else {
					ctx.StrokeStyle("#000")
					ctx.Stroke()
				}
			}
		}
	}

	// Links
	for _, frame := range b.frames {
		for i, _ := range frame.params {
			frame_parameter := &frame.params[i]
			if frame_parameter.Target == nil {
				continue
			}
			link := Link{frame, frame_parameter}
			start := link.StartPos()
			end := link.EndPos()
			delta := Sub(end, start)
			length := Len(delta)

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
			ctx.Circle(Vec2{0, 0}, param_r/4)
			ctx.Fill()

			// black arrow
			ctx.Translate(length, 0)
			ctx.Arrow(13)
			ctx.Fill()

			ctx.Restore()
		}
	}

	// Parameter names
	ctx.FillStyle("#000")
	ctx.TextAlign("left")
	for _, frame := range b.Frames() {
		params := frame.Parameters(o)
		for i, param := range params {
			pos := frame.ParamCenter(i)
			pos.Y -= 3
			pos.X += param_r + margin
			ctx.FillText(param.Name(), pos.X, pos.Y)
		}
	}
}

func (b *Blueprint) LinkInput(t *Touch, e Event) Touching {
	l := PointedLink(b, t.TouchSnapshot)
	if l == nil {
		return nil
	}
	switch e.Code {
	case "KeyF":
		target := b.MakeLinkTarget()
		target.pos = t.Global
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
	frame, name := FindParamBelow(o, t.TouchSnapshot)
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
		target.pos = t.Global
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
	o := FindObjectBelow(parent, t.TouchSnapshot)
	if o == nil {
		return nil
	}
	// TODO: clean this up
	if _, ok := o.typ.(*Blueprint); e.Code == "KeyE" && ok {
		TheVM.active = o
	} else if o.typ == TextType {
		s := string(o.priv.([]byte))
		h := HandleEdit(e, o, &s, true)
		if h == nil {
			return nil
		}
		o.priv = []byte(s)
		return h
	} else if e.Code == "Space" {
		o.MarkForExecution()
	} else {
		return nil
	}
	return NoopTouching{}
}

func (b *Blueprint) FrameInput(parent *Object, t *Touch, e Event) Touching {
	f := FindFrameTitleBelow(b, Pointer.TouchSnapshot)
	if f == nil {
		f = FindFrameBelow(b, Pointer.TouchSnapshot)
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

func (b *Blueprint) Input(o *Object, t *Touch, e Event) Touching {
	// Links
	if result := b.LinkInput(t, e); result != nil {
		return result
	}
	// Params
	if result := b.ParamInput(o, t, e); result != nil {
		return result
	}
	// Objects
	if result := b.ObjectInput(o, t, e); result != nil {
		return result
	}
	// Frames
	if result := b.FrameInput(o, t, e); result != nil {
		return result
	}
	return nil
}

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

type Link struct {
	source *Frame
	param  *FrameParameter
}

type FrameParameter struct {
	Name   string
	Target *Frame
	Stiff  bool
}

type Frame struct {
	blueprint *Blueprint
	pos       Vec2
	size      Vec2
	name      string
	params    []FrameParameter
	param     bool
	Hidden    bool
}

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

	Draw(*Object, *Context2D)
	Input(*Object, *Touch, Event) Touching
}

type Machine struct {
	objects map[*Frame]*Object
}

type Object struct {
	parent  *Object
	frame   *Frame
	execute bool
	running bool
	typ     Type
	priv    interface{}
}

func (vm VM) AvailableBlueprints() (bl []*Blueprint, active int) {
	bl = append(bl, vm.active.typ.(*Blueprint))
	active = 0
	return
}

func (o *Object) MarkForExecution() {
	o.execute = true
	tasks <- o
}

func MakeParameter() *BlueprintParameter {
	return &BlueprintParameter{}
}

func MakeBlueprint(name string) *Blueprint {
	return &Blueprint{
		name:      name,
		frames:    nil,
		instances: make(map[*Object]bool),
	}
}

func MakeObject(typ Type, frame *Frame, parent *Object) *Object {
	o := &Object{
		typ:    typ,
		frame:  frame,
		parent: parent,
	}
	if parent != nil {
		m := parent.priv.(*Machine)
		m.objects[frame] = o
	}
	return o
}

func (bp *Blueprint) AddFrame() *Frame {
	frame := &Frame{
		blueprint: bp,
		pos:       Vec2{0, 0},
		size:      Vec2{100, 100},
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

func (f *Frame) LocalParameters() (params []Parameter) {
	for _, frame_parameter := range f.params {
		params = append(params, &FixedParameter{frame_parameter.Name, nil})
	}
	return params
}

func (f *Frame) TypeParameters(blueprintInstance *Object) []Parameter {
	t := f.Type(blueprintInstance)
	if t == nil {
		return nil
	}
	return t.Parameters()
}

func (f *Frame) Parameters(blueprintInstance *Object) (params []Parameter) {
	params = f.LocalParameters()
	if typ := f.Type(blueprintInstance); typ != nil {
		for _, tp := range typ.Parameters() {
			_, existing := GetParam(params, tp.Name())
			if existing != nil {
				continue
			}
			params = append(params, tp)
		}
	}
	return params
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

func (f *Frame) TitleHitTest(p Vec2) bool {
	l := f.pos
	s := f.size
	switch {
	case p.X < f.TitleLeft():
		return false
	case p.X > l.X+s.X/2:
		return false
	case p.Y < f.TitleTop():
		return false
	case p.Y > f.TitleBottom():
		return false
	}
	return true
}

func (f *Frame) ContentHitTest(p Vec2) bool {
	l := f.pos
	s := f.size
	switch {
	case p.X < l.X-s.X/2:
		return false
	case p.X > l.X+s.X/2:
		return false
	case p.Y < l.Y-s.Y/2:
		return false
	case p.Y > l.Y+s.Y/2:
		return false
	}
	return true
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
func (link Link) StartPos() Vec2 {
	i, _ := GetParam(link.source.LocalParameters(), link.param.Name)
	return link.source.ParamCenter(i)
}

func (link Link) EndPos() Vec2 {
	a := link.StartPos()
	target := link.B()
	return Ray(a, target.pos, target.size)
}

func (link *Link) Delete() {
	link.param.Target = nil
}

func GetParam(params []Parameter, name string) (int, Parameter) {
	for i, param := range params {
		if param.Name() == name {
			return i, param
		}
	}
	return -1, nil
}

func (f *Frame) GetLinkSet(param_name string) *FrameParameter {
	for i, frame_parameter := range f.params {
		if frame_parameter.Name == param_name {
			return &f.params[i]
		}
	}
	return nil
}

func (f *Frame) ForceGetLinkSet(param_name string) *FrameParameter {
	links := f.GetLinkSet(param_name)
	if links == nil {
		f.params = append(f.params, FrameParameter{param_name, nil, false})
		links = &f.params[len(f.params)-1]
	}
	return links
}

func (f *Frame) AddLink(param_name string, target *Frame) *Link {
	frame_parameter := f.ForceGetLinkSet(param_name)
	frame_parameter.Target = target
	return &Link{f, frame_parameter}
}

func (l *Link) SetTarget(target *Frame) {
	l.param.Target = target
}

func (l *Link) B() *Frame {
	return l.param.Target
}
