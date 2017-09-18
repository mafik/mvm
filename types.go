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
	frames_        map[*Frame]bool
	links          map[*Link]bool
	machines       map[*Machine]bool
	active_machine *Machine
}

func (b *Blueprint) Name() string {
	return b.name
}

func (b *Blueprint) Frames() (out []*Frame) {
	for frame, _ := range b.frames_ {
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
	o.priv = &Machine{}
}

func (b *Blueprint) Run(args Args) {
	panic("unimplemented")
}

func (b *Blueprint) String(interface{}) string {
	return fmt.Sprintf("%d frames", len(b.Frames()))
}

type Link struct {
	Blueprint *Blueprint
	A, B      *Frame
	Param     string
	Order     int
}

type Frame struct {
	blueprint *Blueprint
	pos       Vec2
	size      Vec2
	Name      string
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
		name:    name,
		frames_: make(map[*Frame]bool),
		links:   make(map[*Link]bool),
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
	bp.frames_[frame] = true
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
		fmt.Println("Object is nil!")
		return nil
	}
	return o.typ
}

func (f *Frame) Parameters() (params []Parameter) {
	if typ := f.Type(); typ != nil {
		params = typ.Parameters()
	}
	for param_link, _ := range f.blueprint.links {
		if f != param_link.A {
			continue
		}
		_, existing := GetParam(params, param_link.Param)
		if existing != nil {
			continue
		}
		params = append(params, &FixedParameter{param_link.Param, nil, false})
	}
	return params
}

func (f *Frame) Title() string {
	tname := "nil"
	if t := f.Type(); t != nil {
		tname = t.Name()
	}
	return fmt.Sprintf("%s:%s", f.Name, tname)
}

func (f *Frame) Delete() {
	b := f.blueprint
	for link, _ := range b.links {
		if f == link.A || f == link.B {
			delete(b.links, link)
		}
	}
	delete(b.frames_, f)
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
	i, _ := GetParam(link.A.Parameters(), link.Param)
	return link.A.ParamCenter(i)
}

func (link Link) EndPos() Vec2 {
	if link.B == nil {
		return Pointer.Global
	}
	a := link.StartPos()
	return Ray(a, link.B.pos, link.B.size)
}

func (link *Link) Delete() {
	delete(link.Blueprint.links, link)
	for other, _ := range link.Blueprint.links {
		if other.A == link.A &&
			other.Param == link.Param &&
			other.Order > link.Order {
			other.Order--
		}
	}
}

func (link *Link) AppendWidget(widgets *Widgets) {
	line := widgets.Line(link.StartPos(), link.EndPos())
	params := link.A.Parameters()
	_, param := GetParam(params, link.Param)
	if param.Output() {
		line.Start = MakeCircle(param_r/4, "#000")
		line.End = MakeArrow()
	} else {
		line.Start = MakeArrow()
		line.End = MakeCircle(param_r/4, "#000")
	}
	line.Middle = MakeText(fmt.Sprint(link.Order))
	line.Middle.Scale = .75
}

func GetParam(params []Parameter, name string) (int, Parameter) {
	for i, param := range params {
		if param.Name() == name {
			return i, param
		}
	}
	return -1, nil
}
