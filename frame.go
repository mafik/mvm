package mvm

import (
	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/ui"
	"github.com/mafik/mvm/vec2"
)

type Frame struct {
	blueprint *Blueprint
	pos       vec2.Vec2
	size      vec2.Vec2
	name      string
	params    []FrameParameter
	param     bool
	Hidden    bool
}

func (f *Frame) ContentSize() ui.Box {
	return ui.Box{-f.size.Y / 2, f.size.X / 2, f.size.Y / 2, -f.size.X / 2}
}

type FrameParameter struct {
	Name   string
	Target *Frame
	Stiff  bool
}

func (f *Frame) FindFrameParameter(name string) *FrameParameter {
	for i, frame_parameter := range f.params {
		if frame_parameter.Name == name {
			return &f.params[i]
		}
	}
	return nil
}

func (f *Frame) ForceFindFrameParameter(name string) *FrameParameter {
	links := f.FindFrameParameter(name)
	if links == nil {
		f.params = append(f.params, FrameParameter{name, nil, false})
		links = &f.params[len(f.params)-1]
	}
	return links
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

// Frame Title

type FrameTitle struct {
	Frame           *Frame
	Object          *Object
	BlueprintObject *Object
}

func (f FrameTitle) Draw(ctx *ui.Context2D) {
	box := f.Size(ctx)
	ctx.FillStyle("#000")
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

// Raise

type Raise struct {
	Frame *Frame
}

func (Raise) Name() string    { return "Raise" }
func (Raise) Keycode() string { return "KeyG" }
func (r Raise) Activate(ctx ui.TouchContext) ui.Action {
	var newB *Blueprint
	var newO *Object
	var newPos vec2.Vec2
	f := r.Frame
	oldB := f.blueprint
	ctx.AtTopBlueprint().Query(func(path ui.TreePath, pos vec2.Vec2) ui.WalkAction {
		last := path[len(path)-1]
		bw, ok := last.(BlueprintWidget)
		if !ok {
			return ui.Explore
		}
		if bw.Blueprint != oldB {
			newB = bw.Blueprint
			newO = bw.Object
			newPos = pos
			return ui.Explore
		}
		if newB == nil {
			return ui.Return
		}

		oldF := bw.Object.frame
		newB.frames = append(newB.frames, f)
		f.blueprint = newB
		f.pos = newPos

		X := 0
		for x, frame := range oldB.frames {
			if frame == f {
				X = x
			}
			// TODO: create parameter object and point it to the new frame
			for s, _ := range frame.params {
				if f == frame.params[s].Target {
					frame.params[s].Target = nil
				}
			}
		}
		oldB.frames = append(oldB.frames[:X], oldB.frames[X+1:]...)

		for newO, _ := range newB.instances {
			newM := newO.priv.(*Machine)
			oldO, ok := newM.objects[oldF]
			if !ok {
				continue
			}
			if oldO.typ != oldB {
				continue
			}
			oldM := oldO.priv.(*Machine)
			o, ok := oldM.objects[f]
			if !ok {
				continue
			}
			delete(oldM.objects, f)
			newM.objects[f] = o
		}

		return ui.Return
	})
	return nil
}

// Lower

type Lower struct {
	Frame           *Frame
	BlueprintObject *Object
}

func (Lower) Name() string    { return "Lower" }
func (Lower) Keycode() string { return "KeyB" }
func (l Lower) Activate(ctx ui.TouchContext) ui.Action {
	var top *BlueprintWidget
	mine := false
	ctx.AtTopBlueprint().Query(func(path ui.TreePath, pos vec2.Vec2) ui.WalkAction {
		last := path[len(path)-1]
		if bw, ok := last.(BlueprintWidget); ok {
			//fmt.Println("Detected blueprint widget!")
			if top == nil {
				top = &bw
			}
			if mine {
				f := l.Frame
				oldB := f.blueprint
				newB := bw.Blueprint
				newF := bw.Object.frame
				newB.frames = append(newB.frames, f)
				f.blueprint = newB
				f.pos = pos

				X := 0
				for x, frame := range oldB.frames {
					if frame == f {
						X = x
					}
					// TODO: make the frame public and create parameters to maintain connections
					for s, _ := range frame.params {
						if f == frame.params[s].Target {
							frame.params[s].Target = nil
						}
					}
				}
				oldB.frames = append(oldB.frames[:X], oldB.frames[X+1:]...)

				for oldO, _ := range oldB.instances {
					oldM := oldO.priv.(*Machine)
					newO, ok := oldM.objects[newF]
					if !ok {
						continue
					}
					if newO.typ != newB {
						continue
					}
					o, ok := oldM.objects[f]
					if !ok {
						continue
					}
					delete(oldM.objects, f)
					newM := newO.priv.(*Machine)
					newM.objects[f] = o

				}
				return ui.Return
			}
			if bw.Object == l.BlueprintObject {
				mine = true
				//fmt.Println("That's my blueprint!")
			}
		}
		return ui.Explore
	})
	return nil
}

// Schedule

type Schedule struct {
	Frame  *Frame
	Object *Object
}

func (s Schedule) Name() string    { return "Schedule" }
func (s Schedule) Keycode() string { return "Space" }
func (s Schedule) Activate(ui.TouchContext) ui.Action {
	s.Object.MarkForExecution()
	return nil
}

// Enter

type Enter struct {
	Object *Object
}

func (e Enter) Name() string    { return "Enter" }
func (e Enter) Keycode() string { return "KeyE" }
func (e Enter) Activate(ctx ui.TouchContext) ui.Action {
	clientUI := ctx.Path[0].(*ClientUI)
	clientUI.focus = e.Object
	return nil
}

// New Blueprint

type NewBlueprint struct {
	Frame           *Frame
	BlueprintObject *Object
}

func (nb NewBlueprint) Name() string    { return "New blueprint" }
func (nb NewBlueprint) Keycode() string { return "KeyZ" }
func (nb NewBlueprint) Activate(ctx ui.TouchContext) ui.Action {
	b := MakeBlueprint("New blueprint")
	nb.Frame.blueprint.FillWithNew(nb.Frame, b)
	return nil
}

// Clear Frame

type ClearFrame struct {
	Frame  *Frame
	Object *Object
}

func (cf ClearFrame) Name() string    { return "Clear frame" }
func (cf ClearFrame) Keycode() string { return "KeyZ" }
func (cf ClearFrame) Activate(ctx ui.TouchContext) ui.Action {
	m := cf.Object.parent.priv.(*Machine)
	delete(m.objects, cf.Frame)
	return nil
}

// Copy Frame

type CopyFrame struct {
	Frame  *Frame
	Object *Object
}

func (cf CopyFrame) Name() string    { return "Copy frame" }
func (cf CopyFrame) Keycode() string { return "KeyC" }
func (cf CopyFrame) Activate(ctx ui.TouchContext) ui.Action {
	f := cf.Frame.blueprint.AddFrame()
	f.pos = ctx.AtTopBlueprint().Position()
	f.size = cf.Frame.size
	if cf.Object != nil {
		cf.Frame.blueprint.FillWithNew(f, cf.Object.typ)
	}
	return FrameDragging{f, vec2.Vec2{0, 0}}
}

// Frame clone

type CloneFrame struct {
	Frame  *Frame
	Object *Object
}

func (cf CloneFrame) Name() string    { return "Clone frame" }
func (cf CloneFrame) Keycode() string { return "KeyV" }
func (cf CloneFrame) Activate(ctx ui.TouchContext) ui.Action {
	f := cf.Frame.blueprint.AddFrame()
	f.pos = ctx.AtTopBlueprint().Position()
	f.size = cf.Frame.size
	if cf.Object != nil {
		cf.Frame.blueprint.FillWithCopy(f, cf.Object)
	}
	return FrameDragging{f, vec2.Vec2{0, 0}}
}

// Frame deletion

type DeleteFrame struct {
	Frame *Frame
}

func (df DeleteFrame) Name() string    { return "Delete frame" }
func (df DeleteFrame) Keycode() string { return "KeyQ" }
func (df DeleteFrame) Activate(ui.TouchContext) ui.Action {
	df.Frame.Delete()
	return nil
}

// Toggle parameter (frame)

type ToggleParameter struct {
	*Frame
}

func (ToggleParameter) Name() string    { return "Toggle parameter" }
func (ToggleParameter) Keycode() string { return "KeyT" }
func (tp ToggleParameter) Activate(ui.TouchContext) ui.Action {
	tp.param = !tp.param
	return nil
}

// Add Parameter (frame / object parameter)

type AddParameter struct {
	*Frame
}

func (AddParameter) Name() string    { return "Add parameter" }
func (AddParameter) Keycode() string { return "KeyR" }
func (ap AddParameter) Activate(ui.TouchContext) ui.Action {
	ap.ForceFindFrameParameter("")
	return nil
}

// Delete Parameter

type DeleteParameter struct {
	*Frame
	ParameterIndex int
}

func (DeleteParameter) Name() string    { return "Delete parameter" }
func (DeleteParameter) Keycode() string { return "KeyZ" }
func (dp DeleteParameter) Activate(ui.TouchContext) ui.Action {
	dp.params = append(dp.params[:dp.ParameterIndex], dp.params[dp.ParameterIndex+1:]...)
	return nil
}

// Frame Dragging

type FrameDragging struct {
	Frame *Frame
	Cell  vec2.Vec2
}

func StartFrameDragging(p vec2.Vec2, f *Frame) FrameDragging {
	threshold := 1. / 3.
	frac := vec2.Div(p, f.size)
	var cell vec2.Vec2
	switch {
	case frac.X < threshold:
		cell.X = -1
	case frac.X < 2*threshold:
		cell.X = 0
	default:
		cell.X = 1
	}
	switch {
	case frac.Y < threshold:
		cell.Y = -1
	case frac.Y < 2*threshold:
		cell.Y = 0
	default:
		cell.Y = 1
	}
	return FrameDragging{f, cell}
}
func (d FrameDragging) Name() string                       { return "Move" }
func (d FrameDragging) Keycode() string                    { return "KeyF" }
func (d FrameDragging) Activate(ui.TouchContext) ui.Action { return d }
func (d FrameDragging) End(ui.TouchContext)                {}
func (d FrameDragging) Move(ctx ui.TouchContext) ui.Action {
	delta := ctx.Delta()
	e, cell := d.Frame, d.Cell
	start := e.pos
	e.size.Add(vec2.Mul(delta, cell))
	if cell.X == -1 {
		e.pos.X += delta.X
	}
	if cell.Y == -1 {
		e.pos.Y += delta.Y
	}
	if cell.X == 0 && cell.Y == 0 {
		e.pos.Add(delta)
	}
	if e.size.X < 30 {
		e.size.X = 30
	}
	if e.size.Y < 30 {
		e.size.Y = 30
	}
	end := e.pos
	movement := vec2.Sub(end, start)
	e.PropagateStiff(func(f *Frame) {
		if f != e {
			f.pos.Add(movement)
		}
	})
	return d
}

// Frame Params

type FrameElement interface {
	MyFrame() *Frame
}

type FrameParamDrag struct {
	Frame           *Frame
	blueprintObject *Object
	ParamName       string
}

func (d FrameParamDrag) Param() *FrameParameter {
	return d.Frame.ForceFindFrameParameter(d.ParamName)
}
func (d FrameParamDrag) Name() string    { return "Connect" }
func (d FrameParamDrag) Keycode() string { return "KeyF" }
func (d FrameParamDrag) Activate(t ui.TouchContext) ui.Action {
	dummyTarget := d.Frame.blueprint.MakeLinkTarget()
	dummyTarget.pos = t.AtTopBlueprint().Position()
	d.Param().Target = dummyTarget
	return d
}
func (d FrameParamDrag) Move(t ui.TouchContext) ui.Action {
	d.Param().Target.pos.Add(t.Delta())
	return d
}
func (d FrameParamDrag) End(ctx ui.TouchContext) {
	dummy := d.Param().Target
	ctx.AtTopBlueprint().Query(func(path ui.TreePath, p vec2.Vec2) ui.WalkAction {
		elem := path[len(path)-1]
		frameElement, ok := elem.(FrameElement)
		if ok {
			d.Param().Target = frameElement.MyFrame()
			return ui.Return
		}
		return ui.Explore
	})
	dummy.Delete()
}

type FrameParamCircle struct {
	Frame           *Frame
	Index           int
	Param           Parameter
	blueprintObject *Object
}

func (p FrameParamCircle) Draw(ctx *ui.Context2D) {
	ctx.BeginPath()
	ctx.Circle(vec2.Vec2{0, 0}, param_r)
	if p.Param != nil {
		ctx.FillStyle("#fff")
		ctx.Fill()
	}
	if p.Index < len(p.Frame.params) {
		ctx.LineWidth(2)
		ctx.StrokeStyle("#000")
		ctx.Stroke()
	}
}
func (p FrameParamCircle) Options(pos vec2.Vec2) []ui.Option {
	var name string
	if p.Index < len(p.Frame.params) {
		name = p.Frame.params[p.Index].Name
	} else {
		name = p.Param.Name()
	}
	return []ui.Option{FrameParamDrag{p.Frame, p.blueprintObject, name}}
}
func (p FrameParamCircle) Size(ui.TextMeasurer) ui.Box {
	return ui.Box{-param_r, param_r, param_r, -param_r}
}

type FrameParam struct {
	Frame           *Frame
	Index           int
	Param           Parameter
	blueprintObject *Object
}

func (p FrameParam) Children() []interface{} {
	return []interface{}{FrameParamCircle{p.Frame, p.Index, p.Param, p.blueprintObject}}
}
func (p FrameParam) Name() string {
	if p.Param != nil {
		return p.Param.Name()
	} else {
		return p.Frame.params[p.Index].Name
	}
}
func (p FrameParam) Draw(ctx *ui.Context2D) {
	ctx.FillStyle("#000")
	ctx.FillText(p.Name(), param_r+margin, -3)
}
func (p FrameParam) Options(vec2.Vec2) (opts []ui.Option) {
	if p.Index < len(p.Frame.params) {
		opts = append(opts, DeleteParameter{p.Frame, p.Index})
	}
	return
}
func (p FrameParam) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(vec2.Vec2{0, ParamOffset(p.Index)})
}
func (p FrameParam) Size(measurer ui.TextMeasurer) ui.Box {
	return ui.Box{-param_r, param_r + margin + measurer.MeasureText(p.Name()), param_r, -param_r}.Grow(margin / 2)
}
func (p FrameParam) GetText() string { return p.Name() }
func (p FrameParam) SetText(newName string) {
	p.Frame.ForceFindFrameParameter(p.Name()).Name = newName
}

type FrameParams struct {
	Frame           *Frame
	Object          *Object
	blueprintObject *Object
}

func CircleClicked(pos vec2.Vec2, touch vec2.Vec2) bool {
	return Dist(pos, touch) < param_r
}

func ParamOffset(i int) float64 {
	return param_r + margin + float64(i)*(param_r*2+margin)
}

func (p FrameParams) Draw(ctx *ui.Context2D) {
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
func (p FrameParams) Options(vec2.Vec2) []ui.Option { return nil }
func (p FrameParams) Transform(m ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(vec2.Vec2{param_r, p.Frame.size.Y})
}
func (p FrameParams) Children() (children []interface{}) {
	var objectParams []Parameter
	if p.Object != nil {
		objectParams = p.Object.typ.Parameters()
	}
	for i, _ := range p.Frame.params {
		_, param := GetParam(objectParams, p.Frame.params[i].Name)
		children = append(children, FrameParam{p.Frame, len(children), param, p.blueprintObject})
	}

	if p.Object != nil {
		for _, param := range objectParams {
			existing := p.Frame.FindFrameParameter(param.Name())
			if existing != nil {
				continue
			}
			children = append(children, FrameParam{p.Frame, len(children), param, p.blueprintObject})
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
	ctx.FillStyle("#000")
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
	ctx.StrokeStyle("#000")
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
	widgets = append(widgets, FrameParams{fobj.Frame, fobj.Object, fobj.BlueprintObject})
	return widgets
}

func (fobj FramedObject) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(fobj.Frame.pos)
}
