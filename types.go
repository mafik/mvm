package mvm

import (
	"fmt"
)

type VM struct {
	active *Object
}

type Blueprint struct {
	name      string
	frames    map[*Frame]bool
	instances map[*Object]bool
}

func (b *Blueprint) Name() string {
	return b.name
}

func (b *Blueprint) Frames() (out []*Frame) {
	for frame, _ := range b.frames {
		is_link := frame.Type() == LinkTargetType
		if !is_link {
			out = append(out, frame)
		}
	}
	return out
}

func (b *Blueprint) Parameters() (params []Parameter) {
	for frame, _ := range b.frames {
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

func (b *Blueprint) Run(args Args) {
	self := args["self"][0]
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
	index      int
}

type LinkSet struct {
	ParamName string
	Targets   []*Frame
}

type Frame struct {
	blueprint *Blueprint
	pos       Vec2
	size      Vec2
	name      string
	link_sets []LinkSet
	param     bool
}

type Args map[string][]*Object

type Parameter interface {
	Name() string
	Typ() Type
	Output() bool
}

type Type interface {
	Name() string
	Parameters() []Parameter
	Instantiate(*Object)
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
		frames:    make(map[*Frame]bool),
		instances: make(map[*Object]bool),
	}
}

func MakeObject(typ Type, frame *Frame) *Object {
	o := &Object{
		typ:   typ,
		frame: frame,
	}
	typ.Instantiate(o)
	return o
}

func (bp *Blueprint) Add(typ Type) *Frame {
	frame := &Frame{
		blueprint: bp,
		pos:       Vec2{0, 0},
		size:      Vec2{100, 100},
	}
	bp.frames[frame] = true
	for instance, _ := range bp.instances {
		object := MakeObject(typ, frame)
		object.parent = instance
		m := instance.priv.(*Machine)
		m.objects[frame] = object
	}
	return frame
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

func (f *Frame) Parameters() (params []Parameter) {
	if typ := f.Type(); typ != nil {
		params = typ.Parameters()
	}
	for _, link_set := range f.link_sets {
		_, existing := GetParam(params, link_set.ParamName)
		if existing != nil {
			continue
		}
		params = append(params, &FixedParameter{link_set.ParamName, nil, false})
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
	for frame, _ := range b.frames {
		for s, _ := range frame.link_sets {
			link_set := &frame.link_sets[s]
			for i := 0; i < len(link_set.Targets); i++ {
				if f == link_set.Targets[i] {
					link_set.Targets = append(link_set.Targets[:i], link_set.Targets[i+1:]...)
					i--
				}
			}
		}
	}
	delete(b.frames, f)
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
	link_set := link.source.GetLinkSet(link.param_name)
	link_set.Targets = append(link_set.Targets[:link.index], link_set.Targets[link.index+1:]...)
}

func (frame *Frame) DrawLinks(widgets *Widgets) {
	params := frame.Parameters()
	for _, link_set := range frame.link_sets {
		for i, _ := range link_set.Targets {
			link := Link{link_set.ParamName, frame, i}
			line := widgets.Line(link.StartPos(), link.EndPos())
			_, param := GetParam(params, link_set.ParamName)
			if param.Output() {
				line.Start = MakeCircle(param_r/4, "#000")
				line.End = MakeArrow()
			} else {
				line.Start = MakeArrow()
				line.End = MakeCircle(param_r/4, "#000")
			}
			line.Middle = MakeText(fmt.Sprint(i))
			line.Middle.Scale = .75
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

func (f *Frame) GetLinkSet(param_name string) *LinkSet {
	for i, link_set := range f.link_sets {
		if link_set.ParamName == param_name {
			return &f.link_sets[i]
		}
	}
	return nil
}

func (f *Frame) ForceGetLinkSet(param_name string) *LinkSet {
	links := f.GetLinkSet(param_name)
	if links == nil {
		f.link_sets = append(f.link_sets, LinkSet{param_name, nil})
		links = &f.link_sets[len(f.link_sets)-1]
	}
	return links
}

func (f *Frame) AddLink(param_name string, target *Frame) *Link {
	link_set := f.ForceGetLinkSet(param_name)
	link_set.Targets = append(link_set.Targets, target)
	return &Link{param_name, f, len(link_set.Targets) - 1}
}

func (l *Link) SetTarget(target *Frame) {
	l.source.ForceGetLinkSet(l.param_name).Targets[l.index] = target
}

func (l *Link) B() *Frame {
	return l.source.ForceGetLinkSet(l.param_name).Targets[l.index]
}
