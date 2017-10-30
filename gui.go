package mvm

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

var margin float64 = 5
var buttonWidth float64 = 100
var buttonHeight float64 = textSize + margin*2
var textMargin float64 = margin * 1.75
var textSize float64 = 20

type TextMeasurer interface {
	MeasureText(text string) float64
}

type Context2D struct {
	client Client
	buffer bytes.Buffer
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

func (ctx *Context2D) MarshalJSON() ([]byte, error) {
	ctx.buffer.WriteRune(']')
	return ctx.buffer.Bytes(), nil
}

var measureTextCache map[string]float64 = make(map[string]float64)

func (ctx *Context2D) MeasureText(text string) float64 {
	if width, ok := measureTextCache[text]; ok {
		return width
	}
	request := fmt.Sprintf("[\"measureText\",%q]", text)
	result, err := ctx.client.Call(request)
	if err != nil {
		panic("Bad result of MeasureText: " + result.Type)
	}
	measureTextCache[text] = result.Width
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

func (ctx *Context2D) MoveTo2(pos Vec2) {
	ctx.MoveTo(pos.X, pos.Y)
}

func (ctx *Context2D) LineTo2(pos Vec2) {
	ctx.LineTo(pos.X, pos.Y)
}

func (ctx *Context2D) Circle(pos Vec2, radius float64) {
	ctx.Arc(pos.X, pos.Y, radius, 0, 2*math.Pi, false)
}

func (ctx *Context2D) Arrow(size float64) {
	ctx.BeginPath()
	ctx.MoveTo(0, 0)
	ctx.Arc(0, 0, size, math.Pi*5/6, math.Pi*7/6, false)
	ctx.Fill()
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

func (ctx *Context2D) NewButtonList(dir int) *ButtonList {
	return &ButtonList{ctx, Vec2{0, 0}, "#888", "#fff", "#ccc", "#000", dir}
}

func (l *ButtonList) PositionAt(pos Vec2) *ButtonList {
	l.Next = pos
	return l
}

func (l *ButtonList) Colors(activeBg, activeFg, inactiveBg, inactiveFg string) *ButtonList {
	l.ActiveBg = activeBg
	l.ActiveFg = activeFg
	l.InactiveBg = inactiveBg
	l.InactiveFg = inactiveFg
	return l
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
	l.Ctx.FillText(text, l.Next.X, l.Next.Y+5)
	l.Next.Y += (buttonHeight + margin) * float64(l.Dir)
	return l
}

type Drawable interface {
	Draw(ctx *Context2D)
}

func DrawEditingOverlay(ctx *Context2D) {
	ctx.LineWidth(2)
	ctx.StrokeStyle("#f00")
	ctx.Stroke()
	ctx.FillStyle("rgba(255,128,128,0.2)")
	ctx.Fill()
}

func (OverlayLayer) Draw(ctx *Context2D) {
	ctx.Save()
	ctx.Translate(margin, margin)
	for it := TheVM.active; it != nil; it = it.parent {
		text := "#bbb"
		bg := "#444"
		if it == TheVM.active {
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

func (ObjectLayer) Draw(ctx *Context2D) {
	return
}

func (ctx *Context2D) TransformToGlobal(w *Window) {
	ctx.Translate2(Scale(w.size, 0.5))
	ctx.Scale(w.scale)
	ctx.Translate2(Neg(w.center))
}

func (FrameLayer) Draw(ctx *Context2D) {
	ctx.Save()
	ctx.TransformToGlobal(&window)
	blueprint := TheVM.active.typ.(*Blueprint)
	for _, frame := range blueprint.Frames() {
		obj := frame.Object(TheVM.active)
		left := frame.pos.X - frame.size.X/2
		top := frame.pos.Y - frame.size.Y/2

		title := frame.Title()
		titleWidth := ctx.MeasureText(title)
		frameTitleWidth := ButtonSize(titleWidth)

		if obj != nil {
			typeName := obj.typ.Name()

			// White background
			ctx.FillStyle("#fff")
			ctx.BeginPath()
			ctx.Rect2(frame.pos, Sub(frame.size, Vec2{2, 2}))
			ctx.Fill()
			ctx.FillRect(frame.ObjectLeft(ctx), frame.ObjectTop(), frame.ObjectWidth(obj, ctx), buttonHeight)
			typ := obj.typ
			text := typ.String(obj.priv)
			lines := strings.Split(text, "\n")
			ctx.FillStyle("#000")
			ctx.TextAlign("left")

			ctx.FillText(typeName, left+margin+frameTitleWidth, top-textMargin)

			for i, line := range lines {
				ctx.FillText(line, left+margin, top+margin+float64(i+1)*25)
			}

			if typ == TextType {
				width := ctx.MeasureText(lines[len(lines)-1])
				ctx.FillRect(left+margin+width, top+margin+float64(len(lines)-1)*25, 2, 30)
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
		if ctx.client.Editing(obj) {
			DrawEditingOverlay(ctx)
		} else {
			ctx.LineWidth(2)
			ctx.StrokeStyle("#000")
			ctx.Stroke()
		}

		ctx.FillStyle("#000")
		ctx.BeginPath()
		ctx.Rect(left, top, frameTitleWidth, -ButtonSize(0))
		ctx.Fill()
		if ctx.client.Editing(frame) {
			DrawEditingOverlay(ctx)
		}
		ctx.FillStyle("#fff")
		ctx.TextAlign("left")
		ctx.FillText(title, left+margin, top-textMargin)
	}
	ctx.Restore()
}

func (ParamNameLayer) Draw(ctx *Context2D) {
	ctx.Save()
	ctx.TransformToGlobal(&window)
	ctx.FillStyle("#000")
	ctx.TextAlign("left")
	for _, frame := range TheVM.active.typ.(*Blueprint).Frames() {
		params := frame.Parameters()
		for i, param := range params {
			pos := frame.ParamCenter(i)
			pos.Y -= 3
			pos.X += param_r + margin
			ctx.FillText(param.Name(), pos.X, pos.Y)
		}
	}
	ctx.Restore()
}

func (ParamLayer) Draw(ctx *Context2D) {
	ctx.Save()
	ctx.TransformToGlobal(&window)
	ctx.LineWidth(2)
	for _, frame := range TheVM.active.typ.(*Blueprint).Frames() {
		local_params := frame.LocalParameters()
		type_params := frame.TypeParameters()
		params := frame.Parameters()
		n := len(params)
		if n > 0 {
			ctx.BeginPath()
			ctx.MoveTo2(Sub(frame.ParamCenter(0), Vec2{0, param_r + margin}))
			ctx.LineTo2(frame.ParamCenter(n - 1))
			ctx.Stroke()
		}

		for i, param := range params {
			pos := frame.ParamCenter(i)
			ctx.BeginPath()
			ctx.Circle(pos, param_r)
			if idx, _ := GetParam(type_params, param.Name()); idx >= 0 {
				ctx.FillStyle("#fff")
				ctx.Fill()
			}
			if idx, _ := GetParam(local_params, param.Name()); idx >= 0 {
				frameParameter := frame.ForceGetLinkSet(param.Name())
				if ctx.client.Editing(frameParameter) {
					DrawEditingOverlay(ctx)
				} else {
					ctx.StrokeStyle("#000")
					ctx.Stroke()
				}
			}
		}
	}
	ctx.Restore()
}

func (l LinkLayer) Draw(ctx *Context2D) {
	ctx.Save()
	ctx.TransformToGlobal(&window)
	for _, frame := range TheVM.active.typ.(*Blueprint).frames {

		for i, _ := range frame.params {
			frame_parameter := &frame.params[i]
			if frame_parameter.Target == nil {
				continue
			}
			link := Link{frame, frame_parameter}
			start := link.StartPos()
			end := link.EndPos()
			delta := Sub(end, start)
			length := Len(delta)

			ctx.Save()
			ctx.Translate2(start)
			ctx.Rotate(math.Atan2(delta.Y, delta.X))

			if frame_parameter.Stiff && start != end {
				// white line outline
				ctx.StrokeStyle("#fff")
				ctx.BeginPath()
				ctx.MoveTo(0, 0)
				ctx.LineTo(length-4, 0)
				ctx.LineWidth(6.0)
				ctx.Stroke()

				// white arrow outline
				ctx.Save()
				ctx.FillStyle("#fff")
				ctx.Translate(length+4, 0)
				ctx.Arrow(13 + 6)
				ctx.Fill()
				ctx.Restore()
			}
			// line
			ctx.StrokeStyle("#000")
			ctx.LineWidth(2)
			ctx.BeginPath()
			ctx.MoveTo(0, 0)
			ctx.LineTo(length-5, 0)
			ctx.Stroke()

			// black circle
			ctx.FillStyle("#000")
			ctx.BeginPath()
			ctx.Circle(Vec2{0, 0}, param_r/4)
			ctx.Fill()

			// black arrow
			ctx.Translate(length, 0)
			ctx.Arrow(13)
			ctx.Fill()

			ctx.Restore()
		}
	}
	ctx.Restore()
}

var shadowOffset = Vec2{margin, margin}

func (BackgroundLayer) Draw(ctx *Context2D) {
	ctx.Save()
	ctx.FillStyle("#ddd")
	ctx.BeginPath()
	ctx.Rect(0, 0, window.size.X, window.size.Y)
	ctx.Fill()
	ctx.TransformToGlobal(&window)
	for _, t := range Pointer.Touched {
		if fd, ok := t.(*FrameDragging); ok {
			f := fd.frame
			ctx.FillStyle("#ccc")
			ctx.BeginPath()
			f.PropagateStiff(func(f *Frame) {
				ctx.Rect2(Add(f.pos, shadowOffset), f.size)
				for i, _ := range f.Parameters() {
					pos := Add(f.ParamCenter(i), shadowOffset)
					ctx.Circle(pos, param_r)
					ctx.ClosePath()
				}
			})
			ctx.Fill()
		}
	}
	ctx.Restore()
}
