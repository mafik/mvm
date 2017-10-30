package mvm

import (
	"fmt"
)

var editKeys = map[string]bool{
	"Enter":        true,
	"Backspace":    true,
	"Space":        true,
	"KeyQ":         true,
	"KeyW":         true,
	"KeyE":         true,
	"KeyR":         true,
	"KeyT":         true,
	"KeyY":         true,
	"KeyU":         true,
	"KeyI":         true,
	"KeyO":         true,
	"KeyP":         true,
	"BracketLeft":  true,
	"BracketRight": true,
	"Backslash":    true,
	"KeyA":         true,
	"KeyS":         true,
	"KeyD":         true,
	"KeyF":         true,
	"KeyG":         true,
	"KeyH":         true,
	"KeyJ":         true,
	"KeyK":         true,
	"KeyL":         true,
	"Semicolon":    true,
	"Quote":        true,
	"KeyZ":         true,
	"KeyX":         true,
	"KeyC":         true,
	"KeyV":         true,
	"KeyB":         true,
	"KeyN":         true,
	"KeyM":         true,
	"Comma":        true,
	"Period":       true,
	"Slash":        true,
	"Backquote":    true,
	"Digit1":       true,
	"Digit2":       true,
	"Digit3":       true,
	"Digit4":       true,
	"Digit5":       true,
	"Digit6":       true,
	"Digit7":       true,
	"Digit8":       true,
	"Digit9":       true,
	"Digit0":       true,
	"Minus":        true,
	"Equal":        true,
}

func Edit(base string, e Event) string {
	switch e.Code {
	case "Backspace":
		if l := len(base); l > 0 {
			return base[:l-1]
		} else {
			return base
		}
	case "Enter":
		return base + "\n"
	default:
		return base + e.Key
	}
}

type Input interface {
	Input(t *Touch, e Event) Touching
}

func KeyDown(e Event) {
	f := func(t *Touch) Touching {
		return GUI.Input(t, e)
	}
	Pointer.BeginTouching(e.Code, f)
}

func KeyUp(e Event) {
	Pointer.EndTouching(e.Code)
}

func (ll *LayerList) Input(t *Touch, e Event) Touching {
	for _, l := range *ll {
		if ret := l.Input(t, e); ret != nil {
			return ret
		}
	}
	fmt.Println("Unhandled", e.Type, ", Code:", e.Code, ", Key:", e.Key)
	return nil
}

func HandleEdit(e Event, element interface{}, s *string, allowBreaks bool) Touching {
	if e.Code == "Tab" {
		e.Client.ToggleEditing(element)
		return NoopTouching{}
	}
	_, isEdit := editKeys[e.Code]
	if !isEdit {
		return nil
	}
	if e.Code == "Enter" && !allowBreaks {
		return nil
	}
	if e.Client.Editing(element) {
		*s = Edit(*s, e)
		return NoopTouching{}
	}
	return nil
}

func (OverlayLayer) Input(t *Touch, e Event) Touching {
	bp := Pointer.FindBlueprintBelow()
	if bp == nil {
		return nil
	}
	if h := HandleEdit(e, bp, &bp.name, false); h != nil {
		return h
	}
	return nil
}

func (ObjectLayer) Input(t *Touch, e Event) Touching {
	o := Pointer.FindObjectBelow()
	if o == nil {
		return nil
	}
	if _, ok := o.typ.(*Blueprint); e.Code == "KeyE" && ok {
		TheVM.active = o
	} else if o.typ == TextType {
		s := string(o.priv.([]byte))
		h := HandleEdit(e, o, &s, true)
		if h == nil {
			return nil
		}
		o.priv = []byte(s)
		return h
	} else if e.Code == "Space" {
		o.MarkForExecution()
	} else {
		return nil
	}
	return NoopTouching{}
}

func (FrameLayer) Input(t *Touch, e Event) Touching {
	f := Pointer.FindFrameTitleBelow()
	if f == nil {
		f = Pointer.FindFrameBelow()
	}
	if f == nil {
		return nil
	}
	initial_name := f.name
	result := HandleEdit(e, f, &f.name, false)
	if result != nil {
		if f.param {
			b := f.blueprint
			for instance, _ := range b.instances {
				if instance.frame == nil {
					continue
				}
				for i, _ := range instance.frame.params {
					ls := &instance.frame.params[i]
					if ls.Name == initial_name {
						ls.Name = f.name
					}
				}
			}
		}
		return result
	}

	switch e.Code {
	case "KeyF":
		return f.StartDrag(t)
	case "KeyS":
		o := Pointer.FindObjectBelow()
		if o == nil {
			return nil
		}
		b := o.frame.blueprint
		f2 := b.AddFrame()
		f2.pos = o.frame.pos
		f2.size = o.frame.size
		b.FillWithCopy(f2, o)
		return f2.Drag(t)
	case "KeyQ":
		f.Delete()
		return NoopTouching{}
	case "KeyR":
		f.params = append(f.params, FrameParameter{"", nil, false})
		return NoopTouching{}
	case "KeyE":
		f.param = !f.param
		return NoopTouching{}
	case "KeyT":
		new_bp := MakeBlueprint("new blueprint")
		parent_bp := TheVM.active.typ.(*Blueprint)
		parent_bp.FillWithNew(f, new_bp)
		return NoopTouching{}
	}
	return nil
}

func (ParamLayer) Input(t *Touch, e Event) Touching {
	frame, name := Pointer.FindParamBelow()
	if frame == nil {
		return nil
	}
	frameParam := frame.ForceGetLinkSet(name)
	result := HandleEdit(e, frameParam, &frameParam.Name, false)
	if result != nil {
		return result
	}
	switch e.Code {
	case "KeyF":
		target := frame.blueprint.MakeLinkTarget()
		target.pos = t.Global
		return frame.AddLink(name, target)
	case "KeyQ":
		index, _ := GetParam(frame.LocalParameters(), name)
		if index != -1 {
			frame.params = append(frame.params[:index], frame.params[index+1:]...)
			return NoopTouching{}
		}
	}
	return nil
}

func (LinkLayer) Input(t *Touch, e Event) Touching {
	l := t.PointedLink()
	if l == nil {
		return nil
	}
	switch e.Code {
	case "KeyF":
		target := TheVM.active.typ.(*Blueprint).MakeLinkTarget()
		target.pos = t.Global
		l.SetTarget(target)
		return l
	case "KeyZ":
		l.param.Stiff = !l.param.Stiff
		if l.param.Stiff {
			return l.source.Drag(t)
		} else {
			return NoopTouching{}
		}
	case "KeyQ:":
		l.Delete()
		return NoopTouching{}
	}
	return nil
}

type NavTouching struct{}

func (NavTouching) Move(*Touch) {}
func (NavTouching) End(*Touch)  { nav = false }
func Navigate() Touching {
	nav = true
	return NavTouching{}
}

func (BackgroundLayer) Input(t *Touch, e Event) Touching {
	switch e.Code {
	case "KeyW":
		parent := TheVM.active.parent
		if parent != nil {
			TheVM.active = parent
		}
		return NoopTouching{}
	case "KeyD":
		return Navigate()
	case "KeyS":
		bp := TheVM.active.typ.(*Blueprint)
		f := bp.AddFrame()
		f.pos = Pointer.Global
		return &FrameDragging{f, Vec2{1, 1}}
	}
	return nil
}
func (ParamNameLayer) Input(*Touch, Event) Touching { return nil }
