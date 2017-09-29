package mvm

import (
	"fmt"
	"io/ioutil"
	"os"
)

func SetupDefault() {
	welcome := MakeBlueprint("welcome")
	TheVM.active = MakeObject(welcome, nil, nil)
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
