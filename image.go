package mvm

import (
	"errors"
	"fmt"
	"os"
)

var TextType Type = Type{
	name: "text",
	instantiate: func(me *MachineElement) {
		me.object = make([]byte, 0)
	},
}

var FileName string = "mvm.img"
var TheVM VM = VM{
	types: []*Type{&TextType},
}

func SetupDefault() {
	welcome := MakeBlueprint("welcome")
	welcome.Add(&TextType, 0, 0, true)
	TheVM.blueprints = append(TheVM.blueprints, welcome)
}

func LoadImage() error {
	if _, err := os.Stat(FileName); os.IsNotExist(err) {
		fmt.Printf("'%s' not found - loading the default VM image\n", FileName)
		SetupDefault()
		return nil
	}
	return errors.New("Not implemented")
}
