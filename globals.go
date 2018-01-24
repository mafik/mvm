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

type PrimitiveObject struct {
	name        string
	parameters  []Parameter
	instantiate func(*Shell)
	copy        func(from, to *Shell)
	run         func(Args)
	string      func(interface{}) string
	makeWidget  func(*Shell) ui.Widget
}

func (t *PrimitiveObject) Name() string {
	return t.name
}

func (t *PrimitiveObject) Parameters() []Parameter {
	return t.parameters
}

func (*PrimitiveObject) Members() []Member               { return nil }
func (*PrimitiveObject) GetMember(*Shell, string) *Shell { return nil }

func (t *PrimitiveObject) Instantiate(s *Shell) {
	if t.instantiate != nil {
		t.instantiate(s)
	}
}

func (t *PrimitiveObject) Copy(from, to *Shell) {
	if t.copy != nil {
		t.copy(from, to)
	}
}

func (t *PrimitiveObject) Run(args Args) {
	t.run(args)
}

func (t *PrimitiveObject) String(i interface{}) string {
	if t.string != nil {
		return t.string(i)
	} else {
		return fmt.Sprintf("%#v", i)
	}
}

func (t *PrimitiveObject) MakeWidget(s *Shell) ui.Widget {
	if t.makeWidget != nil {
		return t.makeWidget(s)
	}
	return nil
}

var TextObject Object = &PrimitiveObject{
	name: "text",
	instantiate: func(me *Shell) {
		me.priv = []byte{}
	},
	copy: func(from, to *Shell) {
		to.priv = append([]byte{}, from.priv.([]byte)...)
	},
	string: func(i interface{}) string {
		return string(i.([]byte))
	},
	makeWidget: func(s *Shell) ui.Widget {
		return TextWidget{s}
	},
}

type TextWidget struct {
	s *Shell
}

func (w TextWidget) Options(vec2.Vec2) []ui.Option { return nil }
func (w TextWidget) Size(ui.TextMeasurer) ui.Box   { return w.s.frame.ContentSize().Grow(-2) }
func (w TextWidget) Draw(ctx *ui.Context2D) {
	ctx.BeginPath()
	ctx.Rect2(w.s.frame.ContentSize())
	ctx.FillStyle("#fff")
	ctx.Fill()
	text := w.s.object.String(w.s.priv)
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
func (w TextWidget) GetText() string  { return string(w.s.priv.([]byte)) }
func (w TextWidget) SetText(s string) { w.s.priv = []byte(s) }

var CopyObject Object = &PrimitiveObject{
	name: "copy",
	run: func(args Args) {
		from := args.Get("from")
		to := &Shell{nil, nil, false, false, from.object, nil}
		from.object.Copy(from, to)
		args.Set("to", to)
	},
	parameters: []Parameter{
		&FixedParameter{name: "from"},
		&FixedParameter{name: "to"},
	},
}

var FormatObject Object = &PrimitiveObject{
	name: "format",
	run: func(args Args) {
		format := string(args.Get("fmt").priv.([]byte))
		fmt_args := []interface{}{args.Get("args").priv}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, format, fmt_args...)
		args.Get("output").priv = buf.Bytes()
	},
	parameters: []Parameter{
		&FixedParameter{name: "output"},
		&FixedParameter{name: "fmt"},
		&FixedParameter{name: "args"},
	},
}

var ExecObject Object = &PrimitiveObject{
	name: "exec",
	run: func(args Args) {
		name := TextObject.String(args.Get("command").priv)
		cmd_args := []string{}
		if args.Get("args") != nil {
			cmd_args = append(cmd_args, TextObject.String(args.Get("args").priv))
		}
		out, err := exec.Command(name, cmd_args...).Output()
		if err != nil {
			if args.Get("stderr") != nil {
				switch err := err.(type) {
				case *exec.ExitError:
					args.Get("stderr").priv = err.Stderr
				case *exec.Error:
					args.Get("stderr").priv = []byte(err.Error())
				}
			}
			return
		}
		args.Get("stdout").priv = out
	},
	parameters: []Parameter{
		&FixedParameter{name: "command"},
		&FixedParameter{name: "args"},
		&FixedParameter{name: "stdout"},
		&FixedParameter{name: "stderr"},
	},
}

var Objects map[string]Object = map[string]Object{
	"format": FormatObject,
	"text":   TextObject,
	"exec":   ExecObject,
	"copy":   CopyObject,
}

var TheVM *VM = &VM{}
