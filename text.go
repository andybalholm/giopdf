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
	sizeMatrix := f32.NewAffine2D(c.fontSize, 0, 0, 0, c.fontSize, 0)
	for _, g := range glyphs {
		glyphSpace := c.textMatrix.Mul(sizeMatrix)
		c.Path = append(c.Path, transformPath(g.Outlines, glyphSpace)...)
		c.textMatrix = c.textMatrix.Offset(f32.Pt(g.Width*c.fontSize, 0))
	}
	c.Fill()
}