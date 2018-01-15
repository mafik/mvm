package mvm

import (
	"github.com/mafik/mvm/ui"
)

type VM struct {
	root *Object
}

type Args interface {
	Get(string) *Object
	Set(string, *Object)
}

type Parameter interface {
	Name() string
}

type Member interface {
	Name() string
}

type Type interface {
	Name() string
	Parameters() []Parameter
	Members() []Member
	GetMember(*Object, string) *Object
	Instantiate(*Object) // TODO: remove
	Copy(from, to *Object)
	Run(Args)
	String(interface{}) string
	MakeWidget(*Object) ui.Widget
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
