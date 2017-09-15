package mvm

import (
	"bytes"
	"fmt"
	"os/exec"
)

var TextType Type = Type{
	Name: "text",
	Instantiate: func(me *Object) {
		var b bytes.Buffer
		fmt.Fprint(&b, "Hello world!")
		me.priv = b.Bytes()
	},
	String: func(i interface{}) string {
		return string(i.([]byte))
	},
}

var EchoType Type = Type{
	Name: "echo",
	Run: func(args Args) {
		fmt.Printf("Echo: \"%s\"\n", args["text"][0].priv)
	},
	Parameters: []Parameter{
		{Name: "text", Typ: &TextType},
		//{Name: "then", Runnable: true},
	},
}

var FormatType Type = Type{
	Name: "format",
	Run: func(args Args) {
		format := string(args["fmt"][0].priv.([]byte))
		var fmt_args []interface{}
		for _, o := range args["args"] {
			fmt_args = append(fmt_args, o.priv)
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, format, fmt_args...)
		args["output"][0].priv = buf.Bytes()
	},
	Parameters: []Parameter{
		{Name: "output", Typ: &TextType},
		{Name: "fmt", Typ: &TextType},
		{Name: "args"},
	},
}

var CommandType Type = Type{
	Name: "command",
	Run: func(args Args) {
		name := TextType.String(args["name"][0].priv)
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
	Parameters: []Parameter{
		{Name: "name", Typ: &TextType},
		{Name: "args", Typ: &TextType},
		{Name: "stdout", Typ: &TextType, Output: true},
		{Name: "stderr", Typ: &TextType, Output: true},
	},
}

var Types map[string]*Type = map[string]*Type{
	"format":  &FormatType,
	"text":    &TextType,
	"echo":    &EchoType,
	"command": &CommandType,
}

var TheVM *VM = &VM{Blueprints: make(map[*Blueprint]bool)}
