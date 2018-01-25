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

type Text struct {
	Bytes []byte
}

func (*Text) Name() string { return "text" }
func (text *Text) Copy(shell *Shell) {
	shell.object = &Text{append([]byte{}, text.Bytes...)}
}
func (*Text) MakeWidget(shell *Shell) ui.Widget {
	return TextWidget{shell}
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
	text := w.GetText()
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
func (w TextWidget) GetText() string  { return string(w.s.object.(*Text).Bytes) }
func (w TextWidget) SetText(s string) { w.s.object.(*Text).Bytes = []byte(s) }

type CopyType struct{}

var CopyParameters []Parameter = []Parameter{
	&FixedParameter{name: "from"},
	&FixedParameter{name: "to"},
}

func (CopyType) Name() string            { return "copy" }
func (CopyType) Parameters() []Parameter { return CopyParameters }
func (CopyType) Run(args Args) {
	from := args.Get("from")
	copy := Copy(from.object, nil, nil)
	args.Set("to", copy)
}

type FormatType struct{}

var FormatParameters []Parameter = []Parameter{
	&FixedParameter{name: "output"},
	&FixedParameter{name: "fmt"},
	&FixedParameter{name: "args"},
}

func (FormatType) Name() string            { return "format" }
func (FormatType) Parameters() []Parameter { return FormatParameters }
func (FormatType) Run(args Args) {
	format := string(args.Get("fmt").object.(*Text).Bytes)
	fmt_args := []interface{}{args.Get("args").object}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, format, fmt_args...)
	s := MakeShell(nil, nil)
	s.object = &Text{buf.Bytes()}
	args.Set("output", s)
}

type ExecType struct{}

var ExecParameters []Parameter = []Parameter{
	&FixedParameter{name: "command"},
	&FixedParameter{name: "args"},
	&FixedParameter{name: "stdout"},
	&FixedParameter{name: "stderr"},
}

func (ExecType) Name() string            { return "exec" }
func (ExecType) Parameters() []Parameter { return ExecParameters }
func (ExecType) Run(args Args) {
	name := string(args.Get("command").object.(*Text).Bytes)
	cmd_args := []string{}
	if args.Get("args") != nil {
		cmd_args = append(cmd_args, string(args.Get("args").object.(*Text).Bytes))
	}
	out, err := exec.Command(name, cmd_args...).Output()
	if err != nil {
		if args.Get("stderr") != nil {
			switch err := err.(type) {
			case *exec.ExitError:
				args.Get("stderr").object.(*Text).Bytes = err.Stderr
			case *exec.Error:
				args.Get("stderr").object.(*Text).Bytes = []byte(err.Error())
			}
		}
		return
	}
	args.Get("stdout").object.(*Text).Bytes = out
}

var Objects map[string]Object = map[string]Object{
	"format": FormatType{},
	"text":   &Text{},
	"exec":   ExecType{},
	"copy":   CopyType{},
}

var TheVM *VM = &VM{}
