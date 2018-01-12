package mvm

import (
	"github.com/mafik/mvm/vec2"
)

/*
type Plug interface {
	GetObject(parent *Object) *Object
}

func (f *Frame) GetObject(parent *Object) *Object {
	if f.public {
		// try to find value in the blueprint above
		return parent.frame.GetObject(parent.parent)
	}
	return f.findObjectHere(parent)
}

func (f *Frame) findObjectHere(paret *Object) *Object {
}

type ElementPlug struct {
	TargetElement *FrameElement
}

type TreeNode struct {
	Outgoing Plug
	Incoming []Plug
	Stiff    bool
}
*/

type Frame struct {
	blueprint *Blueprint
	pos       vec2.Vec2
	size      vec2.Vec2
	name      string
	elems     []*FrameElement
	param     bool
	public    bool
	Hidden    bool
}

type TreeNode struct {
	Target       *Frame
	TargetMember string
	Stiff        bool
}

type FrameElement struct {
	Frame *Frame
	Name  string
	TreeNode
}

func (el *FrameElement) Index() int {
	for i, other := range el.Frame.elems {
		if other == el {
			return i
		}
	}
	return -1
}

func (f *Frame) FindElement(name string) *FrameElement {
	for i, elem := range f.elems {
		if elem.Name == name {
			return f.elems[i]
		}
	}
	return nil
}

func (f *Frame) GetElement(name string) *FrameElement {
	elem := f.FindElement(name)
	if elem == nil {
		f.elems = append(f.elems, &FrameElement{f, name, TreeNode{nil, "", false}})
		elem = f.elems[len(f.elems)-1]
	}
	return elem
}

func (f *Frame) Object(blueprint_instance *Object) *Object {
	return blueprint_instance.priv.(*Machine).objects[f]
}

func (f *Frame) Name() string {
	return f.name
}

func (f *Frame) Title() string {
	if f.param {
		return "¶" + f.name
	}
	return f.name
}

func (f *Frame) Delete() {
	b := f.blueprint
	X := 0
	for x, frame := range b.frames {
		if frame == f {
			X = x
		}
		for s, _ := range frame.elems {
			if f == frame.elems[s].Target {
				frame.elems[s].Target = nil
			}
		}
	}
	b.frames = append(b.frames[:X], b.frames[X+1:]...)
	for m, _ := range b.instances {
		delete(m.priv.(*Machine).objects, f)
	}
}

func (e *Frame) PropagateStiff(cb func(*Frame)) {
	visited := map[*Frame]bool{}
	for f, q := e, []*Frame{e}; len(q) > 0; f, q = q[len(q)-1], q[:len(q)-1] {
		if _, ok := visited[f]; ok {
			continue
		}
		visited[f] = true
		cb(f)
		for _, param := range f.elems {
			if param.Stiff && param.Target != nil {
				q = append(q, param.Target)
			}
		}
	}
}
