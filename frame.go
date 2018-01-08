package mvm

import (
	"github.com/mafik/mvm/vec2"
)

type Frame struct {
	blueprint *Blueprint
	pos       vec2.Vec2
	size      vec2.Vec2
	name      string
	elems     []FrameElement
	param     bool
	public    bool
	Hidden    bool
}

type FrameElement struct {
	Name         string
	Target       *Frame
	TargetMember string
	Stiff        bool
}

func (f *Frame) FindElement(name string) *FrameElement {
	for i, elem := range f.elems {
		if elem.Name == name {
			return &f.elems[i]
		}
	}
	return nil
}

func (f *Frame) GetElement(name string) *FrameElement {
	elem := f.FindElement(name)
	if elem == nil {
		f.elems = append(f.elems, FrameElement{name, nil, "", false})
		elem = &f.elems[len(f.elems)-1]
	}
	return elem
}

func (f *Frame) Object(blueprint_instance *Object) *Object {
	return blueprint_instance.priv.(*Machine).objects[f]
}

func (f *Frame) Name() string {
	return f.name
}

func (f *Frame) Type() Type { // part of Parameter interface
	return nil // TODO: implement
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
