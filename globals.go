package mvm

import (
	"bytes"
	"fmt"
	"os/exec"
)

type FixedParameter struct {
	name   string
	typ    Type
	output bool
}

func (p *FixedParameter) Name() string {
	return p.name
}

func (p *FixedParameter) Typ() Type {
	return p.typ
}

func (p *FixedParameter) Output() bool {
	return p.output
}

type PrimitiveType struct {
	name        string
	parameters  []Parameter
	instantiate func(*Object)
	run         func(Args)
	string      func(interface{}) string
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

var TextType Type = &PrimitiveType{
	name: "text",
	instantiate: func(me *Object) {
		var b bytes.Buffer
		fmt.Fprint(&b, "Hello world!")
		me.priv = b.Bytes()
	},
	string: func(i interface{}) string {
		return string(i.([]byte))
	},
}

var FormatType Type = &PrimitiveType{
	name: "format",
	run: func(args Args) {
		format := string(args["fmt"][0].priv.([]byte))
		var fmt_args []interface{}
		for _, o := range args["args"] {
			fmt_args = append(fmt_args, o.priv)
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, format, fmt_args...)
		args["output"][0].priv = buf.Bytes()
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
		name := TextType.String(args["command"][0].priv)
		var cmd_args []string
		for _, o := range args["args"] {
			cmd_args = append(cmd_args, TextType.String(o.priv))
		}
		out, err := exec.Command(name, cmd_args...).Output()
		if err != nil {
			if args["stderr"] != nil {
				args["stderr"][0].priv = err.(*exec.ExitError).Stderr
			}
			return
		}
		args["stdout"][0].priv = out
	},
	parameters: []Parameter{
		&FixedParameter{name: "command", typ: TextType},
		&FixedParameter{name: "args", typ: TextType},
		&FixedParameter{name: "stdout", typ: TextType, output: true},
		&FixedParameter{name: "stderr", typ: TextType, output: true},
	},
}

var Types map[string]Type = map[string]Type{
	"format": FormatType,
	"text":   TextType,
	"exec":   ExecType,
}

var TheVM *VM = &VM{Blueprints: make(map[*Blueprint]bool)}
