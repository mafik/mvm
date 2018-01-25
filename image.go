package mvm

import (
	"fmt"
	"io/ioutil"
	"os"
)

func SetupDefault() {
	b := MakeBlueprint("root")
	s := MakeShell(nil, nil)
	b.instances[s] = true
	s.object = MakeMachine(b)
	TheVM.root = s

	var x float64
	for _, o := range Objects {
		f := b.AddFrame()
		Copy(o, f, s)
		f.pos.X = x
		x += f.size.X + margin
	}
}

var FileName string = "mvm.img"

func LoadImage() error {
	byte_slice, err := ioutil.ReadFile(FileName)
	if err != nil {
		fmt.Printf("'%s' not found - loading the default VM image\n", FileName)
		SetupDefault()
		return nil
	}
	ble, err := Unflatten(byte_slice)
	if err != nil {
		return err
	}
	TheVM = ble.(*VM)
	return nil
}

func SaveImage() error {
	bytes, err := Flatten(TheVM)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(FileName, bytes, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
