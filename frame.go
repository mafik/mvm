package mvm

import (
	"github.com/mafik/mvm/vec2"
)

type Target interface {
	Gobbable
	Frame() *Frame
	Position() vec2.Vec2
}

type TreeNode struct {
	Target Target
	Stiff  bool
}

type Frame struct {
	blueprint  *Blueprint
	pos        vec2.Vec2
	size       vec2.Vec2
	name       string
	elems      []*FrameElement
	param      bool
	public     bool
	Hidden     bool
	ShowWindow bool
}

type FrameElement struct {
	TreeNode
	frame *Frame
	Name  string
}

func (el *FrameElement) Frame() *Frame { return el.frame }
func (el *FrameElement) Position() vec2.Vec2 {
	v := el.frame.pos
	return vec2.Add(v, el.frame.ParamCenter(el.Index()))
}
func (el *FrameElement) Index() int {
	for i, other := range el.frame.elems {
		if other == el {
			return i
		}
	}
	return -1
}

func (f *Frame) Frame() *Frame       { return f }
func (f *Frame) Position() vec2.Vec2 { return f.pos }
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
		f.elems = append(f.elems, &FrameElement{TreeNode{nil, false}, f, name})
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
				q = append(q, param.Target.Frame())
			}
		}
	}
}
