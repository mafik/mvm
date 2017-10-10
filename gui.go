package mvm

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

var margin float64 = 5
var buttonWidth float64 = 100
var buttonHeight float64 = 40

type Context2D struct {
	events  chan Event
	updates chan string
	buffer  bytes.Buffer
}

func (ctx *Context2D) sep() {
	if ctx.buffer.Len() == 0 {
		ctx.buffer.WriteRune('[')
	} else {
		ctx.buffer.WriteRune(',')
	}
}

func (ctx *Context2D) add(s string) {
	ctx.sep()
	ctx.buffer.WriteString(s)
}

func (ctx *Context2D) MarshalJSON() []byte {
	ctx.buffer.WriteRune(']')
	return ctx.buffer.Bytes()
}

func (ctx *Context2D) MeasureText(text string) float64 {
	ctx.updates <- fmt.Sprintf("[\"measureText\",%q]", text)
	result := <-ctx.events
	if result.Type != "MeasureText" {
		panic("Bad result of MeasureText")
	}
	return result.Width
}

func (ctx *Context2D) Translate(x, y float64) {
	ctx.add(fmt.Sprintf("[\"translate\",%g,%g]", x, y))
}
func (ctx *Context2D) FillText(text string, x, y float64) {
	ctx.add(fmt.Sprintf("[\"fillText\",%q,%g,%g]", text, x, y))
}
func (ctx *Context2D) FillRect(x, y, w, h float64) {
	ctx.add(fmt.Sprintf("[\"fillRect\",%g,%g,%g,%g]", x, y, w, h))
}
func (ctx *Context2D) Rect(x, y, w, h float64) {
	ctx.add(fmt.Sprintf("[\"rect\",%g,%g,%g,%g]", x, y, w, h))
}
func (ctx *Context2D) Arc(x, y, r, alpha, beta float64, anticlockwise bool) {
	ctx.add(fmt.Sprintf("[\"arc\",%g,%g,%g,%g,%g,%t]", x, y, r, alpha, beta, anticlockwise))
}
func (ctx *Context2D) Ellipse(x, y, rx, ry, rotation, alpha, beta float64, anticlockwise bool) {
	ctx.add(fmt.Sprintf("[\"ellipse\",%g,%g,%g,%g,%g,%g,%g,%t]", x, y, rx, ry, rotation, alpha, beta, anticlockwise))
}
func (ctx *Context2D) MoveTo(x, y float64) {
	ctx.add(fmt.Sprintf("[\"moveTo\",%g,%g]", x, y))
}
func (ctx *Context2D) LineTo(x, y float64) {
	ctx.add(fmt.Sprintf("[\"lineTo\",%g,%g]", x, y))
}
func (ctx *Context2D) SetLineDash(pattern []float64) {
	formatted := make([]string, len(pattern))
	for i, val := range pattern {
		formatted[i] = fmt.Sprintf("%g", val)
	}
	dash := strings.Join(formatted, ",")
	ctx.add(fmt.Sprintf("[\"setLineDash\",[%s]]", dash))
}
func (ctx *Context2D) Rotate(alpha float64) {
	ctx.add(fmt.Sprintf("[\"rotate\",%g]", alpha))
}
func (ctx *Context2D) Scale(s float64) {
	ctx.add(fmt.Sprintf("[\"scale\",%g,%g]", s, s))
}

func (ctx *Context2D) FillStyle(fill string) {
	ctx.add(fmt.Sprintf("[\"fillStyle\",%q]", fill))
}
func (ctx *Context2D) TextAlign(align string) {
	ctx.add(fmt.Sprintf("[\"textAlign\",%q]", align))
}
func (ctx *Context2D) TextBaseline(baseline string) {
	ctx.add(fmt.Sprintf("[\"textBaseline\",%q]", baseline))
}
func (ctx *Context2D) LineWidth(w float64) {
	ctx.add(fmt.Sprintf("[\"lineWidth\",%g]", w))
}
func (ctx *Context2D) StrokeStyle(style string) {
	ctx.add(fmt.Sprintf("[\"strokeStyle\",%q]", style))
}
func (ctx *Context2D) Font(font string) {
	ctx.add(fmt.Sprintf("[\"font\",%q]", font))
}

func (ctx *Context2D) Save()      { ctx.add("[\"save\"]") }
func (ctx *Context2D) Restore()   { ctx.add("[\"restore\"]") }
func (ctx *Context2D) BeginPath() { ctx.add("[\"beginPath\"]") }
func (ctx *Context2D) ClosePath() { ctx.add("[\"closePath\"]") }
func (ctx *Context2D) Fill()      { ctx.add("[\"fill\"]") }
func (ctx *Context2D) Stroke()    { ctx.add("[\"stroke\"]") }
func (ctx *Context2D) Clip()      { ctx.add("[\"clip\"]") }

// Utility functions

func (ctx *Context2D) Translate2(pos Vec2) {
	ctx.Translate(pos.X, pos.Y)
}

func (ctx *Context2D) Rect2(pos, size Vec2) {
	ctx.Rect(pos.X-size.X/2, pos.Y-size.Y/2, size.X, size.Y)
}

func (ctx *Context2D) Hourglass(color string) {
	const LW = 1.5   // line width
	const W = 8      // width
	const H = 10     // top height
	const H2 = H - 2 // angled height
	const h = 1.5    // gap height
	const F = 5      // fill
	ctx.Translate(-W-LW/2, -H-LW/2)
	ctx.BeginPath()
	ctx.MoveTo(-W, -H)
	ctx.LineTo(W, -H)
	ctx.LineTo(W, -H2)
	ctx.LineTo(2, -h)
	ctx.LineTo(2, h)
	ctx.LineTo(W, H2)
	ctx.LineTo(W, H)
	ctx.LineTo(-W, H)
	ctx.LineTo(-W, H2)
	ctx.LineTo(-2, h)
	ctx.LineTo(-2, -h)
	ctx.LineTo(-W, -H2)
	ctx.ClosePath()
	ctx.LineWidth(LW)
	ctx.StrokeStyle(color)
	ctx.Stroke()

	ctx.Translate(0, -LW/2*math.Sqrt(2)-h)
	ctx.BeginPath()
	ctx.MoveTo(0, 0)
	ctx.LineTo(-F, -F)
	ctx.LineTo(F, -F)
	ctx.ClosePath()
	ctx.FillStyle(color)
	ctx.Fill()

}

type ButtonList struct {
	Ctx                                        *Context2D
	Next                                       Vec2
	ActiveBg, ActiveFg, InactiveBg, InactiveFg string
	Dir                                        int
}

func (ctx *Context2D) NewButtonList() *ButtonList {
	return &ButtonList{ctx, Vec2{0, 0}, "#888", "#fff", "#ccc", "#000", 1}
}

func (l *ButtonList) AlignLeft(left float64) *ButtonList {
	l.Next.X = left + margin + buttonWidth/2
	return l
}

func (l *ButtonList) AlignTop(top float64) *ButtonList {
	l.Next.Y = top + margin + buttonHeight/2
	return l
}

func (l *ButtonList) AlignBottom(bottom float64) *ButtonList {
	l.Next.Y = bottom - margin - buttonHeight/2
	return l
}

func (l *ButtonList) Add(text string, active bool) *ButtonList {
	fg, bg := l.InactiveFg, l.InactiveBg
	if active {
		fg, bg = l.ActiveFg, l.ActiveBg
	}
	l.Ctx.BeginPath()
	l.Ctx.Rect(l.Next.X-buttonWidth/2, l.Next.Y-buttonHeight/2, buttonWidth, buttonHeight)
	l.Ctx.FillStyle(bg)
	l.Ctx.Fill()
	l.Ctx.FillStyle(fg)
	l.Ctx.TextAlign("center")
	l.Ctx.FillText(text, l.Next.X, l.Next.Y)
	l.Next.Y += (buttonHeight + margin) * float64(l.Dir)
	return l
}

type Drawable interface {
	Draw(ctx *Context2D)
}

func (OverlayLayer) Draw(ctx *Context2D) {
	return
}

func (ObjectLayer) Draw(ctx *Context2D) {
	return
}

func (FrameLayer) Draw(ctx *Context2D) {
	blueprint := TheVM.active.typ.(*Blueprint)
	for _, frame := range blueprint.Frames() {
		title := frame.Title()
		obj := frame.Object(TheVM.active)
		left := frame.pos.X - frame.size.X/2
		top := frame.pos.Y - frame.size.Y/2
		if obj != nil {
			// White background
			ctx.BeginPath()
			ctx.Rect2(frame.pos, frame.size)
			ctx.FillStyle("#fff")
			ctx.Fill()
			typ := obj.typ
			text := typ.String(obj.priv)
			ctx.FillStyle("#000")
			ctx.TextAlign("left")
			ctx.FillText(text, left+margin, top+margin+25)
			if typ == TextType {
				width := ctx.MeasureText(text)
				ctx.FillRect(left+margin+width, top+margin-5, 2, 25)
			}
			if obj.execute {
				ctx.FillStyle("#f00")
				ctx.BeginPath()
				ctx.Rect2(frame.pos, Add(frame.size, Vec2{10, 10}))
				ctx.Fill()
			}
			if obj.running {
				ctx.Save()
				ctx.Translate2(Add(frame.pos, Vec2{frame.size.X / 2, -frame.size.Y / 2}))
				ctx.Hourglass("#f00")
				ctx.Restore()
			}
		}
		// Black outline
		ctx.BeginPath()
		ctx.Rect2(frame.pos, Sub(frame.size, Vec2{2, 2}))
		ctx.StrokeStyle("#000")
		ctx.Stroke()
		ctx.FillText(title, left, top)
	}
}

func (ParamNameLayer) Draw(ctx *Context2D) {
	ctx.FillStyle("#000")
	for _, frame := range TheVM.active.typ.(*Blueprint).Frames() {
		params := frame.Parameters()
		for i, param := range params {
			pos := frame.ParamCenter(i)
			pos.Y -= 3
			pos.X += param_r + margin
			ctx.FillText(param.Name(), pos.X, pos.Y)
		}
	}
}

func (ParamLayer) Draw(ctx *Context2D) {
	for _, frame := range TheVM.active.typ.(*Blueprint).Frames() {
		local_params := frame.LocalParameters()
		type_params := frame.TypeParameters()
		params := frame.Parameters()
		n := len(params)
		if n > 0 {
			widgets.Line(
				Sub(frame.ParamCenter(0), Vec2{0, param_r + margin}),
				frame.ParamCenter(n-1))
		}

		for i, param := range params {
			pos := frame.ParamCenter(i)
			fill := ""
			if idx, _ := GetParam(type_params, param.Name()); idx >= 0 {
				fill = "#fff"
			}
			stroke := ""
			if idx, _ := GetParam(local_params, param.Name()); idx >= 0 {
				stroke = "#000"
			}
			widgets.Circle(pos, param_r, fill, stroke)
		}
	}
}

func (l LinkLayer) Draw(ctx *Context2D) {
	for _, frame := range TheVM.active.typ.(*Blueprint).frames {
		frame.DrawLinks(&widgets)
	}
}

var shadowOffset = Vec2{margin, margin}

func (BackgroundLayer) Draw(ctx *Context2D) {
	for _, t := range Pointer.Touched {
		if fd, ok := t.(*FrameDragging); ok {
			f := fd.frame
			f.PropagateStiff(func(f *Frame) {
				widgets.Rect(Add(f.pos, shadowOffset), f.size, "#ccc")
				for i, _ := range f.Parameters() {
					pos := Add(f.ParamCenter(i), shadowOffset)
					widgets.Circle(pos, param_r, "#ccc", "")
				}
			})
		}
	}
}
