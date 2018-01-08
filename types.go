package mvm

import (
	"github.com/mafik/mvm/ui"
)

type VM struct {
	root *Object
}

type Args map[string]*Object

type Parameter interface {
	Name() string
	Type() Type
}

type Type interface {
	Name() string
	Parameters() []Parameter
	Instantiate(*Object) // TODO: remove
	Copy(from, to *Object)
	Run(Args)
	String(interface{}) string
	MakeWidget(*Object) ui.Widget
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
		for s, _ := range frame.params {
			if f == frame.params[s].Target {
				frame.params[s].Target = nil
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
		for _, param := range f.params {
			if param.Stiff && param.Target != nil {
				q = append(q, param.Target)
			}
		}
	}
}

func GetParam(params []Parameter, name string) (int, Parameter) {
	for i, param := range params {
		if param.Name() == name {
			return i, param
		}
	}
	return -1, nil
}
