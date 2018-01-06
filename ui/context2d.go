package ui

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/vec2"
)

type TextMeasurer interface {
	MeasureText(text string) float64
}

type Context2D struct {
	textMeasurer TextMeasurer
	buffer       bytes.Buffer
}

func MakeContext2D(m TextMeasurer) Context2D {
	return Context2D{m, bytes.Buffer{}}
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
	result := ctx.textMeasurer.MeasureText(text)
	measureTextCache[text] = result
	return result
}

func (ctx *Context2D) Transform(m matrix.Matrix) {
	ctx.add(fmt.Sprintf("[\"transform\",%g,%g,%g,%g,%g,%g]", m[0], m[1], m[2], m[3], m[4], m[5]))
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

func (ctx *Context2D) Translate2(pos vec2.Vec2) {
	ctx.Translate(pos.X, pos.Y)
}

func (ctx *Context2D) Rect2(box Box) {
	ctx.Rect(box.Left, box.Top, box.Width(), box.Height())
}

func (ctx *Context2D) MoveTo2(pos vec2.Vec2) {
	ctx.MoveTo(pos.X, pos.Y)
}

func (ctx *Context2D) LineTo2(pos vec2.Vec2) {
	ctx.LineTo(pos.X, pos.Y)
}

func (ctx *Context2D) Circle(pos vec2.Vec2, radius float64) {
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
