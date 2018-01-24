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
	Parameters() []Parameter
	Members() []Member
	GetMember(*Shell, string) *Shell
	Instantiate(*Shell) // TODO: remove
	Copy(from, to *Shell)
	Run(Args)
	String(interface{}) string
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
