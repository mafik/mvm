package mvm

import (
	"fmt"
	"math"

	"github.com/mafik/mvm/ui"
	"github.com/mafik/mvm/vec2"
)

var nav bool

func ButtonSize(contentSize float64) float64 {
	return math.Max(buttonHeight, contentSize+margin*2)
}

var keep_running bool = true

func ProcessEvent(e Event) {
	clientUI := MakeClientUI(e.Client)

	switch e.Type {
	case "RenderReady":
		ctx := ui.MakeContext2D(clientUI)
		ui.Draw(clientUI, clientUI, &ctx)
		up, _ := ctx.MarshalJSON()
		up2 := string(up)
		_, err := e.Client.Call(up2)
		if err != nil {
			fmt.Println(err)
		}
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
		var ignore ui.TouchContext
		Quit{}.Activate(ignore)
	case "Size":
		clientUI.size = vec2.Vec2{float64(e.Width), float64(e.Height)}
	case "MouseMove":
		Pointer.Move(clientUI, vec2.Vec2{e.X, e.Y})
	case "MouseDown":
	case "MouseUp":
	case "KeyDown":
		Pointer.StartAction(clientUI, clientUI, e.Code, e.Key)
	case "KeyUp":
		Pointer.EndAction(clientUI, e.Code)
	case "Wheel":
		Pointer.Wheel = e.Y
		Pointer.StartAction(clientUI, clientUI, "Wheel", "Wheel")
		Pointer.EndAction(clientUI, "Wheel")
		/*
			focus := clientUI.focus
			widget := focus.typ.MakeWidget(focus)
			path := ui.TreePath{clientUI, widget}
			transform := &focus.typ.(*Blueprint).transform
			a := ui.ToLocal(clientUI, path, Pointer.Curr)
			alpha := math.Exp(-e.Y / 200)
			transform.Scale(alpha)
			b := ui.ToLocal(clientUI, path, Pointer.Curr)
			fix := vec2.Sub(b, a)
			*transform = matrix.Multiply(matrix.Translate(fix), *transform) // apply translation before scaling
		*/
	default:
		fmt.Printf("Unknown message: %s\n", e.Type)
	}
}

func MakeArgs(f *Frame, blueprint *Object) Args {
	args := make(Args)
	args["self"] = FindObject(f, blueprint)
	for _, frame_parameter := range f.elems {
		if frame_parameter.Target == nil {
			continue
		}
		args[frame_parameter.Name] = frame_parameter.FindParam(blueprint)
	}
	return args
}

func (ls *FrameElement) FindParam(blueprint *Object) *Object {
	return FindObject(ls.Target, blueprint)
}

func (f *Frame) FindParam(blueprint *Object, param string) *Object {
	ls := f.FindElement(param)
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
	Width, Height float64
	X, Y          float64
	Code          string
	Key           string
	Object        *Object
	Client        Client
}

var Pointer = ui.MakeTouch()

/*
func LinkDist(p vec2.Vec2, link *Link) float64 {
	start, end := link.StartPos(), link.EndPos()
	l := Dist(start, end)
	if l == 0 {
		return Dist(p, start)
	}
	l = l * l
	alpha := Clamp(0, 1, Dot(vec2.Sub(p, start), vec2.Sub(end, start))/l)
	proj := vec2.Add(start, vec2.Scale(vec2.Sub(end, start), alpha))
	return Dist(p, proj)
}

func PointedLink(b *Blueprint, v Vec2) (best *Link) {
	best_dist := 8.0
	for _, frame := range b.frames {
		for i, _ := range frame.params {
			frame_parameter := &frame.params[i]
			if frame_parameter.Target == nil {
				continue
			}
			link := &Link{frame, frame_parameter}
			current_dist := LinkDist(v, link)
			if current_dist < best_dist {
				best, best_dist = link, current_dist
			}
		}
	}
	return
}
*/

// Links & Params

var LinkTargetType Type = &PrimitiveType{
	name: "",
	instantiate: func(me *Object) {
		me.frame.size = vec2.Vec2{0, 0}
	},
}

func (b *Blueprint) MakeLinkTarget() *Frame {
	f := b.AddFrame()
	f.Hidden = true
	b.FillWithNew(f, LinkTargetType)
	return f
}

var param_r float64 = 16
var textSize float64 = 20
var lineHeight float64 = 25
var margin float64 = 5
var buttonWidth float64 = 100
var buttonHeight float64 = textSize + margin*2
var textMargin float64 = margin * 1.75
var shadowOffset = vec2.Vec2{margin, margin}

/*
func DrawBreadcrumb(bp *Object, ctx *Context2D) {
	ctx.Save()
	ctx.Translate(margin, margin)
	ctx.TextAlign("center")
	for it := bp; it != nil; it = it.parent {
		text := "#bbb"
		bg := "#444"
		if it == bp {
			text = "#fff"
			bg = "#000"
		}
		ctx.FillStyle(bg)
		ctx.BeginPath()
		ctx.Rect(0, 0, buttonWidth, buttonHeight)
		ctx.Fill()
		if ctx.client.Editing(it.typ) {
			DrawEditingOverlay(ctx)
		}
		ctx.FillStyle(text)
		ctx.FillText(it.typ.Name(), buttonWidth/2, buttonHeight-textMargin)
		ctx.Translate(0, margin+buttonHeight)
	}
	ctx.Restore()
	return
}
*/

/*
P0
- Sequencing
- Blueprint-functions

P1
- Do a proper menu

P2
- Minimum & maximum number of arguments
- Browsing for all machines of a given type
- Highlighting frames with the right type

TODO:
- fix parameter renaming
- display number of blueprint instances
- blueprint renaming


Note: Manipulation
- Users shouldn't be able to manipulate objects directly - it can't be automated
- Instead the user should be functions that perform operations on objects

Note: Input events
- Events are delivered to widgets in reverse-draw order
- Keyboard events activate quick actions by default
- Objects can be placed into edit mode with the "Tab" key

Note: Keyboard
- On small touchscreen, the key events are sent into the crosshair on the center on the screen
- On large touchscreen, embedded keyboard objects & crosshairs can be added. Keyboard object sends the key events into the crosshairs
- On desktop PC
  - with Caps Lock: physical keyboard sends the key events into the crosshair attached to the cursor
  - without Caps Lock: keys open menu and (instantly) activate the appropriate option

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
