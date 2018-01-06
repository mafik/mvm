package mvm

import (
	"bytes"
	"encoding/gob"

	"github.com/mafik/mvm/matrix"
	. "github.com/mafik/mvm/vec2"
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
	return VMGob{ActiveIndex: s.Id(vm.active)}
}

func (gob VMGob) Ungob() Gobbable { return &VM{} }

func (vm *VM) Connect(d Deserializer, gob Gob) {
	vmGob := gob.(VMGob)
	vm.active = d.Get(vmGob.ActiveIndex).(*Object)
}

type BlueprintGob struct {
	Name      string
	Frames    []int
	Instances []int
	Transform matrix.Matrix
}

func (blue *Blueprint) Gob(s Serializer) Gob {
	gob := BlueprintGob{Name: blue.name, Transform: blue.transform}
	for _, frame := range blue.frames {
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
		frames:    nil,
		instances: make(map[*Object]bool),
		transform: gob.Transform,
	}
}

func (blue *Blueprint) Connect(d Deserializer, gob Gob) {
	blueGob := gob.(BlueprintGob)
	blue.name = blueGob.Name
	for _, i := range blueGob.Frames {
		blue.frames = append(blue.frames, d.Get(i).(*Frame))
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
	Params    []FrameParameterGob
	Param     bool
}

type FrameParameterGob struct {
	Name   string
	Target int
	Stiff  bool
}

func (frame *Frame) Gob(s Serializer) Gob {
	gob := FrameGob{
		Blueprint: s.Id(frame.blueprint),
		Pos:       frame.pos,
		Size:      frame.size,
		Name:      frame.name,
		Params:    nil,
		Param:     frame.param,
	}
	for _, frame_parameter := range frame.params {
		id := 0
		if frame_parameter.Target != nil {
			id = s.Id(frame_parameter.Target)
		}
		gob.Params = append(gob.Params, FrameParameterGob{frame_parameter.Name, id, frame_parameter.Stiff})
	}
	return gob
}

func (gob FrameGob) Ungob() Gobbable {
	return &Frame{nil, gob.Pos, gob.Size, gob.Name, nil, gob.Param, false}
}

func (frame *Frame) Connect(d Deserializer, gob Gob) {
	frameGob := gob.(FrameGob)
	frame.blueprint = d.Get(frameGob.Blueprint).(*Blueprint)
	for _, links_gob := range frameGob.Params {
		target, _ := d.Get(links_gob.Target).(*Frame)
		frame.params = append(frame.params, FrameParameter{links_gob.Name, target, links_gob.Stiff})
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
