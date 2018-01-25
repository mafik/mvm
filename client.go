package mvm

import (
	"fmt"

	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/ui"
	"github.com/mafik/mvm/vec2"
)

type Client interface {
	Call(request string) (Event, error)
}

type ClientUI struct {
	Client  Client
	editing map[ui.Editable]bool
	focus   *Shell
	size    vec2.Vec2
}

var clients map[Client]*ClientUI = make(map[Client]*ClientUI)

func MakeClientUI(c Client) *ClientUI {
	if ui, ok := clients[c]; ok {
		return ui
	}
	ui := ClientUI{c, make(map[ui.Editable]bool), TheVM.root, vec2.Vec2{0, 0}}
	clients[c] = &ui
	return &ui
}

func (c *ClientUI) ToggleEditing(i ui.Editable) {
	if c.IsEditing(i) {
		delete(c.editing, i)
	} else {
		c.editing[i] = true
	}
}

func (c *ClientUI) IsEditing(i ui.Editable) bool {
	_, found := c.editing[i]
	return found
}

func (c *ClientUI) Focus() *Shell {
	return c.focus
}

func (c *ClientUI) MeasureText(text string) float64 {
	request := fmt.Sprintf("[\"measureText\",%q]", text)
	result, err := c.Client.Call(request)
	if err != nil {
		panic("Bad result of MeasureText: " + result.Type)
	}
	return result.Width
}

type GoUp struct {
	*ClientUI
}

func (GoUp) Name() string    { return "Go up" }
func (GoUp) Keycode() string { return "KeyW" }
func (gu GoUp) Activate(ui.TouchContext) ui.Action {
	if gu.focus.parent != nil {
		gu.focus = gu.focus.parent
	}
	return nil
}

type Quit struct{}

func (Quit) Name() string    { return "Quit" }
func (Quit) Keycode() string { return "Escape" }
func (Quit) Activate(ui.TouchContext) ui.Action {
	err := SaveImage()
	if err != nil {
		fmt.Println(err)
	}
	keep_running = false
	return nil
}

func (c *ClientUI) Options(vec2.Vec2) []ui.Option {
	return []ui.Option{Quit{}, GoUp{c}}
}

func (c *ClientUI) Children() (children []interface{}) {
	children = append(children, Background{c.size})
	if graphic, ok := c.focus.object.(GraphicObject); ok {
		children = append(children, graphic.MakeWidget(c.focus))
	}
	return
}

func (c *ClientUI) Transform(ui.TextMeasurer) matrix.Matrix {
	return matrix.Translate(vec2.Scale(c.size, 0.5))
}

type Background struct{ vec2.Vec2 }

func (b Background) Draw(ctx *ui.Context2D) {
	ctx.FillStyle("#ccc")
	ctx.FillRect(-b.X/2, -b.Y/2, b.X, b.Y)
}
