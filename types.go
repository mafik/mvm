package mvm

import (
	"fmt"
	"sort"
)

type VM struct {
	Blueprints      map[*Blueprint]bool
	ActiveBlueprint *Blueprint
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

type Blueprint struct {
	name           string
	frames         map[*Frame]bool
	links          map[*Link]bool
	machines       map[*Machine]bool
	active_machine *Machine
}

type ByName []*Blueprint

func (b ByName) Len() int           { return len(b) }
func (b ByName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByName) Less(i, j int) bool { return b[i].name < b[j].name }

type Link struct {
	Blueprint *Blueprint
	A, B      *Frame
	Param     int
	Order     int
}

type Frame struct {
	blueprint *Blueprint
	typ       *Type
	pos       Vec2
	size      Vec2
	global    bool
}

type Args map[string][]*Object

type Parameter struct {
	Name     string
	Typ      *Type
	Runnable bool
	Output   bool
}

type Type struct {
	Name        string
	Parameters  []Parameter
	Instantiate func(*Object)
	Run         func(Args)
	String      func(interface{}) string
}

type Machine struct {
	blueprint *Blueprint
	objects   map[*Frame]*Object
}

type Object struct {
	machine *Machine
	frame   *Frame
	execute bool
	Running bool
	priv    interface{}
}

func (o *Object) MarkForExecution() {
	o.execute = true
	tasks <- o
}

func MakeBlueprint(name string) *Blueprint {
	bp := &Blueprint{
		name:   name,
		frames: make(map[*Frame]bool),
		links:  make(map[*Link]bool),
	}
	bp.active_machine = &Machine{
		blueprint: bp,
		objects:   make(map[*Frame]*Object),
	}
	bp.machines = map[*Machine]bool{bp.active_machine: true}
	return bp
}

func (bp *Blueprint) Add(typ *Type) *Frame {
	frame := &Frame{
		blueprint: bp,
		typ:       typ,
		pos:       Vec2{0, 0},
		size:      Vec2{100, 100},
		global:    true,
	}
	bp.frames[frame] = true
	m := bp.active_machine
	object := &Object{
		machine: m,
		frame:   frame,
	}
	if typ.Instantiate != nil {
		typ.Instantiate(object)
	}
	m.objects[frame] = object
	return frame
}

func (be *Frame) Object() *Object {
	return be.blueprint.active_machine.objects[be]
}

func (f *Frame) Delete() {
	b := f.blueprint
	for link, _ := range b.links {
		if f == link.A || f == link.B {
			delete(b.links, link)
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
	return link.A.ParamCenter(link.Param)
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
	if link.A.typ.Parameters[link.Param].Output {
		line.Start = MakeCircle(param_r/4, "#000")
		line.End = MakeArrow()
	} else {
		line.Start = MakeArrow()
		line.End = MakeCircle(param_r/4, "#000")
	}
	line.Middle = MakeText(fmt.Sprint(link.Order))
	line.Middle.Scale = .75
}
