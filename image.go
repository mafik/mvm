package mvm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

func MakeEcho(b *Blueprint, pos Vec2, text string) (*Frame, *Frame) {
	t := b.Add(TextType)
	t.pos = Add(pos, Vec2{-50, 0})
	t.size = Vec2{100, 50}
	var buf bytes.Buffer
	fmt.Fprint(&buf, text)
	t.Object(TheVM.root).priv = buf.Bytes()

	exec := b.Add(ExecType)
	exec.pos = Add(pos, Vec2{100, -180})
	exec.size = Vec2{100, 50}

	exec.AddLink("unknown", t)
	return exec, t
}

func SetupDefault() {
	welcome := MakeBlueprint("welcome")
	TheVM.root = MakeObject(welcome, nil)
	welcome.instances[TheVM.root] = true
	MakeEcho(welcome, Vec2{-100, -50}, "Hello")
	MakeEcho(welcome, Vec2{100, 50}, "world!")
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
