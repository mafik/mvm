package mvm

import (
	"encoding/json"
)

var margin float64 = 5.0

type Drawable interface {
	Draw() Widgets
}

type Widget struct {
	Type  string
	Pos   Vec2
	Scale float64
	Value interface{}
}

type Widgets struct {
	slice []*Widget
}

func (w Widgets) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.slice)
}

func (w *Widgets) AppendLocal(typ string, pos Vec2, value interface{}) {
	w.slice = append(w.slice, &Widget{typ, pos, 1, value})
}

func (w *Widgets) AppendGlobal(typ string, pos Vec2, value interface{}) {
	w.slice = append(w.slice, &Widget{typ, window.ToLocal(pos), window.scale, value})
}

type TextWidget struct {
	Text  string
	Align string
	Color string
	Caret bool
}

func MakeText(text string) *Widget {
	return &Widget{"text", Vec2{0, 0}, 1, &TextWidget{text, "center", "", false}}
}

func (w *Widgets) Text(text string, pos Vec2) *TextWidget {
	r := &TextWidget{text, "", "", false}
	w.AppendGlobal("text", pos, r)
	return r
}

type RectWidget struct {
	Vec2
	Color string
}

func (w *Widgets) Rect(pos, size Vec2, color string) *RectWidget {
	r := &RectWidget{size, color}
	w.AppendGlobal("rect", pos, r)
	return r
}

type ButtonWidget struct {
	Text *TextWidget
	Rect *RectWidget
}

var button_height float64 = 40.0
var button_width float64 = 100.0

func (w *Widgets) Button(text string, pos, size Vec2, fg, bg string) *ButtonWidget {
	r := &ButtonWidget{&TextWidget{text, "", fg, false}, &RectWidget{size, bg}}
	w.AppendGlobal("button", pos, r)
	return r
}

type ButtonList interface {
	AlignLeft(float64) ButtonList
	AlignBottom(float64) ButtonList
	AlignTop(float64) ButtonList
	PositionAt(Vec2) ButtonList
	Add(text string) ButtonList
	Add2(text string, active bool) ButtonList
	Dir(dir int) ButtonList
	Colors(bg, fg string) ButtonList
	Colors2(active_bg, active_fg, inactive_bg, inactive_fg string) ButtonList
}

type ButtonListContext struct {
	widgets                                        *Widgets
	next                                           Vec2
	active_bg, active_fg, inactive_bg, inactive_fg string
	dir                                            int
}

func (w *Widgets) ButtonList() ButtonList {
	return &ButtonListContext{w, Vec2{0, 0}, "#888", "#fff", "#ccc", "#000", 1}
}

func (c *ButtonListContext) Colors(bg, fg string) ButtonList {
	c.active_bg = bg
	c.active_fg = fg
	return c
}

func (c *ButtonListContext) Colors2(active_bg, active_fg, inactive_bg, inactive_fg string) ButtonList {
	c.active_bg = active_bg
	c.active_fg = active_fg
	c.inactive_bg = inactive_bg
	c.inactive_fg = inactive_fg
	return c
}

func (c *ButtonListContext) AlignLeft(left float64) ButtonList {
	c.next.X = left + margin + button_width/2
	return c
}

func (c *ButtonListContext) AlignTop(top float64) ButtonList {
	c.next.Y = top + margin + button_height/2
	return c
}

func (c *ButtonListContext) AlignBottom(bottom float64) ButtonList {
	c.next.Y = bottom - margin - button_height/2
	return c
}

func (c *ButtonListContext) PositionAt(pos Vec2) ButtonList {
	c.next = pos
	return c
}

func (c *ButtonListContext) Add(text string) ButtonList {
	return c.Add2(text, true)
}

func (c *ButtonListContext) Add2(text string, active bool) ButtonList {
	fg, bg := c.inactive_fg, c.inactive_bg
	if active {
		fg, bg = c.active_fg, c.active_bg
	}
	c.widgets.AppendLocal("button", c.next, ButtonWidget{
		&TextWidget{text, "center", fg, false},
		&RectWidget{Vec2{button_width, button_height}, bg}})
	c.next.Y += (button_height + margin) * float64(c.dir)
	return c
}

func (c *ButtonListContext) Dir(dir int) ButtonList {
	c.dir = dir
	return c
}

type LineWidget struct {
	Vec2
	Width  *float64
	Start  *Widget
	Middle *Widget
	End    *Widget
	Dash   []int
}

func (w *Widgets) Line(a, b Vec2) *LineWidget {
	l := &LineWidget{Sub(b, a), nil, nil, nil, nil, nil}
	w.AppendGlobal("line", a, l)
	return l
}

type CircleWidget struct {
	Fill   string
	Stroke string
	R      float64
}

func MakeCircle(r float64, fill, stroke string) *Widget {
	return &Widget{"circle", Vec2{0, 0}, 1, &CircleWidget{fill, stroke, r}}
}

func MakeArrow() *Widget {
	return &Widget{"arrow", Vec2{0, 0}, 1, nil}
}

func (w *Widgets) Circle(pos Vec2, r float64, fill, stroke string) {
	w.AppendGlobal("circle", pos, &CircleWidget{fill, stroke, r})
}

type HourglassWidget struct {
	Color string
}

func (w *Widgets) Hourglass(pos Vec2, color string) {
	w.AppendGlobal("hourglass", pos, &HourglassWidget{color})
}

func (h HighlightLayer) Draw() (widgets Widgets) {
	if fd, ok := Pointer.Touched.(*FrameDragging); ok {
		widgets.Rect(Add(fd.frame.pos, Vec2{margin, margin}), fd.frame.size, "#ccc")
	}
	return
}

func (f FrameLayer) Draw() (widgets Widgets) {
	blueprint := TheVM.active.typ.(*Blueprint)
	for _, frame := range blueprint.Frames() {
		title := frame.Title()
		obj := frame.Object(TheVM.active)
		typ := obj.typ
		s := ""
		if obj != nil {
			s = typ.String(obj.priv)
			if obj.execute {
				widgets.Rect(frame.pos, Add(frame.size, Vec2{10, 10}), "#f00")
			}
			if obj.running {
				widgets.Hourglass(Add(frame.pos, Vec2{frame.size.X / 2, -frame.size.Y / 2}), "#f00")
			}
		}
		w := widgets.Button(s, frame.pos, frame.size, "#000", "#fff")
		if typ == TextType {
			w.Text.Caret = true
		}
		widgets.Text(title, Sub(frame.pos, Scale(frame.size, .5)))
	}
	return
}

func (p ParamLayer) Draw() (widgets Widgets) {
	for _, frame := range TheVM.active.typ.(*Blueprint).Frames() {
		local_params := frame.LocalParameters()
		type_params := frame.Type().Parameters()
		params := frame.Parameters()
		n := len(params)
		if n > 0 {
			widgets.Line(
				Sub(frame.ParamCenter(0), Vec2{0, param_r + margin}),
				frame.ParamCenter(n-1))
		}

		for i, param := range params {
			pos := frame.ParamCenter(i)
			fill := ""
			if idx, _ := GetParam(type_params, param.Name()); idx >= 0 {
				fill = "#fff"
			}
			stroke := ""
			if idx, _ := GetParam(local_params, param.Name()); idx >= 0 {
				stroke = "#000"
			}
			widgets.Circle(pos, param_r, fill, stroke)
			pos.Y -= 1
			pos.X += param_r + margin
			widgets.Text(param.Name(), pos)
		}
	}
	return
}

func (l LinkLayer) Draw() (widgets Widgets) {
	for _, frame := range TheVM.active.typ.(*Blueprint).frames {
		frame.DrawLinks(&widgets)
	}
	return
}

func (BackgroundLayer) Draw() (widgets Widgets) {
	return
}
