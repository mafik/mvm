package mvm

import (
	"fmt"

	. "github.com/mafik/mvm/vec2"
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
		ll := GUI(e.Client)
		return ll.Input(t, e)
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

func (layer OverlayLayer) Input(t *Touch, e Event) Touching {
	bp := FindBlueprintBelow(layer.o, Pointer.TouchSnapshot)
	if bp == nil {
		return nil
	}
	if h := HandleEdit(e, bp, &bp.name, false); h != nil {
		return h
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

func (layer BackgroundLayer) Input(t *Touch, e Event) Touching {
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
		bp := layer.o.typ.(*Blueprint)
		f := bp.AddFrame()
		f.pos = Pointer.Global
		return &FrameDragging{f, Vec2{1, 1}}
	}
	return nil
}
