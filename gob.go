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
	gob.Register(ElementGob{})
	gob.Register(ShellGob{})
	for _, o := range Objects {
		gob.Register(o)
	}
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
	vm.root = d.Get(vmGob.ActiveIndex).(*Shell)
}

type BlueprintGob struct {
	Name      string
	Frames    []int
	Instances []int
	Transform matrix.Matrix
	Color     int
}

func (blue *Blueprint) Gob(s Serializer) Gob {
	gob := BlueprintGob{Name: blue.name, Transform: blue.transform, Color: blue.color}
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
		instances: make(map[*Shell]bool),
		transform: gob.Transform,
		color:     gob.Color,
	}
}

func (blue *Blueprint) Connect(d Deserializer, gob Gob) {
	blueGob := gob.(BlueprintGob)
	blue.name = blueGob.Name
	for _, i := range blueGob.Frames {
		blue.frames = append(blue.frames, d.Get(i).(*Frame))
	}
	for _, i := range blueGob.Instances {
		blue.instances[d.Get(i).(*Shell)] = true
	}
}

type FrameGob struct {
	Blueprint  int
	Pos        Vec2
	Size       Vec2
	Name       string
	Elems      []int
	Param      bool
	Public     bool
	ShowWindow bool
}

func (frame *Frame) Gob(s Serializer) Gob {
	gob := FrameGob{
		Blueprint:  s.Id(frame.blueprint),
		Pos:        frame.pos,
		Size:       frame.size,
		Name:       frame.name,
		Elems:      nil,
		Param:      frame.param,
		Public:     frame.public,
		ShowWindow: frame.ShowWindow,
	}
	for _, frame_element := range frame.elems {
		gob.Elems = append(gob.Elems, s.Id(frame_element))
	}
	return gob
}

func (gob FrameGob) Ungob() Gobbable {
	return &Frame{nil, gob.Pos, gob.Size, gob.Name, nil, gob.Param, gob.Public, false, gob.ShowWindow}
}

func (frame *Frame) Connect(d Deserializer, gob Gob) {
	frameGob := gob.(FrameGob)
	frame.blueprint = d.Get(frameGob.Blueprint).(*Blueprint)
	for _, elemId := range frameGob.Elems {
		frame.elems = append(frame.elems, d.Get(elemId).(*FrameElement))
	}
}

type ElementGob struct {
	Frame  int
	Name   string
	Target int
	Stiff  bool
}

func (e *FrameElement) Gob(s Serializer) Gob {
	return ElementGob{
		Frame:  s.Id(e.frame),
		Name:   e.Name,
		Target: s.Id(e.Target),
		Stiff:  e.Stiff,
	}
}

func (gob ElementGob) Ungob() Gobbable {
	return &FrameElement{TreeNode: TreeNode{nil, gob.Stiff}, frame: nil, Name: gob.Name}
}

func (e *FrameElement) Connect(d Deserializer, gob Gob) {
	elementGob := gob.(ElementGob)
	e.Target, _ = d.Get(elementGob.Target).(Target)
	e.frame = d.Get(elementGob.Frame).(*Frame)
}

type MachineGob struct {
	Blueprint int
	Shells    map[int]int
}

func (mach *Machine) Gob(s Serializer) Gob {
	gob := MachineGob{Blueprint: s.Id(mach.Blueprint), Shells: make(map[int]int)}
	for frame, shell := range mach.shells {
		gob.Shells[s.Id(frame)] = s.Id(shell)
	}
	return gob
}

func (gob MachineGob) Ungob() Gobbable { return &Machine{nil, make(map[*Frame]*Shell)} }

func (m *Machine) Connect(d Deserializer, gob Gob) {
	mGob := gob.(MachineGob)
	m.Blueprint = d.Get(mGob.Blueprint).(*Blueprint)
	for f, s := range mGob.Shells {
		m.shells[d.Get(f).(*Frame)] = d.Get(s).(*Shell)
	}
}

type ShellGob struct {
	Parent  int
	Frame   int
	Execute bool
	Object  interface{}
}

func (shell *Shell) Gob(s Serializer) Gob {
	gob := ShellGob{
		Execute: shell.execute,
	}
	if gobbable, ok := shell.object.(Gobbable); ok {
		gob.Object = gobbable.Gob(s)
	} else {
		gob.Object = shell.object
	}
	if shell.parent != nil {
		gob.Parent = s.Id(shell.parent)
	}
	if shell.frame != nil {
		gob.Frame = s.Id(shell.frame)
	}
	return gob
}

func (gob ShellGob) Ungob() Gobbable {
	s := &Shell{
		execute: gob.Execute,
	}
	if objGob, ok := gob.Object.(Gob); ok {
		s.object = objGob.Ungob().(Object)
	} else {
		s.object = gob.Object.(Object)
	}
	return s
}

func (shell *Shell) Connect(d Deserializer, gob Gob) {
	objGob := gob.(ShellGob)
	if parent, ok := d.Get(objGob.Parent).(*Shell); ok {
		shell.parent = parent
	}
	if frame, ok := d.Get(objGob.Frame).(*Frame); ok {
		shell.frame = frame
	}
	if gobbable, ok := shell.object.(Gobbable); ok {
		gobbable.Connect(d, objGob.Object.(Gob))
	}
}
