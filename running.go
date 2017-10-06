package mvm

import (
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

type Layer interface {
	Drawable
	Deletable
	Draggable
	Input
}

// LayerList of touchable elements
type LayerList []Layer
type OverlayLayer struct{}
type ObjectLayer struct{}
type FrameLayer struct{}
type ParamNameLayer struct{}
type ParamLayer struct{}
type LinkLayer struct{}
type BackgroundLayer struct{}

var GUI LayerList = []Layer{
	OverlayLayer{},
	ParamNameLayer{},
	LinkLayer{},
	ParamLayer{},
	ObjectLayer{},
	FrameLayer{},
	BackgroundLayer{},
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

func (frame *Frame) MarkForExecution() {
	if frame == nil {
		return
	}
	obj := frame.Object(TheVM.active)
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
		object := MakeArgs(o.frame, o.parent)["then"]
		if object != nil {
			object.MarkForExecution()
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
		for _, t := range Pointer.Touched {
			t.Move(&Pointer)
		}
	case "MouseDown":
	case "MouseUp":
	case "KeyDown":
		KeyDown(e)
	case "KeyUp":
		KeyUp(e)
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

var param_r float64 = 16.0

func Update(updates chan string) {
	widgets := Widgets{}
	for i := len(GUI) - 1; i >= 0; i-- {
		widgets.slice = append(widgets.slice, GUI[i].Draw().slice...)
	}

	bar := widgets.ButtonList().AlignTop(0).AlignLeft(0).Colors2("#000", "#fff", "#444", "#bbb")
	for it := TheVM.active; it != nil; it = it.parent {
		bar = bar.Add2(it.typ.Name(), it == TheVM.active)
	}
	_, dragging := Pointer.Touched["Shift"]
	widgets.ButtonList().
		Dir(-1).
		AlignLeft(0).
		AlignBottom(window.size.Y).
		Add2("navigate", nav).
		Add2("drag", dragging)

	if menu != nil {
		buttons := widgets.ButtonList().
			PositionAt(*menu).
			Add2("cancel", menu_index == 0)
		for i, t := range menu_types {
			buttons.Add2(t.Name(), menu_index == i+1)
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

func MakeArgs(f *Frame, blueprint *Object) Args {
	args := make(Args)
	args["self"] = FindObject(f, blueprint)
	for _, frame_parameter := range f.params {
		if frame_parameter.Target == nil {
			continue
		}
		args[frame_parameter.Name] = frame_parameter.FindParam(blueprint)
	}
	return args
}

func (ls *FrameParameter) FindParam(blueprint *Object) *Object {
	return FindObject(ls.Target, blueprint)
}

func (f *Frame) FindParam(blueprint *Object, param string) *Object {
	ls := f.GetLinkSet(param)
	if ls == nil {
		return nil
	}
	return ls.FindParam(blueprint)
}

func FindObject(f *Frame, blueprint *Object) (o *Object) {
	if f.param {
		o = blueprint.frame.FindParam(blueprint.parent, f.name)
	}
	if o == nil {
		m := blueprint.priv.(*Machine)
		o = m.objects[f]
	}
	return
}

func (o *Object) Run(events chan Event) {
	typ := o.typ
	args := MakeArgs(o.frame, o.parent)
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
P0
- Sequencing
- Blueprint-functions

P1
- Do a proper menu
- Sort out keybindings

P2
- Minimum & maximum number of arguments
- Browsing for all machines of a given type
- Highlighting frames with the right type

TODO:
- fix object dragging with CapsLock

Note: Events in complex objects
- complex objects can send many types of events
- they will write those events into their output frames

Note: Sequencing and data updates
- functions can be scheduled automatically with a "then" relation
- "then" relation is executed when in the start frame:
  - a simple function has completed
  - a data object was updated

Note: Scheduling
- scheduled functions are put into a queue (and marked with a number)
- functions in the queue are executed sequentially

Note: Background tasks
- long-running functions can go into background and let rest of the system run
- background functions send updates back to the main thread
- the updates are handled immediately

Note: Running blueprints
- execution starts with a frame called "run"

Note: Initializing frames
- frame can be initialized with a key shortcut
- initializing blueprints calls the "init" frame
- each type has a "Constructor" type - that builds it

Note: Imitation mode
- imitiation mode is not available in the root blueprint
- in imitation mode the first frame is called "run"
- one of the frames is marked as "last"
- every subsequent frame is linked to the "last" with a "then" relation
- object construction creates a frame with constructor
- writing into an object creates an external frame (outside of the blueprint)
  that writes given text into an object

*/
