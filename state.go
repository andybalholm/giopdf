package giopdf

import (
	"image/color"

	"gioui.org/op"
)

type graphicsState struct {
	fillColor   color.NRGBA
	strokeColor color.NRGBA
	lineWidth   float32
	lineCap     int
	lineJoin    int
	miterLimit  float32

	transforms []op.TransformStack
}

func rgbColor(r, g, b float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

func gray(g float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(g * 255),
		G: uint8(g * 255),
		B: uint8(g * 255),
		A: 255,
	}
}

// SetRGBStrokeColor sets the color to be used for stroking (outlining) shapes.
// The RGB values must be in the range from 0 to 1.
func (s *graphicsState) SetRGBStrokeColor(r, g, b float32) {
	s.strokeColor = rgbColor(r, g, b)
}

// SetRGBStrokeColor sets the color to be used for filling shapes.
// The RGB values must be in the range from 0 to 1.
func (s *graphicsState) SetRGBFillColor(r, g, b float32) {
	s.fillColor = rgbColor(r, g, b)
}

// SetFillGray sets the fill color to a gray in the range from 0 (black) to
// 1 (white).
func (s *graphicsState) SetFillGray(g float32) {
	s.fillColor = gray(g)
}

// SetStrokeGray sets the stroke color to a gray in the range from 0 (black) to
// 1 (white).
func (s *graphicsState) SetStrokeGray(g float32) {
	s.strokeColor = gray(g)
}

// SetLineWidth sets the width of the lines to use for stroking (outlining)
// shapes.
func (s *graphicsState) SetLineWidth(w float32) {
	s.lineWidth = w
}

// SetLineCap sets the style for the caps at the end of lines:
// 0 (butt cap), 1 (round cap), or 2 (square cap).
func (s *graphicsState) SetLineCap(c int) {
	s.lineCap = c
}

// SetLineJoin sets the style for the joins at corners of stroked paths:
// 0 (miter join), 1 (round join), or 2 (bevel join).
func (s *graphicsState) SetLineJoin(j int) {
	s.lineJoin = j
}

// SetMiterLimit sets the limit for the length of a mitered join.
func (s *graphicsState) SetMiterLimit(m float32) {
	s.miterLimit = m
}
