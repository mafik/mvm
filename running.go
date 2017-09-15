/*

Objective: functions
Description: object encapsulates some a functionality

Objective: type checking
Description: functions can express capabilities for their arguments and return values

Q: Why there are Blueprints & Machines? Couldn't we do with Blueprints alone?
A: No, because then fixing one blueprint wouldn't fix all of it's instances. The
   user would have to update each one individually.

Q: Do we need "global" objects within blueprints?
A: ...?


*/

package mvm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
)

type Window struct {
	scale  float64
	size   Vec2 // local units
	center Vec2 // global units
}

func (w *Window) ToLocal(v Vec2) Vec2 {
	return Add(Scale(Sub(v, w.center), w.scale), Scale(w.size, 0.5))
}
func (w *Window) ToGlobal(v Vec2) Vec2 {
	return Add(w.center, Scale(Sub(v, Scale(w.size, 0.5)), 1/w.scale))
}

var window Window = Window{1, Vec2{0, 0}, Vec2{0, 0}}
var nav bool

var GUI Container = Container{
	elements: []interface{}{
		FrameLayer{},
		ParamLayer{},
		LinkLayer{},
	},
}

type FrameLayer struct{}
type ParamLayer struct{}
type LinkLayer struct{}

func NextOrder(frame *Frame, i int) (order int) {
	blue := frame.blueprint
	for link, _ := range blue.links {
		if link.A == frame && link.Param == i {
			order++
		}
	}
	return
}

// Menu
var menu *Vec2
var menu_index int
var menu_types []Type

func (f *Frame) ParamCenter(i int) Vec2 {
	ret := f.pos
	ret.X += -f.size.X/2 + param_r
	ret.Y += float64(i)*(param_r*2+margin) + f.size.Y/2 + margin + param_r
	return ret
}

func Input(e Event) {
	f := Pointer.PointedFrame()
	if f == nil {
		return
	}
	if f.typ == TextType {
		buf := bytes.NewBuffer(f.Object().priv.([]byte))
		switch e.Key {
		case "Backspace":
			cut := true
			trimmed := bytes.TrimRightFunc(buf.Bytes(), func(r rune) bool {
				ret := cut
				cut = false
				return ret
			})
			buf.Truncate(len(trimmed))
		case "Enter":
			buf.WriteString("\n")
		default:
			buf.WriteString(e.Key)
		}
		f.Object().priv = buf.Bytes()
	}
}

func (frame *Frame) MarkForExecution() {
	if frame == nil {
		return
	}
	obj := frame.Object()
	obj.MarkForExecution()
}

var keep_running bool = true

func ProcessEvent(e Event, updates chan string) {
	switch e.Type {
	case "RenderDone":
		return
	case "RenderReady":
		Update(updates)
		return
	}
	switch e.Type {
	case "ContextMenu":
	case "Finished":
		o := e.Object
		o.Running = false
		for _, object := range o.Args()["then"] {
			if object != nil {
				object.MarkForExecution()
			}
		}
	case "Interrupt":
		fmt.Println("Received SIGINT - saving the image and exiting")
		err := SaveImage()
		if err != nil {
			fmt.Println(err)
		}
		keep_running = false
		return
	case "Size":
		window.size = Vec2{float64(e.Width), float64(e.Height)}
	case "MouseMove":
		Pointer.Last = Pointer.TouchSnapshot
		Pointer.Local = Vec2{e.X, e.Y}
		Pointer.UpdateGlobal()
		if nav {
			window.center.Add(Pointer.Delta())
		}
		Pointer.UpdateGlobal()
		if Pointer.Touched != nil {
			Pointer.Touched.Move(&Pointer)
		}

	case "MouseDown":
	case "MouseUp":
	case "KeyDown":
		switch e.Code {
		case "Insert":
			bp := MakeBlueprint("new")
			TheVM.Blueprints[bp] = true
			TheVM.ActiveBlueprint = bp
		case "PageUp":
			blues, i := TheVM.OrderedBlueprints()
			TheVM.ActiveBlueprint = blues[(i+1)%len(blues)]
		case "PageDown":
			blues, i := TheVM.OrderedBlueprints()
			TheVM.ActiveBlueprint = blues[(i+len(blues)-1)%len(blues)]
		case "Tab":
			nav = true
		case "ShiftLeft":
			Pointer.BeginTouching("Shift", GUI.Drag)
		case "ControlLeft":
			local := Pointer.Local
			menu = &local
			menu_types = nil
			for _, typ := range Types {
				menu_types = append(menu_types, typ)
			}
		case "Enter":
			f := Pointer.PointedFrame()
			if f == nil {
				break
			}
			if f.typ == TextType {
				Input(e)
			} else {
				f.MarkForExecution()
			}
		case "Delete":
			GUI.Delete(Pointer)
		case "Space":
			Input(e)
		case "Backspace":
			Input(e)
		case "KeyQ":
			Input(e)
		case "KeyW":
			Input(e)
		case "KeyE":
			Input(e)
		case "KeyR":
			Input(e)
		case "KeyT":
			Input(e)
		case "KeyY":
			Input(e)
		case "KeyU":
			Input(e)
		case "KeyI":
			Input(e)
		case "KeyO":
			Input(e)
		case "KeyP":
			Input(e)
		case "BracketLeft":
			Input(e)
		case "BracketRight":
			Input(e)
		case "Backslash":
			Input(e)
		case "KeyA":
			Input(e)
		case "KeyS":
			Input(e)
		case "KeyD":
			Input(e)
		case "KeyF":
			Input(e)
		case "KeyG":
			Input(e)
		case "KeyH":
			Input(e)
		case "KeyJ":
			Input(e)
		case "KeyK":
			Input(e)
		case "KeyL":
			Input(e)
		case "Semicolon":
			Input(e)
		case "Quote":
			Input(e)
		case "KeyZ":
			Input(e)
		case "KeyX":
			Input(e)
		case "KeyC":
			Input(e)
		case "KeyV":
			Input(e)
		case "KeyB":
			Input(e)
		case "KeyN":
			Input(e)
		case "KeyM":
			Input(e)
		case "Comma":
			Input(e)
		case "Period":
			Input(e)
		case "Slash":
			Input(e)
		case "Backquote":
			Input(e)
		case "Digit1":
			Input(e)
		case "Digit2":
			Input(e)
		case "Digit3":
			Input(e)
		case "Digit4":
			Input(e)
		case "Digit5":
			Input(e)
		case "Digit6":
			Input(e)
		case "Digit7":
			Input(e)
		case "Digit8":
			Input(e)
		case "Digit9":
			Input(e)
		case "Digit0":
			Input(e)
		case "Minus":
			Input(e)
		case "Equal":
			Input(e)
		default:
			fmt.Printf("Pressed keycode: %s \"%s\"\n", e.Code, e.Key)
		}
	case "KeyUp":
		switch e.Code {
		case "Tab":
			nav = false
		case "ShiftLeft":
			Pointer.EndTouching("Shift")
		case "ControlLeft":
			if menu != nil && menu_index > 0 {
				t := menu_types[menu_index-1]
				blueprint := TheVM.ActiveBlueprint
				frame := blueprint.Add(t)
				frame.pos = window.ToGlobal(*menu)
			}
			menu = nil
		}
	case "Wheel":
		a := Pointer.Global
		window.scale *= math.Exp(-e.Y / 200)
		Pointer.UpdateGlobal()
		b := Pointer.Global
		fix := Sub(a, b)
		window.center.Add(fix)
		Pointer.UpdateGlobal()
	default:
		fmt.Printf("Unknown message: %s\n", e.Type)
	}
	switch {
	case menu != nil:
		local := Pointer.Local
		menu_index = int((local.Y - menu.Y + button_height/2) / (button_height + margin))
		if menu_index < 0 {
			menu_index = 0
		}
		if menu_index > len(Types) {
			menu_index = len(Types)
		}
	}
}

var param_r float64 = 15.0

func Update(updates chan string) {
	widgets := Widgets{}
	for _, l := range GUI.elements {
		if drawable, ok := l.(Drawable); ok {
			widgets.slice = append(widgets.slice, drawable.Widgets().slice...)
		}
	}
	// TODO: Replace with Drawable
	// if drag != nil {
	// 	drag.DragHighlight(&widgets)
	// }
	widgets.ButtonList().
		Dir(-1).
		AlignLeft(0).
		AlignBottom(window.size.Y).
		Add("navigate", nav).
		Add("drag", Pointer.Source == "Shift")

	if menu != nil {
		buttons := widgets.ButtonList().
			PositionAt(*menu).
			Add("cancel", menu_index == 0)
		for i, t := range menu_types {
			buttons.Add(t.Name(), menu_index == i+1)
		}
	}

	update, err := json.Marshal(widgets)
	if err != nil {
		updates <- fmt.Sprintf(
			`[{"type": "text", "x": %d, "y": %f, "value": "%v"}]`,
			0, window.size.Y, err)
		return
	}
	updates <- string(update)
}

func (o *Object) Args() Args {
	args := make(Args)
	m := o.machine
	typ := o.frame.typ
	for i, param := range typ.Parameters() {
		args[param.Name()] = make([]*Object, NextOrder(o.frame, i))
	}
	for l, _ := range o.frame.blueprint.links {
		if l.A == o.frame {
			name := typ.Parameters()[l.Param].Name()
			args[name][l.Order] = m.objects[l.B]
		}
	}
	return args
}

func (o *Object) Run(events chan Event) {
	typ := o.frame.typ
	args := o.Args()
	fmt.Printf("Running %v...\n", typ.Name())
	o.Running = true
	o.execute = false
	go func() {
		typ.Run(args)
		events <- Event{Type: "Finished", Object: o}
	}()
}

var tasks chan *Object = make(chan *Object, 100)

type Event struct {
	Type          string
	Width, Height uint
	X, Y          float64
	Code          string
	Key           string
	Object        *Object
}

/*

2017-04-30:
- Command object
- Asynchronous execution (for long-running commands)
- Indicator for running
- Directions for input and Output links
- Chrome renderer reconnects on restart
- Order of arguments
- Numbers on arg lines

2017-05-01:
- Adding and switching blueprints

Todo:
- Think about using immediate mode interface
- Display blueprint name
- Unique names for blueprints

- Show red message if connected type is wrong
- Minimum & maximum number of arguments

- Browsing machines

- Do a proper menu

*/
