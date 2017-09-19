package mvm

import (
	"fmt"
	"sort"
)

type VM struct {
	Blueprints      map[*Blueprint]bool
	ActiveBlueprint *Blueprint
}

type Blueprint struct {
	name           string
	frames         map[*Frame]bool
	machines       map[*Machine]bool
	active_machine *Machine
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

func (b *Blueprint) Parameters() []Parameter {
	return nil
}

func (b *Blueprint) Instantiate(o *Object) {
	o.priv = &Machine{
		objects: make(map[*Frame]*Object),
	}
}

func (b *Blueprint) Run(args Args) {
	panic("unimplemented")
}

func (b *Blueprint) String(interface{}) string {
	return fmt.Sprintf("%d frames", len(b.Frames()))
}

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
	machine *Machine
	frame   *Frame
	execute bool
	Running bool
	typ     Type
	priv    interface{}
}

func (vm VM) OrderedBlueprints() ([]*Blueprint, int) {
	var bl ByName
	for b, _ := range vm.Blueprints {
		bl = append(bl, b)
	}
	sort.Sort(bl)
	var active int
	for i, b := range bl {
		if b == vm.ActiveBlueprint {
			active = i
		}
	}
	return bl, active
}

type ByName []*Blueprint

func (b ByName) Len() int           { return len(b) }
func (b ByName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByName) Less(i, j int) bool { return b[i].name < b[j].name }

func (o *Object) MarkForExecution() {
	o.execute = true
	tasks <- o
}

func MakeBlueprint(name string) *Blueprint {
	bp := &Blueprint{
		name:     name,
		frames:   make(map[*Frame]bool),
		machines: make(map[*Machine]bool),
	}
	bp.active_machine = &Machine{
		objects: make(map[*Frame]*Object),
	}
	bp.machines = map[*Machine]bool{bp.active_machine: true}
	return bp
}

func (bp *Blueprint) Add(typ Type) *Frame {
	frame := &Frame{
		blueprint: bp,
		pos:       Vec2{0, 0},
		size:      Vec2{100, 100},
	}
	bp.frames[frame] = true
	m := bp.active_machine
	object := &Object{
		machine: m,
		frame:   frame,
		typ:     typ,
	}
	typ.Instantiate(object)
	m.objects[frame] = object
	return frame
}

func (f *Frame) Object() *Object {
	return f.blueprint.active_machine.objects[f]
}

func (f *Frame) Type() Type {
	o := f.Object()
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

func (f *Frame) Title() string {
	tname := "nil"
	if t := f.Type(); t != nil {
		tname = t.Name()
	}
	return fmt.Sprintf("%s:%s", f.name, tname)
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
	for m, _ := range b.machines {
		delete(m.objects, f)
	}
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
