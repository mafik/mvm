package mvm

import (
	"bytes"
	"fmt"
)

var editKeys = map[string]bool{
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
	switch e.Code {
	case "Tab":
		nav = false
	case "ControlLeft":
		if menu != nil && menu_index > 0 {
			t := menu_types[menu_index-1]
			blueprint := TheVM.active.typ.(*Blueprint)
			frame := blueprint.AddFrame()
			blueprint.FillWithNew(frame, t)
			frame.pos = window.ToGlobal(*menu)
		}
		menu = nil
	}
}

func (ll *LayerList) Input(t *Touch, e Event) Touching {
	for _, l := range *ll {
		if ret := l.Input(t, e); ret != nil {
			return ret
		}
	}
	fmt.Println("Unhandled", e.Type, e.Code, e.Key)
	return nil
}

func (OverlayLayer) Input(t *Touch, e Event) Touching {
	bp := Pointer.FindBlueprintBelow()
	_, isEdit := editKeys[e.Code]
	if bp != nil && isEdit {
		bp.name = Edit(bp.name, e)
		return NoopTouching{}
	}

	switch e.Code {
	case "Tab":
		nav = true
		return NoopTouching{}
	case "CapsLock":
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
	case "ShiftLeft":
		return GUI.Drag(t)
	case "ControlLeft":
		local := Pointer.Local
		menu = &local
		menu_types = nil
		for _, typ := range Types {
			menu_types = append(menu_types, typ)
		}
		menu_types = append(menu_types, MakeBlueprint("new blueprint"))
		return NoopTouching{}
	case "Delete":
		GUI.Delete(Pointer)
		return NoopTouching{}
	case "Escape":
		parent := TheVM.active.parent
		if parent != nil {
			TheVM.active = parent
		}
		return NoopTouching{}
	default:
		return nil
	}
}

func (FrameLayer) Input(t *Touch, e Event) Touching {
	o := Pointer.FindObjectBelow()
	if o == nil {
		f := Pointer.FindFrameTitleBelow()
		if f == nil {
			return nil
		}

		initial_name := f.name
		switch e.Key {
		case "Backspace":
			if l := len(f.name); l > 0 {
				f.name = f.name[:l-1]
			}
		case "Enter":
			f.param = !f.param
		default:
			f.name += e.Key
		}
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
		return NoopTouching{}
	}
	if _, ok := o.typ.(*Blueprint); e.Code == "Enter" && ok {
		TheVM.active = o
	} else if o.typ == TextType {
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
	} else if e.Code == "Space" {
		o.MarkForExecution()
	}
	return NoopTouching{}
}

func (ParamLayer) Input(t *Touch, e Event) Touching {
	return nil
}

func (LinkLayer) Input(t *Touch, e Event) Touching {
	return nil
}

func (BackgroundLayer) Input(t *Touch, e Event) Touching {
	return nil
}
