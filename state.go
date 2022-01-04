package giopdf

import "image/color"

type graphicsState struct {
	fillColor   color.NRGBA
	strokeColor color.NRGBA
	lineWidth   float32
}

func rgbColor(r, g, b float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// SetRGBStrokeColor implements the 'RG' operator.
func (s *graphicsState) SetRGBStrokeColor(r, g, b float32) {
	s.strokeColor = rgbColor(r, g, b)
}

// SetRGBFillColor implements the 'rg' operator.
func (s *graphicsState) SetRGBFillColor(r, g, b float32) {
	s.fillColor = rgbColor(r, g, b)
}

// SetLineWidth implements the 'w' operator.
func (s *graphicsState) SetLineWidth(w float32) {
	s.lineWidth = w
}
