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
	case "ContextMenu":
	case "Finished":
		s := e.Shell
		s.running = false
		then := MakeArgs(s.frame, s.parent).Get("then")
		if then != nil {
			then.MarkForExecution()
		}
	case "Interrupt":
		var ignore ui.TouchContext
		Quit{}.Activate(ignore)
	case "Size":
		clientUI.size = vec2.Vec2{e.Width, e.Height}
	case "MouseMove":
		Pointer.Move(clientUI, vec2.Vec2{e.X, e.Y})
	case "MouseDown":
	case "MouseUp":
	case "KeyDown":
		Pointer.StartAction(clientUI, clientUI, e.Code, e.Key)
	case "KeyUp":
		Pointer.EndAction(clientUI)
	case "Wheel":
		Pointer.Wheel = e.Y
		Pointer.StartAction(clientUI, clientUI, "Wheel", "Wheel")
		Pointer.EndAction(clientUI)
	case "TouchStart":
		for _, t := range e.Changed {
			clientUI.Touches[t.Id] = ui.MakeTouch(vec2.Vec2{t.X, t.Y})
			clientUI.Touches[t.Id].OpenMenu(clientUI, clientUI)
		}
	case "TouchMove":
		for _, t := range e.Changed {
			clientUI.Touches[t.Id].Move(clientUI, vec2.Vec2{t.X, t.Y})
		}
	case "TouchEnd":
		for _, t := range e.Changed {
			clientUI.Touches[t.Id].EndAction(clientUI)
			delete(clientUI.Touches, t.Id)
		}

	default:
		fmt.Printf("Unknown message: %s\n", e.Type)
	}
}

type FrameArgs struct {
	Frame     *Frame
	Blueprint *Shell
}

func (args FrameArgs) Get(name string) *Shell {
	elem := args.Frame.FindElement(name)
	if elem == nil {
		return nil
	}
	return elem.Target.Get(args.Blueprint)
}

func (args FrameArgs) Set(name string, s *Shell) {
	elem := args.Frame.FindElement(name)
	if elem == nil {
		// TODO: create a new frame and store the result there OR alert the user
		return
	}
	elem.Target.Set(args.Blueprint, s)
}

func MakeArgs(f *Frame, blueprint *Shell) Args {
	return FrameArgs{f, blueprint}
}

func (ls *FrameElement) FindParam(blueprint *Shell) *Shell {
	return FindShell(ls.Target.Frame(), blueprint)
}

func (f *Frame) FindParam(blueprint *Shell, param string) *Shell {
	ls := f.FindElement(param)
	if ls == nil {
		return nil
	}
	return ls.FindParam(blueprint)
}

func FindShell(f *Frame, machineShell *Shell) *Shell {
	if f.param {
		return machineShell.frame.FindParam(machineShell.parent, f.name)
	}
	m := machineShell.object.(*Machine)
	return m.shells[f]
}

func (s *Shell) Run(events chan Event) {
	object, ok := s.object.(RunnableObject)
	if !ok {
		s.execute = false
	}
	args := MakeArgs(s.frame, s.parent)
	fmt.Printf("Running %v...\n", object.Name())
	s.running = true
	s.execute = false
	go func() {
		object.Run(args)
		events <- Event{Type: "Finished", Shell: s}
	}()
}

var tasks chan *Shell = make(chan *Shell, 100)

type EventTouch struct {
	X, Y float64
	Id   int
}

type Event struct {
	Type          string
	Width, Height float64
	X, Y          float64
	Code          string
	Key           string
	Changed       []EventTouch
	Shell         *Shell
	Client        Client
}

var Pointer = ui.MakeTouch(vec2.Vec2{0, 0})

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

func (b *Blueprint) MakeLinkTarget() *Frame {
	f := b.AddFrame()
	f.size = vec2.Vec2{0, 0}
	f.Hidden = true
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
P0
- Sequencing
- Blueprint-functions

P1
- Do a proper menu

P2
- Minimum & maximum number of arguments
- Browsing for all machines of a given type
- Highlighting frames with the right type


Goffi integration TODOs:
- C.int
- alternative ( C.int, C.int ) -> ( C.int )
- dlopen flags object: RTLD_GLOBAL, RTLD_LAZY, ...
- OpenLibrary ( path string, flags C.int ) -> (Library, error)
- Library (REF)
- GetFunction ( Library, symbol string, Type, args Type...) -> (Function, error)
- Function (REF)


TODO:
- remove TouchContext & actionContext - put Action and TreePath in Touch
- delete FrameElement if it's not pointing anywhere
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
