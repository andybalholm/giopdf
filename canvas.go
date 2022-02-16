package giopdf

import (
	"image"
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

	stateStack      []graphicsState
	setClippingPath bool

	ops *op.Ops
}

func NewCanvas(ops *op.Ops) *Canvas {
	return &Canvas{
		ops: ops,
		graphicsState: graphicsState{
			fillColor:   color.NRGBA{0, 0, 0, 255},
			strokeColor: color.NRGBA{0, 0, 0, 255},
			lineWidth:   1,
			miterLimit:  10,
			hScale:      100,
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

		Dashes: stroke.Dashes{
			Dashes: c.dashes,
			Phase:  c.dashPhase,
		},
	}

	switch c.lineCap {
	case 1:
		s.Cap = stroke.RoundCap
	case 2:
		s.Cap = stroke.SquareCap
	}

	switch c.lineJoin {
	case 0:
		s.Join = stroke.BevelJoin
		s.Miter = c.miterLimit
	case 1:
		s.Join = stroke.RoundJoin
		s.Miter = 0
	case 2:
		s.Join = stroke.BevelJoin
		s.Miter = 0
	}

	paint.FillShape(c.ops, c.strokeColor, s.Op(c.ops))
}

func (c *Canvas) finishPath() {
	if c.setClippingPath {
		ps := toPathSpec(c.ops, c.Path, true)
		cs := clip.Outline{ps}.Op().Push(c.ops)
		c.clippingPaths = append(c.clippingPaths, cs)
	}
	c.setClippingPath = false
	c.Path = c.Path[:0]
}

// Fill fills the current path.
func (c *Canvas) Fill() {
	c.fill()
	c.finishPath()
}

// Stroke strokes (outlines) the current path.
func (c *Canvas) Stroke() {
	c.stroke()
	c.finishPath()
}

// CloseAndStroke closes the current path before stroking it it.
func (c *Canvas) CloseAndStroke() {
	c.ClosePath()
	c.stroke()
	c.finishPath()
}

// FillAndStroke fills the current path and then strokes (outlines) it.
func (c *Canvas) FillAndStroke() {
	c.fill()
	c.stroke()
	c.finishPath()
}

// NoOpPaint finishes the current path without filling or stroking it.
// It is normally used to apply a clipping path after calling Clip.
func (c *Canvas) NoOpPaint() {
	c.finishPath()
}

// Clip causes the current path to be added to the clipping path after it is
// painted.
func (c *Canvas) Clip() {
	c.setClippingPath = true
}

// CloseFillAndStroke closes the current path before filling and stroking it.
func (c *Canvas) CloseFillAndStroke() {
	c.ClosePath()
	c.fill()
	c.stroke()
	c.finishPath()
}

// Save pushes a copy of the current graphics state onto the state stack.
func (c *Canvas) Save() {
	c.stateStack = append(c.stateStack, c.graphicsState)
	c.transforms = nil
	c.clippingPaths = nil
}

// Restore restores the graphics state, popping it off the stack.
func (c *Canvas) Restore() {
	// First pop off the TransformStack and clip.Stack entries that were saved since the last Save call.
	for i := len(c.transforms) - 1; i >= 0; i-- {
		c.transforms[i].Pop()
	}
	for i := len(c.clippingPaths) - 1; i >= 0; i-- {
		c.clippingPaths[i].Pop()
	}

	n := len(c.stateStack) - 1
	c.graphicsState = c.stateStack[n]
	c.stateStack = c.stateStack[:n]
}

// Transform changes the coordinate system according to the transformation
// matrix specified.
func (ca *Canvas) Transform(a, b, c, d, e, f float32) {
	m := f32.NewAffine2D(a, c, e, b, d, f)
	s := op.Affine(m).Push(ca.ops)
	ca.transforms = append(ca.transforms, s)
}

// Image draws an image. The image is placed in the unit square of the user
// coordinate system.
func (c *Canvas) Image(img image.Image) {
	io := paint.NewImageOp(img)
	size := io.Size()
	c.Save()
	c.Transform(1/float32(size.X), 0, 0, -1/float32(size.Y), 0, 1)
	io.Add(c.ops)
	paint.PaintOp{}.Add(c.ops)
	c.Restore()
}
