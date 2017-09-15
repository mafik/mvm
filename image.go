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
	t.Object().priv = buf.Bytes()

	exec := b.Add(ExecType)
	exec.pos = Add(pos, Vec2{100, -50})
	exec.size = Vec2{100, 50}

	b.links[&Link{b, exec, t, 0, 0}] = true
	return exec, t
}

func SetupDefault() {
	welcome := MakeBlueprint("welcome")
	MakeEcho(welcome, Vec2{-100, -50}, "Hello")
	MakeEcho(welcome, Vec2{100, 50}, "world!")

	TheVM.Blueprints[welcome] = true
	TheVM.ActiveBlueprint = welcome
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
