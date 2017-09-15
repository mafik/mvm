package mvm

import (
	"encoding/json"
	"fmt"
)

var margin float64 = 5.0

type Drawable interface {
	Widgets() Widgets // TODO: Rename to Draw.
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
	PositionAt(Vec2) ButtonList
	Add(text string, active bool) ButtonList
	Dir(dir int) ButtonList
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

func (c *ButtonListContext) AlignLeft(left float64) ButtonList {
	c.next.X = left + margin + button_width/2
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

func (c *ButtonListContext) Add(text string, active bool) ButtonList {
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
	Color string
	R     float64
}

func MakeCircle(r float64, color string) *Widget {
	return &Widget{"circle", Vec2{0, 0}, 1, &CircleWidget{color, r}}
}

func MakeArrow() *Widget {
	return &Widget{"arrow", Vec2{0, 0}, 1, nil}
}

func (w *Widgets) Circle(pos Vec2, r float64, color string) {
	w.AppendGlobal("circle", pos, &CircleWidget{color, r})
}

type HourglassWidget struct {
	Color string
}

func (w *Widgets) Hourglass(pos Vec2, color string) {
	w.AppendGlobal("hourglass", pos, &HourglassWidget{color})
}

func (f FrameLayer) Widgets() (widgets Widgets) {
	blueprint := TheVM.ActiveBlueprint
	machine := blueprint.active_machine
	for frame, _ := range blueprint.frames {
		obj := machine.objects[frame]
		typ := frame.typ
		if obj.execute {
			widgets.Rect(frame.pos, Add(frame.size, Vec2{10, 10}), "#f00")
		}
		var s string
		if typ.String != nil {
			s = typ.String(obj.priv)
		} else {
			s = fmt.Sprintf("%#v", obj.priv)
		}
		w := widgets.Button(s, frame.pos, frame.size, "#000", "#fff")
		if typ == &TextType {
			w.Text.Caret = true
		}
		widgets.Text(typ.Name, Sub(frame.pos, Scale(frame.size, .5)))
		if obj.Running {
			widgets.Hourglass(Add(frame.pos, Vec2{frame.size.X / 2, -frame.size.Y / 2}), "#f00")
		}
	}
	return
}

func (p ParamLayer) Widgets() (widgets Widgets) {
	for frame, _ := range TheVM.ActiveBlueprint.frames {
		typ := frame.typ
		if param_count := len(typ.Parameters); param_count > 0 {
			widgets.Line(
				Sub(frame.ParamCenter(0), Vec2{0, param_r + margin}),
				frame.ParamCenter(param_count-1))
			for j, param := range typ.Parameters {
				pos := frame.ParamCenter(j)
				widgets.Circle(pos, param_r, "#fff")
				pos.Y -= 1
				pos.X += param_r + margin
				widgets.Text(param.Name, pos)
			}
		}
	}
	return
}

func (l LinkLayer) Widgets() (widgets Widgets) {
	for param_link, _ := range TheVM.ActiveBlueprint.links {
		param_link.AppendWidget(&widgets)
	}
	return
}