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

type CType struct{ value int }

func (ctype CType) Name() string {
	switch ctype.value {
	case 0:
		return "ctype:void"
	case 1:
		return "ctype:uint8"
	case 2:
		return "ctype:uint16"
	case 3:
		return "ctype:uint32"
	case 4:
		return "ctype:uint64"
	case 5:
		return "ctype:int8"
	case 6:
		return "ctype:int16"
	case 7:
		return "ctype:int32"
	case 8:
		return "ctype:int64"
	case 9:
		return "ctype:float"
	case 10:
		return "ctype:double"
	case 11:
		return "ctype:pointer"
	default:
		return "ctype:???"
	}
}

type CTypes []CType

var CTypesArray CTypes = CTypes{
	CType{value: 0},
	CType{value: 1},
	CType{value: 2},
	CType{value: 3},
	CType{value: 4},
	CType{value: 5},
	CType{value: 6},
	CType{value: 7},
	CType{value: 8},
	CType{value: 9},
	CType{value: 10},
	CType{value: 11},
}

func (CTypes) Name() string { return "C types" }
func (self CTypes) Members() (m []Member) {
	for _, t := range self {
		m = append(m, t)
	}
	return m
}
func (self CTypes) GetMember(name string) *Shell {
	for _, t := range self {
		if t.Name() == name {
			s := MakeShell(nil, nil)
			s.object = t
			return s
		}
	}
	return nil
}

type CTypesGob struct{}

func (CTypesGob) Ungob() Gobbable        { return CTypesArray }
func (CTypes) Gob(Serializer) Gob        { return CTypesGob{} }
func (CTypes) Connect(Deserializer, Gob) {}

type Ptr uintptr
type PtrWidget struct{ *Shell }

type Wrapper interface {
	Unwrap() interface{}
}

func (Ptr) Name() string                        { return "pointer" }
func (Ptr) MakeWidget(s *Shell) ui.Widget       { return PtrWidget{s} }
func (self Ptr) Unwrap() interface{}            { return uintptr(self) }
func (PtrWidget) Options(vec2.Vec2) []ui.Option { return nil }
func (w PtrWidget) Draw(ctx *ui.Context2D) {
	s := fmt.Sprintf("0x%x", w.object.(Ptr))
	ctx.TextAlign("center")
	ctx.FillStyle("#000")
	ctx.FillText(s, 0, 0)
}

type CString struct{}

var CStringParameters []Parameter = []Parameter{
	&FixedParameter{"s"},
	&FixedParameter{"result"},
}

func (CString) Name() string            { return "CString" }
func (CString) Parameters() []Parameter { return CStringParameters }
func (CString) Run(args Args) {
	//s := string(args.Get("s").object.(*Text).Bytes)
	var ptr uintptr = 0 // fcall.CString(s)
	shell := MakeShell(nil, nil)
	shell.object = Ptr(ptr)
	args.Set("result", shell)
}

type Function struct {
	//f      fcall.Function
	name   string
	rtype  CType
	atypes []CType
}

func (f *Function) Name() string { return f.name }
func (f *Function) Parameters() (params []Parameter) {
	for i, _ := range f.atypes {
		params = append(params, &FixedParameter{fmt.Sprint(i)})
	}
	params = append(params, &FixedParameter{"ret"})
	return
}
func (f *Function) Run(args Args) {
	var fargs []interface{}
	for i, _ := range f.atypes {
		farg := args.Get(fmt.Sprint(i)).object.(Wrapper).Unwrap()
		fargs = append(fargs, farg)
	}
	//ret := f.f(fargs).(Object)
	//shell := MakeShell(nil, nil)
	//shell.object = ret
	//args.Set("ret", shell)
}

type GetFunction struct{}

var GetFunctionParameters []Parameter = []Parameter{
	&FixedParameter{"name"},
	&FixedParameter{"rtype"},
	&FixedParameter{"atypes"},
	&FixedParameter{"result"},
}

func (GetFunction) Name() string            { return "GetFunction" }
func (GetFunction) Parameters() []Parameter { return GetFunctionParameters }
func (GetFunction) Run(args Args) {
	/*
	name := string(args.Get("name").object.(*Text).Bytes)
	rtype := args.Get("rtype").object.(CType)
	atype := args.Get("atypes").object.(CType)
	f, err := fcall.GetFunction(name, rtype.value, atype.value)
	if err != nil {
		panic(err)
	}
	shell := MakeShell(nil, nil)
	shell.object = &Function{f, name, rtype, []CType{atype}}
	args.Set("result", shell)
	*/
}

var Gobs []Gob = []Gob{
	CTypesGob{},
}

var Objects []Object = []Object{
	FormatType{},
	&Text{},
	ExecType{},
	CopyType{},
	Ptr(0),
	CTypesArray,
	CString{},
	GetFunction{},
	&Function{"fn", CType{0}, nil},
}

var TheVM *VM = &VM{}
