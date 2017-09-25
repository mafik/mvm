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
	//fmt.Printf("Looking up ID for: %T %v\n", ble, ble)
	id, ok := f.ids[ble]
	if !ok {
		id = len(f.gobbables)
		//fmt.Printf("Not found - adding as %d\n", id)
		f.gobbables = append(f.gobbables, ble)
		f.ids[ble] = id
	} else {
		//fmt.Printf("Found at %d\n", id)
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
	f.ids[nil] = 0
	f.gobbables = append(f.gobbables, nil)
	f.Id(ble) // the main Gobbable is saved at 1
	var gobs []Gob
	gobs = append(gobs, nil)
	for i := 1; i < len(f.gobbables); i++ {
		b := f.gobbables[i]
		//fmt.Printf("\nProcessing %d: %T %v\n", i, b, b)
		gobs = append(gobs, b.Gob(&f))
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

	gobbables := Gobbables{}
	for _, gob := range gobs {
		var ble Gobbable
		if gob != nil {
			ble = gob.Ungob()
		}
		gobbables = append(gobbables, ble)
	}
	for i, ble := range gobbables {
		//fmt.Printf("Connecting gobbable %d\n", i)
		if ble != nil {
			ble.Connect(gobbables, gobs[i])
		}
	}
	return gobbables[1], nil
}

func (slice Gobbables) Get(i int) Gobbable { return slice[i] }

type VMGob struct {
	ActiveIndex int
}

func (vm *VM) Gob(s Serializer) Gob {
	return VMGob{ActiveIndex: s.Id(vm.root)}
}

func (gob VMGob) Ungob() Gobbable { return &VM{} }

func (vm *VM) Connect(d Deserializer, gob Gob) {
	vmGob := gob.(VMGob)
	vm.root = d.Get(vmGob.ActiveIndex).(*Object)
}

type BlueprintGob struct {
	Name      string
	Frames    []int
	Instances []int
}

func (blue *Blueprint) Gob(s Serializer) Gob {
	gob := BlueprintGob{Name: blue.name}
	for frame, _ := range blue.frames {
		gob.Frames = append(gob.Frames, s.Id(frame))
	}
	for instance, _ := range blue.instances {
		gob.Instances = append(gob.Instances, s.Id(instance))
	}
	return gob
}

func (gob BlueprintGob) Ungob() Gobbable {
	return &Blueprint{
		name:      gob.Name,
		frames:    make(map[*Frame]bool),
		instances: make(map[*Object]bool),
	}
}

func (blue *Blueprint) Connect(d Deserializer, gob Gob) {
	blueGob := gob.(BlueprintGob)
	blue.name = blueGob.Name
	for _, i := range blueGob.Frames {
		blue.frames[d.Get(i).(*Frame)] = true
	}
	for _, i := range blueGob.Instances {
		blue.instances[d.Get(i).(*Object)] = true
	}
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
	Parent        int
	Frame         int
	Execute       bool
	PrimitiveType string
	BlueprintType int
	Private       interface{}
}

func (obj *Object) Gob(s Serializer) Gob {
	gob := ObjectGob{
		Execute:       obj.execute,
		PrimitiveType: "",
		Private:       obj.priv,
	}
	if obj.parent != nil {
		gob.Parent = s.Id(obj.parent)
	}
	if obj.frame != nil {
		gob.Frame = s.Id(obj.frame)
	}
	if blue, ok := obj.typ.(*Blueprint); ok {
		gob.BlueprintType = s.Id(blue)
		gob.Private = gob.Private.(*Machine).Gob(s)
	} else {
		gob.PrimitiveType = obj.typ.Name()
	}
	return gob
}

func (gob ObjectGob) Ungob() Gobbable {
	o := &Object{
		execute: gob.Execute,
	}
	if gob.BlueprintType == 0 {
		o.priv = gob.Private
	} else {
		o.priv = gob.Private.(MachineGob).Ungob()
	}
	return o
}

func (obj *Object) Connect(d Deserializer, gob Gob) {
	objGob := gob.(ObjectGob)
	if parent, ok := d.Get(objGob.Parent).(*Object); ok {
		obj.parent = parent
	}
	if frame, ok := d.Get(objGob.Frame).(*Frame); ok {
		obj.frame = frame
	}
	if objGob.BlueprintType == 0 {
		obj.typ = Types[objGob.PrimitiveType]
	} else {
		obj.typ = d.Get(objGob.BlueprintType).(*Blueprint)
		obj.priv.(*Machine).Connect(d, objGob.Private.(MachineGob))
	}
}
