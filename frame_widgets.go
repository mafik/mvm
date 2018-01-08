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
func (f *Frame) PayloadRight(obj *Object, m ui.TextMeasurer) float64 {
	return f.PayloadLeft(m) + f.PayloadWidth(obj, m)
}
func (f *Frame) PayloadWidth(obj *Object, m ui.TextMeasurer) float64 {
	return ButtonSize(m.MeasureText(obj.typ.Name()))
}
func (f *Frame) PayloadHeight() float64 { return f.TitleHeight() }

func (f *Frame) ParamCenter(i int) vec2.Vec2 {
	return vec2.Vec2{
		X: f.ContentLeft() + param_r,
		Y: f.ContentBottom() + float64(i)*(param_r*2+margin) + margin + param_r,
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
	Frame           *Frame
	Object          *Object
	BlueprintObject *Object
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
		Schedule{f.Frame, f.Object},
		DeleteFrame{f.Frame},
		CopyFrame{f.Frame, f.Object},
		CloneFrame{f.Frame, f.Object},
		ToggleParameter{f.Frame},
		TogglePublic{f.Frame},
		AddParameter{f.Frame},
		Raise{f.Frame},
		Lower{f.Frame, f.BlueprintObject},
	}
	if f.Object != nil {
		options = append(options, ClearFrame{f.Frame, f.Object})
	} else {
		options = append(options, NewBlueprint{f.Frame, f.BlueprintObject})
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

type FrameElementCircle struct {
	Frame           *Frame
	Index           int
	Param           Parameter
	IsMember        bool
	Name            string
	blueprintObject *Object
}

func (p FrameElementCircle) Draw(ctx *ui.Context2D) {
	ctx.BeginPath()
	ctx.Circle(vec2.Vec2{0, 0}, param_r)
	if p.Param != nil || p.IsMember {
		ctx.FillStyle("#fff")
		ctx.Fill()
	}
	if p.Index < len(p.Frame.elems) {
		ctx.LineWidth(2)
		ctx.StrokeStyle("#000")
		ctx.Stroke()
	}
	if p.IsMember {
		ctx.BeginPath()
		ctx.Circle(vec2.Vec2{0, 0}, param_r-4)
		ctx.FillStyle("#000")
		ctx.Fill()
	}
}
func (p FrameElementCircle) Options(pos vec2.Vec2) []ui.Option {
	if !p.IsMember {
		return []ui.Option{ParameterDragging{p.Frame, p.blueprintObject, p.Name}}
	}
	return nil
}
func (p FrameElementCircle) Size(ui.TextMeasurer) ui.Box {
	return ui.Box{-param_r, param_r, param_r, -param_r}
}

type FrameElementWidget struct {
	Frame           *Frame
	Index           int
	Param           Parameter
	IsMember        bool
	Name            string
	blueprintObject *Object
}

func (p FrameElementWidget) Children() []interface{} {
	return []interface{}{FrameElementCircle{p.Frame, p.Index, p.Param, p.IsMember, p.Name, p.blueprintObject}}
}
func (p FrameElementWidget) Draw(ctx *ui.Context2D) {
	ctx.FillStyle("#000")
	ctx.FillText(p.Name, param_r+margin, -3)
}
func (p FrameElementWidget) Options(vec2.Vec2) (opts []ui.Option) {
	if p.Index < len(p.Frame.elems) {
		opts = append(opts, DeleteParameter{p.Frame, p.Index})
	}
	return
}
func (p FrameElementWidget) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(vec2.Vec2{0, ParamOffset(p.Index)})
}
func (p FrameElementWidget) Size(measurer ui.TextMeasurer) ui.Box {
	return ui.Box{-param_r, param_r + margin + measurer.MeasureText(p.Name), param_r, -param_r}.Grow(margin / 2)
}
func (p FrameElementWidget) GetText() string { return p.Name }
func (p FrameElementWidget) SetText(newName string) {
	p.Frame.GetElement(p.Name).Name = newName
}

type FrameElementList struct {
	Frame           *Frame
	Object          *Object
	blueprintObject *Object
}

func (p FrameElementList) Draw(ctx *ui.Context2D) {
	/*
		params := p.Frame.Parameters(p.Object)
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
	return matrix.Translate(vec2.Vec2{param_r, p.Frame.size.Y})
}
func (p FrameElementList) Children() (children []interface{}) {
	var objectParams []Parameter
	if p.Object != nil {
		objectParams = p.Object.typ.Parameters()
	}
	var objectMembers []string
	if p.Object != nil {
		objectMembers = p.Object.typ.Members()
	}
	for i, _ := range p.Frame.elems {
		name := p.Frame.elems[i].Name
		_, param := GetParam(objectParams, name)
		isMember := false
		for _, member := range objectMembers {
			if member == name {
				isMember = true
				break
			}
		}
		children = append(children, FrameElementWidget{p.Frame, len(children), param, isMember, name, p.blueprintObject})
	}

	if p.Object != nil {
		for _, param := range objectParams {
			name := param.Name()
			existing := p.Frame.FindElement(name)
			if existing != nil {
				continue
			}
			children = append(children, FrameElementWidget{p.Frame, len(children), param, false, name, p.blueprintObject})
		}
		for _, member := range objectMembers {
			existing := p.Frame.FindElement(member)
			if existing != nil {
				continue
			}
			children = append(children, FrameElementWidget{p.Frame, len(children), nil, true, member, p.blueprintObject})

		}
	}

	return children
}

// Frame Payload

type FramePayload struct {
	frame  *Frame
	object *Object
}

func (fp FramePayload) Draw(ctx *ui.Context2D) {
	box := fp.Size(ctx)
	ctx.FillStyle("#fff")
	ctx.BeginPath()
	ctx.Rect2(box)
	ctx.Fill()
	ctx.FillStyle(fp.frame.blueprint.DarkColor())
	ctx.FillText(fp.object.typ.Name(), box.Left+margin, box.Bottom-textMargin)
}
func (fp FramePayload) Size(m ui.TextMeasurer) ui.Box {
	return ui.Box{fp.frame.PayloadTop(),
		fp.frame.PayloadRight(fp.object, m),
		fp.frame.PayloadBottom(),
		fp.frame.PayloadLeft(m)}
}
func (fp FramePayload) Options(pos vec2.Vec2) []ui.Option {
	return []ui.Option{FrameDragging{fp.frame, vec2.Vec2{0, 0}}, Enter{fp.object}}
}

// FrameTile

type FrameTile struct {
	Frame  *Frame
	Object *Object
}

func (ft FrameTile) Children() []interface{} {
	o := ft.Object
	if o != nil {
		w := o.typ.MakeWidget(o)
		if w != nil {
			return []interface{}{w}
		}
	}
	return nil
}
func (ft FrameTile) Size(m ui.TextMeasurer) ui.Box {
	return ft.Frame.ContentSize()
}
func (ft FrameTile) Draw(ctx *ui.Context2D) {
	ctx.BeginPath()
	ctx.Rect2(ft.Size(ctx).Grow(-2))
	ctx.Save()
	ctx.Clip()
}
func (ft FrameTile) PostDraw(ctx *ui.Context2D) {
	ctx.Restore()
	ctx.BeginPath()
	ctx.Rect2(ft.Size(ctx).Grow(-1))
	ctx.LineWidth(2)
	ctx.StrokeStyle(ft.Frame.blueprint.DarkColor())
	ctx.Stroke()
}
func (ft FrameTile) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(vec2.Scale(ft.Frame.size, 0.5))
}
func (ft FrameTile) MyFrame() *Frame { return ft.Frame }

// Frame Scaffolding

type FramedObject struct {
	Frame           *Frame
	Object          *Object
	BlueprintObject *Object
}

func (fobj FramedObject) Draw(ctx *ui.Context2D) {
	obj := fobj.Object
	size := fobj.Frame.size

	// Indicators
	if obj != nil && obj.execute {
		ctx.FillStyle("#f00")
		ctx.BeginPath()
		ctx.Rect2(ui.Box{-5, -5, size.X + 10, size.Y + 10})
		ctx.Fill()
	}
	if obj != nil && obj.running {
		ctx.Hourglass("#f00")
	}
}

func (fobj FramedObject) Options(p vec2.Vec2) []ui.Option {
	size := fobj.Frame.size
	box := ui.Box{0, size.X, size.Y, 0}
	if box.Contains(p) {
		return []ui.Option{StartFrameDragging(p, fobj.Frame), DeleteFrame{fobj.Frame}}
	}
	return nil
}

func (fobj FramedObject) Children() []interface{} {
	widgets := make([]interface{}, 0, 5)
	if fobj.Object != nil {
		widgets = append(widgets, FramePayload{fobj.Frame, fobj.Object})
	}
	widgets = append(widgets, FrameTile{fobj.Frame, fobj.Object})
	widgets = append(widgets, FrameTitle{fobj.Frame, fobj.Object, fobj.BlueprintObject})
	widgets = append(widgets, FrameElementList{fobj.Frame, fobj.Object, fobj.BlueprintObject})
	return widgets
}

func (fobj FramedObject) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(fobj.Frame.pos)
}
