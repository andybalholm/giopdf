package giopdf

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/x/stroke"
)

// A Canvas implements the PDF imaging model, drawing to a Gio operations list.
// Most of its methods correspond directly to PDF page description operators.
type Canvas struct {
	PathBuilder
	graphicsState

	ops *op.Ops
}

func NewCanvas(ops *op.Ops) *Canvas {
	return &Canvas{
		ops: ops,
		graphicsState: graphicsState{
			fillColor:   color.NRGBA{0, 0, 0, 255},
			strokeColor: color.NRGBA{0, 0, 0, 255},
			lineWidth:   1,
		},
	}
}

func (c *Canvas) fill() {
	ps := toPathSpec(c.ops, c.Path, true)
	paint.FillShape(c.ops, c.fillColor, clip.Outline{ps}.Op())
}

func (c *Canvas) stroke() {
	var p stroke.Path
	var pos, lastMove f32.Point
	for _, e := range c.Path {
		switch e.Op {
		case 'm':
			lastMove = e.End
			pos = e.End
			p.Segments = append(p.Segments, stroke.MoveTo(e.End))
		case 'l':
			pos = e.End
			p.Segments = append(p.Segments, stroke.LineTo(e.End))
		case 'c':
			pos = e.End
			p.Segments = append(p.Segments, stroke.CubeTo(e.CP1, e.CP2, e.End))
		case 'h':
			if pos != lastMove {
				p.Segments = append(p.Segments, stroke.LineTo(lastMove))
				pos = lastMove
			}
		}
	}

	s := stroke.Stroke{
		Path:  p,
		Width: c.lineWidth,

		Cap:   stroke.FlatCap,
		Join:  stroke.BevelJoin,
		Miter: 10,
	}

	// TODO: support dashes, joins, and caps

	paint.FillShape(c.ops, c.strokeColor, s.Op(c.ops))
}

// Fill fills the current path.
func (c *Canvas) Fill() {
	c.fill()
	c.Path = c.Path[:0]
}

// Stroke strokes (outlines) the current path.
func (c *Canvas) Stroke() {
	c.stroke()
	c.Path = c.Path[:0]
}

// CloseAndStroke closes the current path before stroking it it.
func (c *Canvas) CloseAndStroke() {
	c.ClosePath()
	c.stroke()
	c.Path = c.Path[:0]
}

// FillAndStroke fills the current path and then strokes (outlines) it.
func (c *Canvas) FillAndStroke() {
	c.fill()
	c.stroke()
	c.Path = c.Path[:0]
}

// CloseFillAndStroke closes the current path before filling and stroking it.
func (c *Canvas) CloseFillAndStroke() {
	c.ClosePath()
	c.fill()
	c.stroke()
	c.Path = c.Path[:0]
}
