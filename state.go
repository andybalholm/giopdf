package giopdf

import (
	"image/color"

	"gioui.org/op"
)

type graphicsState struct {
	fillColor   color.NRGBA
	strokeColor color.NRGBA
	lineWidth   float32

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

// SetLineWidth sets the width of the lines to use for stroking (outlining)
// shapes.
func (s *graphicsState) SetLineWidth(w float32) {
	s.lineWidth = w
}
