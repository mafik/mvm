package mvm

import (
	"github.com/mafik/mvm/ui"
)

type VM struct {
	root *Shell
}

type Args interface {
	Get(string) *Shell
	Set(string, *Shell)
}

type Parameter interface {
	Name() string
}

type Member interface {
	Name() string
}

type Object interface {
	Name() string
}

type StatefulObject interface {
	Object
	Copy(*Shell)
	// TODO: Destroy()
}

type ComplexObject interface {
	Object
	Members() []Member
	GetMember(string) *Shell
}

type RunnableObject interface {
	Object
	Parameters() []Parameter
	Run(Args)
}

type GraphicObject interface {
	Object
	MakeWidget(*Shell) ui.Widget
}

func GetParam(params []Parameter, name string) (int, Parameter) {
	for i, param := range params {
		if param.Name() == name {
			return i, param
		}
	}
	return -1, nil
}

func GetMember(members []Member, name string) (int, Member) {
	for i, member := range members {
		if member.Name() == name {
			return i, member
		}
	}
	return -1, nil
}
