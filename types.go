package mvm

type VM struct {
	blueprints []*Blueprint
	types      []*Type
}

type Blueprint struct {
	name     string
	elements []*BlueprintElement
	machines []*Machine
}

type BlueprintElement struct {
	blueprint *Blueprint
	index     int
	typ       *Type
	pos       Position
	global    bool
}

type Position struct {
	x, y float64
}

type Args map[string][]*MachineElement

type Type struct {
	name        string
	instantiate func(*MachineElement)
	run         func(Args)
}

type Machine struct {
	blueprint *Blueprint
	elements  []*MachineElement
}

type MachineElement struct {
	machine *Machine
	index   int
	object  interface{}
}

func MakeBlueprint(name string) *Blueprint {
	bp := &Blueprint{ name: name }
	bp.machines = append(bp.machines, &Machine{blueprint: bp})
	return bp
}

func (bp *Blueprint) Add(typ *Type, x, y float64, global bool) *BlueprintElement {
	elem := &BlueprintElement{
		typ:    typ,
		pos:    Position{x, y},
		global: global,
	}
	bp.elements = append(bp.elements, elem)
	return elem
}
