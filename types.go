package mvm

import (
	"fmt"
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
		is_link := frame.Type() == LinkTargetType
		if !is_link {
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
	param_name string
	source     *Frame
}

type FrameParameter struct {
	Name   string
	Target *Frame
}

type Frame struct {
	blueprint *Blueprint
	pos       Vec2
	size      Vec2
	name      string
	params    []FrameParameter
	param     bool
}

type Args map[string]*Object

type Parameter interface {
	Name() string
	Typ() Type
	Output() bool
}

type Type interface {
	Name() string
	Parameters() []Parameter
	Instantiate(*Object)
	Copy(from, to *Object)
	Run(Args)
	String(interface{}) string
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

func (f *Frame) Type() Type {
	o := f.Object(TheVM.active)
	if o == nil {
		return nil
	}
	return o.typ
}

func (f *Frame) LocalParameters() (params []Parameter) {
	for _, frame_parameter := range f.params {
		params = append(params, &FixedParameter{frame_parameter.Name, nil, false})
	}
	return params
}

func (f *Frame) Parameters() (params []Parameter) {
	params = f.LocalParameters()
	if typ := f.Type(); typ != nil {
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

func (f *Frame) Output() bool {
	return false // TODO: implement
}

func (f *Frame) Typ() Type {
	return nil // TODO: implement
}

func (f *Frame) Title() string {
	tname := "nil"
	if t := f.Type(); t != nil {
		tname = t.Name()
	}
	name := fmt.Sprintf("%s:%s", f.name, tname)
	if f.param {
		name = "¶" + name
	}
	return name
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

func (f *Frame) HitTestTitle(p Vec2) bool {
	l := f.pos
	s := f.size
	switch {
	case p.X < l.X-s.X/2:
		return false
	case p.X > l.X+s.X/2:
		return false
	case p.Y < l.Y-s.Y/2-20:
		return false
	case p.Y > l.Y-s.Y/2:
		return false
	}
	return true
}

func (f *Frame) HitTest(p Vec2) bool {
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

func (link Link) StartPos() Vec2 {
	i, _ := GetParam(link.source.Parameters(), link.param_name)
	return link.source.ParamCenter(i)
}

func (link Link) EndPos() Vec2 {
	a := link.StartPos()
	target := link.B()
	return Ray(a, target.pos, target.size)
}

func (link *Link) Delete() {
	frame_parameter := link.source.GetLinkSet(link.param_name)
	frame_parameter.Target = nil
}

func (frame *Frame) DrawLinks(widgets *Widgets) {
	params := frame.Parameters()
	for _, frame_parameter := range frame.params {
		if frame_parameter.Target == nil {
			continue
		}
		link := Link{frame_parameter.Name, frame}
		line := widgets.Line(link.StartPos(), link.EndPos())
		_, param := GetParam(params, frame_parameter.Name)
		if param.Output() {
			line.Start = MakeCircle(param_r/4, "#000", "")
			line.End = MakeArrow()
		} else {
			line.Start = MakeArrow()
			line.End = MakeCircle(param_r/4, "#000", "")
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
		f.params = append(f.params, FrameParameter{param_name, nil})
		links = &f.params[len(f.params)-1]
	}
	return links
}

func (f *Frame) AddLink(param_name string, target *Frame) *Link {
	frame_parameter := f.ForceGetLinkSet(param_name)
	frame_parameter.Target = target
	return &Link{param_name, f}
}

func (l *Link) SetTarget(target *Frame) {
	l.source.ForceGetLinkSet(l.param_name).Target = target
}

func (l *Link) B() *Frame {
	return l.source.ForceGetLinkSet(l.param_name).Target
}
