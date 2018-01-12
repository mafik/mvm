package mvm

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mafik/mvm/ui"
	"github.com/mafik/mvm/vec2"
)

type FixedParameter struct {
	name string
}

func (p *FixedParameter) Name() string {
	return p.name
}

type PrimitiveType struct {
	name        string
	parameters  []Parameter
	instantiate func(*Object)
	copy        func(from, to *Object)
	run         func(Args)
	string      func(interface{}) string
	makeWidget  func(*Object) ui.Widget
}

func (t *PrimitiveType) Name() string {
	return t.name
}

func (t *PrimitiveType) Parameters() []Parameter {
	return t.parameters
}

func (*PrimitiveType) Members() []Member                 { return nil }
func (*PrimitiveType) GetMember(*Object, string) *Object { return nil }

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

func (t *PrimitiveType) MakeWidget(o *Object) ui.Widget {
	if t.makeWidget != nil {
		return t.makeWidget(o)
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
	makeWidget: func(o *Object) ui.Widget {
		return TextWidget{o}
	},
}

type TextWidget struct {
	o *Object
}

func (w TextWidget) Options(vec2.Vec2) []ui.Option { return nil }
func (w TextWidget) Size(ui.TextMeasurer) ui.Box   { return w.o.frame.ContentSize().Grow(-2) }
func (w TextWidget) Draw(ctx *ui.Context2D) {
	text := w.o.typ.String(w.o.priv)
	lines := strings.Split(text, "\n")
	ctx.FillStyle("#000")
	ctx.TextAlign("center")
	h := lineHeight * float64(len(lines))
	for i, line := range lines {
		ctx.FillText(line, 0, float64(i+1)*lineHeight-h/2-5)
	}
	width := ctx.MeasureText(lines[len(lines)-1])
	ctx.FillRect(width/2, h/2, 2, -lineHeight)
}
func (w TextWidget) GetText() string  { return string(w.o.priv.([]byte)) }
func (w TextWidget) SetText(s string) { w.o.priv = []byte(s) }

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
		&FixedParameter{name: "output"},
		&FixedParameter{name: "fmt"},
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
		&FixedParameter{name: "command"},
		&FixedParameter{name: "args"},
		&FixedParameter{name: "stdout"},
		&FixedParameter{name: "stderr"},
	},
}

var Types map[string]Type = map[string]Type{
	"format": FormatType,
	"text":   TextType,
	"exec":   ExecType,
}

var TheVM *VM = &VM{}
