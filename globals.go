package mvm

import (
	"bytes"
	"fmt"
	"os/exec"
)

type FixedParameter struct {
	name string
	typ  Type
}

func (p *FixedParameter) Name() string {
	return p.name
}

func (p *FixedParameter) Typ() Type {
	return p.typ
}

type PrimitiveType struct {
	name        string
	parameters  []Parameter
	instantiate func(*Object)
	copy        func(from, to *Object)
	run         func(Args)
	string      func(interface{}) string
	draw        func(*Object, *Context2D)
	input       func(*Object, *Touch, Event) Touching
}

func (t *PrimitiveType) Name() string {
	return t.name
}

func (t *PrimitiveType) Parameters() []Parameter {
	return t.parameters
}

func (t *PrimitiveType) Instantiate(o *Object) {
	if t.instantiate != nil {
		t.instantiate(o)
	}
}

func (t *PrimitiveType) Copy(from, to *Object) {
	if t.copy != nil {
		t.copy(from, to)
	}
}

func (t *PrimitiveType) Run(args Args) {
	t.run(args)
}

func (t *PrimitiveType) String(i interface{}) string {
	if t.string != nil {
		return t.string(i)
	} else {
		return fmt.Sprintf("%#v", i)
	}
}

func (t *PrimitiveType) Draw(o *Object, c *Context2D) {
	if t.draw != nil {
		t.draw(o, c)
	}
}

func (t *PrimitiveType) Input(o *Object, touch *Touch, e Event) Touching {
	if t.input != nil {
		return t.input(o, touch, e)
	}
	return nil
}

var TextType Type = &PrimitiveType{
	name: "text",
	instantiate: func(me *Object) {
		me.priv = []byte{}
	},
	copy: func(from, to *Object) {
		to.priv = append([]byte{}, from.priv.([]byte)...)
	},
	string: func(i interface{}) string {
		return string(i.([]byte))
	},
}

var FormatType Type = &PrimitiveType{
	name: "format",
	run: func(args Args) {
		format := string(args["fmt"].priv.([]byte))
		fmt_args := []interface{}{args["args"].priv}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, format, fmt_args...)
		args["output"].priv = buf.Bytes()
	},
	parameters: []Parameter{
		&FixedParameter{name: "output", typ: TextType},
		&FixedParameter{name: "fmt", typ: TextType},
		&FixedParameter{name: "args"},
	},
}

var ExecType Type = &PrimitiveType{
	name: "exec",
	run: func(args Args) {
		name := TextType.String(args["command"].priv)
		cmd_args := []string{}
		if args["args"] != nil {
			cmd_args = append(cmd_args, TextType.String(args["args"].priv))
		}
		out, err := exec.Command(name, cmd_args...).Output()
		if err != nil {
			if args["stderr"] != nil {
				switch err := err.(type) {
				case *exec.ExitError:
					args["stderr"].priv = err.Stderr
				case *exec.Error:
					args["stderr"].priv = []byte(err.Error())
				}
			}
			return
		}
		args["stdout"].priv = out
	},
	parameters: []Parameter{
		&FixedParameter{name: "command", typ: TextType},
		&FixedParameter{name: "args", typ: TextType},
		&FixedParameter{name: "stdout", typ: TextType},
		&FixedParameter{name: "stderr", typ: TextType},
	},
}

var Types map[string]Type = map[string]Type{
	"format": FormatType,
	"text":   TextType,
	"exec":   ExecType,
}

var TheVM *VM = &VM{}
