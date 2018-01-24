package mvm

import (
	"github.com/mafik/mvm/vec2"
)

type Target interface {
	Gobbable
	Frame() *Frame
	Position() vec2.Vec2
	Get(blueprint *Shell) *Shell
	Set(blueprint *Shell, value *Shell)
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
	return vec2.Add(el.frame.pos, el.PositionInFrame())
}
func (el *FrameElement) PositionInFrame() vec2.Vec2 {
	return el.frame.ParamCenter(el.Index())
}
func (el *FrameElement) Get(blueprint *Shell) *Shell {
	shell := el.frame.Get(blueprint)
	return shell.object.GetMember(shell, el.Name)
}
func (el *FrameElement) Set(blueprint *Shell, value *Shell) {
	//machine := blueprint.priv.(*Machine)
	//shell := machine.shells[el.frame]
	// TODO: shell.object.SetMember(shell, value)
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
func (f *Frame) Get(blueprint *Shell) *Shell {
	machine := blueprint.priv.(*Machine)
	return machine.shells[f]
}
func (f *Frame) Set(blueprint *Shell, value *Shell) {
	value.parent = blueprint
	value.frame = f
	machine := blueprint.priv.(*Machine)
	machine.shells[f] = value
}

type ElementPack struct {
	FrameElement *FrameElement
	Param        Parameter
	Member       Member
}

func (f *Frame) ZipElements(s *Shell) (zip []ElementPack) {
	var params []Parameter
	var members []Member
	if s != nil {
		members = s.object.Members()
		params = s.object.Parameters()
	}
	for _, el := range f.elems {
		i, param := GetParam(params, el.Name)
		if i >= 0 {
			last := len(params) - 1
			params[i], params[last] = params[last], params[i]
			params = params[:last]
		}
		i, member := GetMember(members, el.Name)
		if i >= 0 {
			last := len(members) - 1
			members[i], members[last] = members[last], members[i]
			members = members[:last]
		}
		zip = append(zip, ElementPack{el, param, member})
	}
	for _, param := range params {
		zip = append(zip, ElementPack{nil, param, nil})
	}
	for _, member := range members {
		zip = append(zip, ElementPack{nil, nil, member})
	}
	return
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
		f.elems = append(f.elems, &FrameElement{TreeNode{nil, false}, f, name})
		elem = f.elems[len(f.elems)-1]
	}
	return elem
}

func (f *Frame) Shell(blueprint_instance *Shell) *Shell {
	return blueprint_instance.priv.(*Machine).shells[f]
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
		delete(m.priv.(*Machine).shells, f)
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
