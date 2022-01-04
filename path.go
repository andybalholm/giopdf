package giopdf

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// A PathElement represents a segement of a path.
type PathElement struct {
	// Op is the PDF path construction operator corresponding to this
	// PathElement. Possible values are:
	//
	//     m moveto
	//     l lineto
	//     c curveto
	//     h closepath
	Op byte

	// CP1 and CP2 are the control points for a bezier curve segment ('c').
	CP1, CP2 f32.Point
	End      f32.Point
}

type PathBuilder struct {
	Path         []PathElement
	currentPoint f32.Point
	lastMoveTo   f32.Point
}

// MoveTo implements the 'm' operator.
func (p *PathBuilder) MoveTo(x, y float32) {
	pt := f32.Point{x, y}
	p.currentPoint = pt
	p.lastMoveTo = pt

	if len(p.Path) > 0 && p.Path[len(p.Path)-1].Op == 'm' {
		p.Path[len(p.Path)-1].End = pt
		return
	}

	p.Path = append(p.Path, PathElement{Op: 'm', End: pt})
}

// LineTo implements the 'l' operator.
func (p *PathBuilder) LineTo(x, y float32) {
	pt := f32.Point{x, y}
	p.currentPoint = pt
	p.Path = append(p.Path, PathElement{Op: 'l', End: pt})
}

// CurveTo implements the 'c' operator.
func (p *PathBuilder) CurveTo(x1, y1, x2, y2, x3, y3 float32) {
	e := PathElement{
		Op:  'c',
		CP1: f32.Point{x1, y1},
		CP2: f32.Point{x2, y2},
		End: f32.Point{x3, y3},
	}
	p.currentPoint = e.End
	p.Path = append(p.Path, e)
}

// CurveV implements the 'v' operator.
func (p *PathBuilder) CurveV(x2, y2, x3, y3 float32) {
	e := PathElement{
		Op:  'c',
		CP1: p.currentPoint,
		CP2: f32.Point{x2, y2},
		End: f32.Point{x3, y3},
	}
	p.currentPoint = e.End
	p.Path = append(p.Path, e)
}

// CurveY implements the 'y' operator.
func (p *PathBuilder) CurveY(x1, y1, x3, y3 float32) {
	e := PathElement{
		Op:  'c',
		CP1: f32.Point{x1, y1},
		CP2: f32.Point{x3, y3},
		End: f32.Point{x3, y3},
	}
	p.currentPoint = e.End
	p.Path = append(p.Path, e)
}

// ClosePath implements the 'h' operator.
func (p *PathBuilder) ClosePath() {
	if len(p.Path) > 0 && p.Path[len(p.Path)-1].Op == 'h' {
		return
	}
	p.Path = append(p.Path, PathElement{Op: 'h'})
	p.currentPoint = p.lastMoveTo
}

// Rectangle implements the 're' operator.
func (p *PathBuilder) Rectangle(x, y, width, height float32) {
	p.MoveTo(x, y)
	p.LineTo(x+width, y)
	p.LineTo(x+width, y+height)
	p.LineTo(x, y+height)
	p.ClosePath()
}

func toPathSpec(ops *op.Ops, p []PathElement, alwaysClose bool) clip.PathSpec {
	var path clip.Path
	path.Begin(ops)
	closed := true

	for _, e := range p {
		switch e.Op {
		case 'm':
			if alwaysClose && !closed {
				path.Close()
			}
			path.MoveTo(e.End)
			closed = false

		case 'l':
			path.LineTo(e.End)
			closed = false

		case 'c':
			path.CubeTo(e.CP1, e.CP2, e.End)
			closed = false

		case 'h':
			path.Close()
			closed = true
		}
	}

	if alwaysClose && !closed {
		path.Close()
	}

	return path.End()
}
