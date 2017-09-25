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
		HighlightLayer{},
		FrameLayer{},
		ParamLayer{},
		LinkLayer{},
	},
}

type HighlightLayer struct{}
type FrameLayer struct{}
type ParamLayer struct{}
type LinkLayer struct{}

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
	o := Pointer.FindObjectBelow()
	if o == nil {
		return
	}
	if o.typ == TextType {
		buf := bytes.NewBuffer(o.priv.([]byte))
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
		o.priv = buf.Bytes()
	}
}

func (frame *Frame) MarkForExecution() {
	if frame == nil {
		return
	}
	obj := frame.Object(TheVM.root)
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
		o.running = false
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
			blues, _ := TheVM.AvailableBlueprints()
			for _, b := range blues {
				menu_types = append(menu_types, b)
			}
			menu_types = append(menu_types, MakeBlueprint("new blueprint"))
		case "Space":
			o := Pointer.FindObjectBelow()
			if o == nil {
				break
			}
			if o.typ == TextType {
				Input(e)
			} else {
				o.MarkForExecution()
			}
		case "Delete":
			GUI.Delete(Pointer)
		case "Enter":
			o := Pointer.FindObjectBelow()
			if o == nil {
				break
			}
			if _, ok := o.typ.(*Blueprint); ok {
				TheVM.root = o
			} else {
				Input(e)
			}
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
				blueprint := TheVM.root.typ.(*Blueprint)
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
		if menu_index > len(menu_types) {
			menu_index = len(menu_types)
		}
	}
}

var param_r float64 = 15.0

func Update(updates chan string) {
	widgets := Widgets{}
	for _, l := range GUI.elements {
		if drawable, ok := l.(Drawable); ok {
			widgets.slice = append(widgets.slice, drawable.Draw().slice...)
		} else {
			fmt.Println("Warning: found non-Drawable layer")
		}
	}
	blueprints_slice, active_i := TheVM.AvailableBlueprints()
	for i, bp := range blueprints_slice {
		widgets.ButtonList().
			AlignLeft(float64(i)*(button_width+margin)).
			AlignTop(0).
			Colors("#000", "#fff", "#444", "#ccc").
			Add(bp.Name(), i == active_i)
	}
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
	m := o.parent.priv.(*Machine)
	for _, link_set := range o.frame.link_sets {
		for _, target := range link_set.Targets {
			args[link_set.ParamName] = append(args[link_set.ParamName], m.objects[target])
		}
	}
	return args
}

func (o *Object) Run(events chan Event) {
	typ := o.typ
	args := o.Args()
	fmt.Printf("Running %v...\n", typ.Name())
	o.running = true
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
TODO

P0
- Sequencing
- Blueprint-functions

P1
- Highlighting frames with the right type
- Do a proper menu

P2
- Minimum & maximum number of arguments
- Browsing machines

*/
