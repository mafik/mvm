package mvm

import (
	"bytes"
	"fmt"
	"os/exec"
)

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
		{Name: "output", Typ: &TextType},
		{Name: "fmt", Typ: &TextType},
		{Name: "args"},
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
		{Name: "command", Typ: &TextType},
		{Name: "args", Typ: &TextType},
		{Name: "stdout", Typ: &TextType, Output: true},
		{Name: "stderr", Typ: &TextType, Output: true},
	},
}

var Types map[string]Type = map[string]Type{
	"format": FormatType,
	"text":   TextType,
	"exec":   ExecType,
}

var TheVM *VM = &VM{Blueprints: make(map[*Blueprint]bool)}
