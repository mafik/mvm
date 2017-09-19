package mvm

import (
	"bytes"
	"encoding/gob"
)

type Serializer interface {
	Id(Gobbable) int
}

type Deserializer interface {
	Get(int) Gobbable
}

type Gobbable interface {
	Gob(Serializer) Gob
	Connect(Deserializer, Gob)
}

type Gob interface {
	Ungob() Gobbable
}

type Flattener struct {
	ids       map[Gobbable]int
	gobbables []Gobbable
}

func (f *Flattener) Id(ble Gobbable) (id int) {
	id, ok := f.ids[ble]
	if !ok {
		id = len(f.gobbables)
		f.gobbables = append(f.gobbables, ble)
		f.ids[ble] = id
	}
	return
}

func (f *Flattener) Get(i int) Gobbable {
	return f.gobbables[i]
}

func RegisterGobs() {
	gob.Register(VMGob{})
	gob.Register(BlueprintGob{})
	gob.Register(MachineGob{})
	gob.Register(FrameGob{})
	gob.Register(ObjectGob{})
}

func Flatten(ble Gobbable) ([]byte, error) {
	RegisterGobs()
	f := Flattener{ids: make(map[Gobbable]int)}
	f.Id(ble) // the main Gobbable is saved at 0
	var gobs []Gob
	for i := 0; i < len(f.gobbables); i++ {
		ble = f.gobbables[i]
		gobs = append(gobs, ble.Gob(&f))
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(gobs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type Gobbables []Gobbable

func Unflatten(data []byte) (Gobbable, error) {
	RegisterGobs()
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var gobs []Gob
	err := dec.Decode(&gobs)
	if err != nil {
		return nil, err
	}

	var gobbables Gobbables
	for _, gob := range gobs {
		gobbables = append(gobbables, gob.Ungob())
	}
	for i, ble := range gobbables {
		ble.Connect(gobbables, gobs[i])
	}
	return gobbables[0], nil
}

func (slice Gobbables) Get(i int) Gobbable { return slice[i] }

type VMGob struct {
	Blueprints  []int
	ActiveIndex int
}

func (vm *VM) Gob(s Serializer) Gob {
	gob := VMGob{ActiveIndex: s.Id(vm.ActiveBlueprint)}
	for blue, _ := range vm.Blueprints {
		gob.Blueprints = append(gob.Blueprints, s.Id(blue))
	}
	return gob
}

func (gob VMGob) Ungob() Gobbable { return &VM{Blueprints: make(map[*Blueprint]bool)} }

func (vm *VM) Connect(d Deserializer, gob Gob) {
	vmGob := gob.(VMGob)
	for _, i := range vmGob.Blueprints {
		vm.Blueprints[d.Get(i).(*Blueprint)] = true
	}
	vm.ActiveBlueprint = d.Get(vmGob.ActiveIndex).(*Blueprint)
}

type BlueprintGob struct {
	Name          string
	Frames        []int
	Machines      []int
	ActiveMachine int
}

func (blue *Blueprint) Gob(s Serializer) Gob {
	gob := BlueprintGob{Name: blue.name, ActiveMachine: s.Id(blue.active_machine)}
	for frame, _ := range blue.frames_ {
		gob.Frames = append(gob.Frames, s.Id(frame))
	}
	for mach, _ := range blue.machines {
		gob.Machines = append(gob.Machines, s.Id(mach))
	}
	return gob
}

func (gob BlueprintGob) Ungob() Gobbable {
	return &Blueprint{
		name:     gob.Name,
		frames_:  make(map[*Frame]bool),
		machines: make(map[*Machine]bool),
	}
}

func (blue *Blueprint) Connect(d Deserializer, gob Gob) {
	blueGob := gob.(BlueprintGob)
	blue.name = blueGob.Name
	for _, i := range blueGob.Frames {
		blue.frames_[d.Get(i).(*Frame)] = true
	}
	for _, i := range blueGob.Machines {
		blue.machines[d.Get(i).(*Machine)] = true
	}
	blue.active_machine = d.Get(blueGob.ActiveMachine).(*Machine)
}

type FrameGob struct {
	Blueprint int
	Pos       Vec2
	Size      Vec2
	Name      string
	LinkSets  []LinkSetGob
}

type LinkSetGob struct {
	ParamName string
	Targets   []int
}

func (frame *Frame) Gob(s Serializer) Gob {
	gob := FrameGob{
		Blueprint: s.Id(frame.blueprint),
		Pos:       frame.pos,
		Size:      frame.size,
		Name:      frame.name,
		LinkSets:  nil,
	}
	for _, link_set := range frame.link_sets {
		ids := []int{}
		for _, target := range link_set.Targets {
			ids = append(ids, s.Id(target))
		}
		gob.LinkSets = append(gob.LinkSets, LinkSetGob{link_set.ParamName, ids})
	}
	return gob
}

func (gob FrameGob) Ungob() Gobbable {
	return &Frame{nil, gob.Pos, gob.Size, gob.Name, nil}
}

func (frame *Frame) Connect(d Deserializer, gob Gob) {
	frameGob := gob.(FrameGob)
	frame.blueprint = d.Get(frameGob.Blueprint).(*Blueprint)
	for _, links_gob := range frameGob.LinkSets {
		targets := []*Frame{}
		for _, target := range links_gob.Targets {
			targets = append(targets, d.Get(target).(*Frame))
		}
		frame.link_sets = append(frame.link_sets, LinkSet{links_gob.ParamName, targets})
	}
}

type MachineGob struct {
	Blueprint int
	Objects   map[int]int
}

func (mach *Machine) Gob(s Serializer) Gob {
	gob := MachineGob{Objects: make(map[int]int)}
	for frame, obj := range mach.objects {
		gob.Objects[s.Id(frame)] = s.Id(obj)
	}
	return gob
}

func (gob MachineGob) Ungob() Gobbable { return &Machine{make(map[*Frame]*Object)} }

func (m *Machine) Connect(d Deserializer, gob Gob) {
	mGob := gob.(MachineGob)
	for f, o := range mGob.Objects {
		m.objects[d.Get(f).(*Frame)] = d.Get(o).(*Object)
	}
}

type ObjectGob struct {
	Machine       int
	Frame         int
	Execute       bool
	PrimitiveType string
	BlueprintType int
	Private       interface{}
}

func (obj *Object) Gob(s Serializer) Gob {
	priv := obj.priv
	if machine, ok := priv.(*Machine); ok {
		priv = s.Id(machine)
	}
	gob := ObjectGob{
		Machine:       s.Id(obj.machine),
		Frame:         s.Id(obj.frame),
		Execute:       obj.execute,
		PrimitiveType: "",
		BlueprintType: -1,
		Private:       priv,
	}
	if blue, ok := obj.typ.(*Blueprint); ok {
		gob.BlueprintType = s.Id(blue)
	} else {
		gob.PrimitiveType = obj.typ.Name()
	}
	return gob
}

func (gob ObjectGob) Ungob() Gobbable {
	o := &Object{
		execute: gob.Execute,
		priv:    gob.Private,
	}
	return o
}

func (obj *Object) Connect(d Deserializer, gob Gob) {
	objGob := gob.(ObjectGob)
	obj.machine = d.Get(objGob.Machine).(*Machine)
	obj.frame = d.Get(objGob.Frame).(*Frame)
	if objGob.BlueprintType == -1 {
		obj.typ = Types[objGob.PrimitiveType]
	} else {
		obj.typ = d.Get(objGob.BlueprintType).(*Blueprint)
		obj.priv = d.Get(objGob.Private.(int))
	}
}
