package ui

import (
	"fmt"
	"math"
	"time"

	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/vec2"
)

type Widget interface {
	Options(vec2.Vec2) []Option
}

type MenuRoot interface {
	GetMenuLayer() *MenuLayer
}

type Action interface {
	Move(TouchContext) Action
	End(TouchContext)
}

type PreMoveAction interface {
	PreMove(TouchContext)
}

type Option interface {
	Name() string
	Keycode() string
	Activate(TouchContext) Action
}

type OptionContext struct {
	Path    WidgetPath
	Options []Option
}

func QueryOptions(travCtx TraversalContext, root interface{}, v vec2.Vec2) []OptionContext {
	result := []OptionContext{}
	WalkAtPoint(travCtx, root, v, func(path WidgetPath, v vec2.Vec2) WalkAction {
		last := path[len(path)-1]
		opts := OptionContext{path, nil}
		if editable, ok := last.(Editable); ok {
			opts.Options = append(opts.Options, EditOption{travCtx, editable})
		}
		if widget, ok := last.(Widget); ok {
			opts.Options = append(opts.Options, widget.Options(v)...)
		}
		if len(opts.Options) > 0 {
			result = append(result, opts)
		}
		return Explore
	})
	return result
}

func PickOption(options []OptionContext, keycode string, key string) (WidgetPath, Option) {
	for _, optCtx := range options {
		for _, opt := range optCtx.Options {
			if editOpt, ok := opt.(EditOption); ok && editKeys[keycode] {
				if editOpt.IsEditing() {
					return optCtx.Path, TypeOption{editOpt.Editable, keycode, key}
				}
			}
		}
	}
	for i, _ := range options {
		optCtx := options[len(options)-1-i]
		for _, opt := range optCtx.Options {
			if opt.Keycode() == keycode {
				return optCtx.Path, opt
			}
		}
	}
	return nil, nil
}

type MenuLayer struct {
	menus []*Menu
}

func MakeMenuLayer() MenuLayer {
	return MenuLayer{}
}

func (l *MenuLayer) Children() (children []interface{}) {
	for _, menu := range l.menus {
		children = append(children, menu)
	}
	return
}

func (l *MenuLayer) OpenMenu(ctx TouchContext, opts []OptionContext) *Menu {
	m := &Menu{ctx, l, ctx.Position(), opts, MakeSpring(100, 200, 5000), time.Now()}
	l.menus = append(l.menus, m)
	return m
}

type Spring struct {
	value       float64
	target      float64
	speed       float64
	lastUpdated time.Time
}

func MakeSpring(initialValue, target, speed float64) Spring {
	return Spring{initialValue, target, speed, time.Now()}
}

func (s *Spring) Update() {
	now := time.Now()
	t := now.Sub(s.lastUpdated).Seconds()
	s.lastUpdated = now
	if t > 0.05 {
		fmt.Println("Slow update:", t)
		t = 0.05
	}
	if t < 0.001 {
		fmt.Println("Fast update:", t)
		t = 0.001
	}
	a := s.speed
	s.speed += t * (s.target - s.value) * 1000
	s.speed *= math.Pow(0.65, t*100)
	b := s.speed
	s.value += t * (a + b) / 2
}

func (s *Spring) Value() float64 {
	return s.value
}

type Menu struct {
	ctx         TouchContext
	layer       *MenuLayer
	pos         vec2.Vec2
	opts        []OptionContext
	size        Spring
	lastUpdated time.Time
}

func (m *Menu) Move(ctx TouchContext) Action {
	return m
}

func (m *Menu) End(ctx TouchContext) {
	for i, other := range m.layer.menus {
		if m == other {
			m.layer.menus = append(m.layer.menus[:i], m.layer.menus[i+1:]...)
			return
		}
	}
	panic("Menu was missing from the layer!")
}

func (m *Menu) Transform(TextMeasurer) matrix.Matrix {
	return matrix.Translate(m.pos)
}

const Tau = math.Pi * 2

// alpha ∈ [0,Tau); beta > alpha
func (m *Menu) LoopOptions(cb func(o Option, alpha, beta float64)) {
	big := Tau / float64(len(m.opts))
	start := -big / 2
	for _, opts := range m.opts {
		small := big / float64(len(opts.Options))
		for j, o := range opts.Options {
			alpha := math.Mod(start+small*float64(j)+Tau, Tau)
			beta := alpha + small
			cb(o, alpha, beta)
		}
		start += big
	}
}

func (m *Menu) ChooseOption(angle float64) Option {
	angle2 := angle + Tau
	var chosen Option = nil
	m.LoopOptions(func(o Option, alpha, beta float64) {
		if (angle >= alpha && angle < beta) || (angle2 >= alpha && angle2 < beta) {
			chosen = o
		}
	})
	return chosen
}

func (m *Menu) Draw(ctx *Context2D) {
	m.size.Update()
	size := m.size.Value()

	outerR := size
	outerD := math.Asin(5. / outerR)
	innerR := size / 2
	innerD := math.Asin(5. / innerR)

	// Keep the menu centered below the finger
	now := time.Now()
	delta := vec2.Sub(m.ctx.Position(), m.pos)
	dist := vec2.Len(delta)
	if dist < innerR {
		deltaT := now.Sub(m.lastUpdated)
		alpha := 1 - math.Pow(0.5, deltaT.Seconds())
		m.pos.Add(vec2.Scale(delta, alpha))
	}
	m.lastUpdated = now

	// TODO: slide the menu away from screen edge

	// Activate the right option
	angle := math.Mod(math.Atan2(delta.Y, delta.X)+Tau, Tau)
	chosen := m.ChooseOption(angle)
	if dist > outerR {
		// TODO: correct tree path to match the activated option!
		a := chosen.Activate(m.ctx)
		if a == nil {
			m.ctx.Touch.action = nil
		} else {
			m.ctx.Touch.action.action = a
		}
		m.End(m.ctx)
		return
	}

	// Draw menu background
	ctx.FillStyle("#18c")
	m.LoopOptions(func(o Option, alpha, beta float64) {
		if beta-innerD*2 < alpha {
			return
		}
		if dist > innerR && o == chosen {
			ctx.FillStyle("#4bf")
		}
		ctx.BeginPath()
		ctx.Arc(0, 0, outerR, alpha+outerD, beta-outerD, false)
		ctx.Arc(0, 0, innerR, beta-innerD, alpha+innerD, true)
		ctx.Fill()
		if dist > innerR && o == chosen {
			ctx.FillStyle("#18c")
		}
	})

	// Draw menu labels
	ctx.FillStyle("#000")
	ctx.TextBaseline("middle")
	m.LoopOptions(func(o Option, alpha, beta float64) {
		ctx.Save()
		x := innerR + 5
		dir := (alpha + beta) / 2
		if dir > math.Pi/2 && dir < 3*math.Pi/2 {
			dir += math.Pi
			x = -x
			ctx.TextAlign("right")
		} else {
			ctx.TextAlign("left")
		}
		ctx.Rotate(dir)
		ctx.FillText(o.Name(), x, 0)
		ctx.Restore()
	})
}
