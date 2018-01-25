package mvm

import (
	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/ui"
	"github.com/mafik/mvm/vec2"
)

type FramePart interface {
	MyFrame() *Frame
}

func (f *Frame) ContentSize() ui.Box {
	return ui.Box{-f.size.Y / 2, f.size.X / 2, f.size.Y / 2, -f.size.X / 2}
}

func (f *Frame) ContentLeft() float64   { return 0 }
func (f *Frame) ContentTop() float64    { return 0 }
func (f *Frame) ContentBottom() float64 { return f.ContentHeight() }
func (f *Frame) ContentRight() float64  { return f.ContentWidth() }
func (f *Frame) ContentWidth() float64  { return f.size.X }
func (f *Frame) ContentHeight() float64 { return f.size.Y }

func (f *Frame) TitleLeft() float64   { return f.ContentLeft() }
func (f *Frame) TitleTop() float64    { return f.TitleBottom() - buttonHeight }
func (f *Frame) TitleBottom() float64 { return f.ContentTop() }
func (f *Frame) TitleRight(m ui.TextMeasurer) float64 {
	return f.TitleLeft() + f.TitleWidth(m)
}
func (f *Frame) TitleWidth(m ui.TextMeasurer) float64 {
	return ButtonSize(m.MeasureText(f.Title()))
}
func (f *Frame) TitleHeight() float64 { return buttonHeight }

func (f *Frame) PayloadLeft(m ui.TextMeasurer) float64 { return f.TitleRight(m) }
func (f *Frame) PayloadTop() float64                   { return f.TitleTop() }
func (f *Frame) PayloadBottom() float64                { return f.TitleBottom() }
func (f *Frame) PayloadRight(s *Shell, m ui.TextMeasurer) float64 {
	return f.PayloadLeft(m) + f.PayloadWidth(s, m)
}
func (f *Frame) PayloadWidth(s *Shell, m ui.TextMeasurer) float64 {
	return ButtonSize(m.MeasureText(s.object.Name()))
}
func (f *Frame) PayloadHeight() float64 { return f.TitleHeight() }

func (f *Frame) ParamCenter(i int) vec2.Vec2 {
	y := 0.0
	if f.ShowWindow {
		y = f.size.Y
	}
	return vec2.Vec2{
		X: f.ContentLeft() + param_r,
		Y: y + float64(i)*(param_r*2+margin) + margin + param_r,
	}
}

func CircleClicked(pos vec2.Vec2, touch vec2.Vec2) bool {
	return Dist(pos, touch) < param_r
}

func ParamOffset(i int) float64 {
	return param_r + margin + float64(i)*(param_r*2+margin)
}

// Frame Title

type FrameTitle struct {
	Frame          *Frame
	Shell          *Shell
	BlueprintShell *Shell
}

func (f FrameTitle) Draw(ctx *ui.Context2D) {
	box := f.Size(ctx)
	ctx.FillStyle(f.Frame.blueprint.DarkColor())
	ctx.BeginPath()
	ctx.Rect2(box)
	ctx.Fill()
	ctx.FillStyle("#fff")
	ctx.FillText(f.Frame.Title(), box.Left+margin, box.Bottom-textMargin)
}
func (f FrameTitle) Options(pos vec2.Vec2) []ui.Option {
	options := []ui.Option{
		FrameDragging{f.Frame, vec2.Vec2{0, 0}},
		Schedule{f.Frame, f.Shell},
		DeleteFrame{f.Frame},
		CopyFrame{f.Frame, f.Shell},
		CloneFrame{f.Frame, f.Shell},
		ToggleParameter{f.Frame},
		TogglePublic{f.Frame},
		ToggleShowWindow{f.Frame},
		AddParameter{f.Frame},
		Raise{f.Frame},
		Lower{f.Frame, f.BlueprintShell},
	}
	if f.Shell != nil {
		options = append(options, ClearFrame{f.Frame, f.Shell})
	} else {
		options = append(options, NewBlueprint{f.Frame, f.BlueprintShell})
	}
	return options
}
func (f FrameTitle) Size(m ui.TextMeasurer) ui.Box {
	frame := f.Frame
	return ui.Box{frame.TitleTop(), frame.TitleRight(m), frame.TitleBottom(), frame.TitleLeft()}
}
func (f FrameTitle) GetText() string  { return f.Frame.name }
func (f FrameTitle) SetText(s string) { f.Frame.name = s }
func (f FrameTitle) MyFrame() *Frame  { return f.Frame }

type FrameElementPointer struct {
	Frame   *Frame
	Index   int
	Machine *Shell
}

func (self FrameElementPointer) Zip() ElementPack {
	machine := self.Machine.object.(*Machine)
	shell := machine.shells[self.Frame]
	return self.Frame.ZipElements(shell)[self.Index]
}
func (self FrameElementPointer) Member() Member {
	return self.Zip().Member
}
func (self FrameElementPointer) Param() Member {
	return self.Zip().Param
}
func (self FrameElementPointer) PositionInFrame() vec2.Vec2 {
	return vec2.Vec2{0, ParamOffset(self.Index)}
}
func (self FrameElementPointer) IsMember() bool {
	return self.Member() != nil
}
func (self FrameElementPointer) IsDefined() bool {
	return self.Member() != nil || self.Param() != nil
}
func (self FrameElementPointer) FrameElement() *FrameElement {
	if self.Index < len(self.Frame.elems) {
		return self.Frame.elems[self.Index]
	}
	return nil
}
func (self FrameElementPointer) MakeFrameElement() *FrameElement {
	if el := self.FrameElement(); el != nil {
		return el
	}
	return self.Frame.GetElement(self.Name())
}
func (self FrameElementPointer) Name() string {
	pack := self.Zip()
	if pack.FrameElement != nil {
		return pack.FrameElement.Name
	}
	if pack.Member != nil {
		return pack.Member.Name()
	}
	return pack.Param.Name()
}

type FrameElementCircle struct {
	FrameElementPointer
}

func (p FrameElementCircle) Draw(ctx *ui.Context2D) {
	ctx.BeginPath()
	if p.IsMember() {
		ctx.Rect2(ui.Box{-param_r, param_r, param_r, -param_r}.Grow(-1))
	} else {
		ctx.Circle(vec2.Vec2{0, 0}, param_r)
	}
	if p.IsDefined() {
		ctx.FillStyle("#fff")
		ctx.Fill()
	}
	if p.FrameElement() != nil {
		ctx.LineWidth(2)
		ctx.StrokeStyle("#000")
		ctx.Stroke()
	}
}
func (p FrameElementCircle) Options(pos vec2.Vec2) []ui.Option {
	return []ui.Option{ParameterDragging{p.FrameElementPointer}}
}
func (p FrameElementCircle) Size(ui.TextMeasurer) ui.Box {
	return ui.Box{-param_r, param_r, param_r, -param_r}
}

type FrameElementWidget struct {
	FrameElementPointer
}

func (p FrameElementWidget) Children() []interface{} {
	return []interface{}{FrameElementCircle{p.FrameElementPointer}}
}
func (p FrameElementWidget) Draw(ctx *ui.Context2D) {
	ctx.FillStyle("#000")
	ctx.FillText(p.GetText(), param_r+margin, -3)
}
func (p FrameElementWidget) Options(vec2.Vec2) (opts []ui.Option) {
	if el := p.FrameElement(); el != nil {
		opts = append(opts, DeleteParameter{el})
	}
	return
}
func (p FrameElementWidget) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(p.PositionInFrame())
}
func (p FrameElementWidget) Size(measurer ui.TextMeasurer) ui.Box {
	return ui.Box{-param_r, param_r + margin + measurer.MeasureText(p.GetText()), param_r, -param_r}.Grow(margin / 2)
}
func (p FrameElementWidget) GetText() string {
	return p.Name()
}
func (p FrameElementWidget) SetText(newName string) {
	p.MakeFrameElement().Name = newName
}

type FrameElementList struct {
	Frame *Frame
	Shell *Shell
}

func (p FrameElementList) Draw(ctx *ui.Context2D) {
	/*
		params := p.Frame.Parameters(p.Shell)
		n := len(params)
		if n > 0 {
			ctx.LineWidth(2)
			ctx.BeginPath()
			ctx.MoveTo2(vec2.Vec2{0, 0})
			ctx.LineTo2(vec2.Vec2{0, ParamOffset(n - 1)})
			ctx.Stroke()
		}
	*/
}
func (p FrameElementList) Options(vec2.Vec2) []ui.Option { return nil }
func (p FrameElementList) Transform(m ui.TextMeasurer) matrix.Matrix {
	y := 0.
	if p.Frame.ShowWindow {
		y = p.Frame.size.Y
	}
	return matrix.Translate(vec2.Vec2{param_r, y})
}
func (p FrameElementList) Children() (children []interface{}) {
	zip := p.Frame.ZipElements(p.Shell)
	for i, _ := range zip {
		children = append(children, FrameElementWidget{FrameElementPointer{p.Frame, i, p.Shell.parent}})
	}
	return children
}

// Frame Payload

type FramePayload struct {
	frame *Frame
	shell *Shell
}

func (fp FramePayload) Draw(ctx *ui.Context2D) {
	box := fp.Size(ctx)
	ctx.FillStyle("#fff")
	ctx.BeginPath()
	ctx.Rect2(box)
	ctx.Fill()
	ctx.FillStyle(fp.frame.blueprint.DarkColor())
	ctx.FillText(fp.shell.object.Name(), box.Left+margin, box.Bottom-textMargin)
}
func (fp FramePayload) Size(m ui.TextMeasurer) ui.Box {
	return ui.Box{fp.frame.PayloadTop(),
		fp.frame.PayloadRight(fp.shell, m),
		fp.frame.PayloadBottom(),
		fp.frame.PayloadLeft(m)}
}
func (fp FramePayload) Options(pos vec2.Vec2) []ui.Option {
	return []ui.Option{FrameDragging{fp.frame, vec2.Vec2{0, 0}}, Enter{fp.shell}}
}

type FrameBlueprintPayload struct {
	FramePayload
	Blueprint *Blueprint
}

func (fbp FrameBlueprintPayload) GetText() string     { return fbp.Blueprint.name }
func (fbp FrameBlueprintPayload) SetText(text string) { fbp.Blueprint.name = text }

// FrameWindow

type FrameWindow struct {
	Frame *Frame
	Shell *Shell
}

func (ft FrameWindow) Children() []interface{} {
	s := ft.Shell
	if s == nil {
		return nil
	}
	if graphic, ok := s.object.(GraphicObject); ok {
		w := graphic.MakeWidget(s)
		if w != nil {
			return []interface{}{w}
		}
	}
	return nil
}
func (ft FrameWindow) Size(m ui.TextMeasurer) ui.Box {
	return ft.Frame.ContentSize()
}
func (ft FrameWindow) Draw(ctx *ui.Context2D) {
	ctx.BeginPath()
	ctx.Rect2(ft.Size(ctx).Grow(-2))
	ctx.Save()
	ctx.Clip()
}
func (ft FrameWindow) PostDraw(ctx *ui.Context2D) {
	ctx.Restore()
	ctx.BeginPath()
	ctx.Rect2(ft.Size(ctx).Grow(-1))
	ctx.LineWidth(2)
	ctx.StrokeStyle(ft.Frame.blueprint.DarkColor())
	ctx.Stroke()
}
func (ft FrameWindow) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(vec2.Scale(ft.Frame.size, 0.5))
}
func (ft FrameWindow) MyFrame() *Frame { return ft.Frame }

// Frame Scaffolding

type FrameWidget struct {
	Frame          *Frame
	Shell          *Shell
	BlueprintShell *Shell
}

func (w FrameWidget) Draw(ctx *ui.Context2D) {
	shell := w.Shell
	f := w.Frame

	// Indicators
	if shell != nil && shell.execute {
		ctx.FillStyle("#f00")
		ctx.BeginPath()
		ctx.Rect2(ui.Box{f.TitleTop() - 5, f.PayloadRight(shell, ctx) + 5, f.TitleBottom() + 5, f.TitleLeft() - 5})
		ctx.Fill()
	}
	if shell != nil && shell.running {
		ctx.Save()
		ctx.Translate(-4, -5)
		ctx.Hourglass("#f00")
		ctx.Restore()
	}
}

func (w FrameWidget) Options(p vec2.Vec2) []ui.Option {
	if w.Frame.Hidden {
		return nil
	}
	if !w.Frame.ShowWindow {
		return nil
	}
	size := w.Frame.size
	box := ui.Box{0, size.X, size.Y, 0}
	if box.Contains(p) {
		return []ui.Option{StartFrameDragging(p, w.Frame)}
	}
	return nil
}

func (w FrameWidget) Children() []interface{} {
	widgets := make([]interface{}, 0, 5)
	if w.Shell != nil {
		blueprint, ok := w.Shell.object.(*Blueprint)
		if ok {
			widgets = append(widgets, FrameBlueprintPayload{
				FramePayload: FramePayload{w.Frame, w.Shell},
				Blueprint:    blueprint,
			})
		} else {
			widgets = append(widgets, FramePayload{w.Frame, w.Shell})
		}
	}
	if w.Frame.ShowWindow {
		widgets = append(widgets, FrameWindow{w.Frame, w.Shell})
	}
	widgets = append(widgets, FrameTitle{w.Frame, w.Shell, w.BlueprintShell})
	widgets = append(widgets, FrameElementList{w.Frame, w.Shell})
	return widgets
}

func (w FrameWidget) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(w.Frame.pos)
}
