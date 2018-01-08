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
	Members() []string
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
