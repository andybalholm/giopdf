package giopdf

import (
	"fmt"

	"gioui.org/op"
	"github.com/ledongthuc/pdf"
)

// RenderPage draws the contents of a PDF page to ops.
// The caller should do the appropriate transformation and scaling to ensure
// that the content is not rendered upside down. (The PDF coordinate system
// starts in the lower left, not in the upper left like Gio's.)
func RenderPage(ops *op.Ops, page pdf.Page) {
	r := newRenderer(ops)
	pdf.Interpret(page.V.Key("Contents"), r.do)
}

type renderer struct {
	*Canvas
}

func newRenderer(ops *op.Ops) *renderer {
	return &renderer{
		Canvas: NewCanvas(ops),
	}
}

func (r *renderer) do(stk *pdf.Stack, op string) {
	switch op {
	default:
		n := stk.Len()
		args := make([]pdf.Value, n)
		for i := n - 1; i >= 0; i-- {
			args[i] = stk.Pop()
		}
		fmt.Println(op, args)

	case "B":
		r.FillAndStroke()

	case "b":
		r.CloseFillAndStroke()

	case "c":
		y3 := float32(stk.Pop().Float64())
		x3 := float32(stk.Pop().Float64())
		y2 := float32(stk.Pop().Float64())
		x2 := float32(stk.Pop().Float64())
		y1 := float32(stk.Pop().Float64())
		x1 := float32(stk.Pop().Float64())
		r.CurveTo(x1, y1, x2, y2, x3, y3)

	case "f", "F":
		r.Fill()

	case "h":
		r.ClosePath()

	case "l", "m":
		y := float32(stk.Pop().Float64())
		x := float32(stk.Pop().Float64())
		switch op {
		case "l":
			r.LineTo(x, y)
		case "m":
			r.MoveTo(x, y)
		}

	case "RG", "rg":
		B := float32(stk.Pop().Float64())
		G := float32(stk.Pop().Float64())
		R := float32(stk.Pop().Float64())
		switch op {
		case "RG":
			r.SetRGBStrokeColor(R, G, B)
		case "rg":
			r.SetRGBFillColor(R, G, B)
		}

	case "S":
		r.Stroke()

	case "s":
		r.CloseAndStroke()

	case "w":
		r.SetLineWidth(float32(stk.Pop().Float64()))
	}
}
