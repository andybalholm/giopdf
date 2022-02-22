package giopdf

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
)

type graphicsState struct {
	fillColor   color.NRGBA
	strokeColor color.NRGBA

	lineWidth  float32
	lineCap    int
	lineJoin   int
	miterLimit float32

	dashes    []float32
	dashPhase float32

	font              Font
	fontSize          float32
	hScale            float32
	textMatrix        f32.Affine2D
	lineMatrix        f32.Affine2D
	textRenderingMode int

	transforms    []op.TransformStack
	clippingPaths []clip.Stack
}

func rgbColor(r, g, b float32, alpha byte) color.NRGBA {
	return color.NRGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: alpha,
	}
}

func gray(g float32, alpha byte) color.NRGBA {
	return color.NRGBA{
		R: uint8(g * 255),
		G: uint8(g * 255),
		B: uint8(g * 255),
		A: alpha,
	}
}

// SetRGBStrokeColor sets the color to be used for stroking (outlining) shapes.
// The RGB values must be in the range from 0 to 1.
func (s *graphicsState) SetRGBStrokeColor(r, g, b float32) {
	s.strokeColor = rgbColor(r, g, b, s.strokeColor.A)
}

// SetRGBStrokeColor sets the color to be used for filling shapes.
// The RGB values must be in the range from 0 to 1.
func (s *graphicsState) SetRGBFillColor(r, g, b float32) {
	s.fillColor = rgbColor(r, g, b, s.fillColor.A)
}

// SetFillGray sets the fill color to a gray in the range from 0 (black) to
// 1 (white).
func (s *graphicsState) SetFillGray(g float32) {
	s.fillColor = gray(g, s.fillColor.A)
}

// SetStrokeGray sets the stroke color to a gray in the range from 0 (black) to
// 1 (white).
func (s *graphicsState) SetStrokeGray(g float32) {
	s.strokeColor = gray(g, s.strokeColor.A)
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

// SetDash sets the dash pattern for stroking paths.
func (s *graphicsState) SetDash(lengths []float32, phase float32) {
	s.dashes = lengths
	s.dashPhase = phase
}

// SetStrokeAlpha sets the opacity for stroking paths.
func (s *graphicsState) SetStrokeAlpha(alpha float32) {
	s.strokeColor.A = uint8(alpha * 255)
}

// SetFillAlpha sets the opacity for filling paths.
func (s *graphicsState) SetFillAlpha(alpha float32) {
	s.fillColor.A = uint8(alpha * 255)
}

// SetFont sets the font and size for text.
func (s *graphicsState) SetFont(f Font, size float32) {
	s.font = f
	s.fontSize = size
}

// SetTextMatrix sets the text matrix and the text line matrix.
func (s *graphicsState) SetTextMatrix(a, b, c, d, e, f float32) {
	s.lineMatrix = f32.NewAffine2D(a, c, e, b, d, f)
	s.textMatrix = s.lineMatrix
}

// TextMove starts a new line of text offset by x and y from the start of the
// current line.
func (s *graphicsState) TextMove(x, y float32) {
	s.lineMatrix = f32.NewAffine2D(1, 0, x, 0, 1, y).Mul(s.lineMatrix)
	s.textMatrix = s.lineMatrix
}

// SetHScale sets the horizontal scaling percent for text.
func (s *graphicsState) SetHScale(scale float32) {
	s.hScale = scale
}

// SetTextRendering sets the text rendering mode.
func (s *graphicsState) SetTextRendering(mode int) {
	s.textRenderingMode = mode
}
