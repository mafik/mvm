package ui

type EditStatus interface {
	IsEditing(Editable) bool
	ToggleEditing(Editable)
}

type Editable interface {
	Sized
	GetText() string
	SetText(string)
}

type EditOption struct {
	EditStatus EditStatus
	Editable   Editable
}

func (EditOption) Name() string    { return "Edit" }
func (EditOption) Keycode() string { return "Tab" }
func (o EditOption) Activate(TouchContext) Action {
	o.EditStatus.ToggleEditing(o.Editable)
	return nil
}
func (o EditOption) IsEditing() bool {
	return o.EditStatus.IsEditing(o.Editable)
}

type TypeOption struct {
	Editable Editable
	keycode  string
	key      string
}

func (TypeOption) Name() string    { return "Type" }
func (TypeOption) Keycode() string { return "" }
func (t TypeOption) Activate(TouchContext) Action {
	t.Editable.SetText(Edit(t.Editable.GetText(), t.keycode, t.key))
	return nil
}

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

func Edit(base string, keycode string, key string) string {
	switch keycode {
	case "Backspace":
		if l := len(base); l > 0 {
			return base[:l-1]
		} else {
			return base
		}
	case "Enter":
		return base + "\n"
	default:
		return base + key
	}
}
