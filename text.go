package giopdf

import (
	"gioui.org/f32"
)

// BeginText starts a text object, setting the text matrix and text line matrix
// to the identity matrix.
func (c *Canvas) BeginText() {
	c.SetTextMatrix(1, 0, 0, 1, 0, 0)
}

// EndText ends a text object.
func (c *Canvas) EndText() {

}

// ShowText displays a string of text.
func (c *Canvas) ShowText(s string) {
	glyphs := c.font.ToGlyphs(s)
	vSize := c.fontSize
	hSize := c.fontSize * c.hScale / 100
	sizeMatrix := f32.NewAffine2D(hSize, 0, 0, 0, vSize, 0)
	for _, g := range glyphs {
		glyphSpace := c.textMatrix.Mul(sizeMatrix)
		c.Path = append(c.Path, transformPath(g.Outlines, glyphSpace)...)
		// TODO: clipping
		switch c.textRenderingMode {
		case 0, 4:
			c.Fill()
		case 1, 5:
			c.Stroke()
		case 2, 6:
			c.FillAndStroke()
		case 3, 7:
			// Invisible
		}
		c.textMatrix = c.textMatrix.Offset(f32.Pt(g.Width*hSize, 0))
	}
}

// Kern moves the next text character to the left the specified amount.
// The distance is in units of 1/1000 of an em.
func (c *Canvas) Kern(amount float32) {
	distance := c.fontSize * c.hScale / 100 * amount / 1000
	c.textMatrix = c.textMatrix.Offset(f32.Pt(-distance, 0))
}
