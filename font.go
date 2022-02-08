package giopdf

import (
	"fmt"

	"gioui.org/f32"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// A Glyph represents a character from a font. It uses a coordinate system
// where the origin is at the left end of the baseline of the glyph, y
// increases vertically, and the font size is one unit.
type Glyph struct {
	Outlines []PathElement
	Width    float32
}

// A Font converts text strings to slices of Glyphs, so that they can be
// displayed.
type Font interface {
	ToGlyphs(s string) []Glyph
}

// A SimpleFont is a font with a simple 8-bit encoding.
type SimpleFont struct {
	Glyphs [256]Glyph
}

func (f *SimpleFont) ToGlyphs(s string) []Glyph {
	result := make([]Glyph, len(s))
	for i := 0; i < len(s); i++ {
		result[i] = f.Glyphs[s[i]]
	}
	return result
}

func scalePoint(p fixed.Point26_6, ppem fixed.Int26_6) f32.Point {
	return f32.Pt(float32(p.X)/float32(ppem), -float32(p.Y)/float32(ppem))
}

// SimpleFontFromSFNT converts a parsed SFNT (TrueType or OpenType) font to a
// SimpleFont with the specified encoding.
func SimpleFontFromSFNT(sf *sfnt.Font, encoding [256]rune) (*SimpleFont, error) {
	buf := new(sfnt.Buffer)
	ppem := fixed.I(int(sf.UnitsPerEm()))

	simple := new(SimpleFont)

	for i, c := range encoding {
		var g Glyph
		gi, err := sf.GlyphIndex(buf, c)
		if err != nil {
			return nil, err
		}
		width, err := sf.GlyphAdvance(buf, gi, ppem, font.HintingNone)
		if err != nil {
			return nil, err
		}
		g.Width = float32(width) / float32(ppem)

		segments, err := sf.LoadGlyph(buf, gi, ppem, nil)
		if err != nil {
			return nil, err
		}
		var p PathBuilder
		for _, seg := range segments {
			p0 := scalePoint(seg.Args[0], ppem)
			p1 := scalePoint(seg.Args[1], ppem)
			p2 := scalePoint(seg.Args[2], ppem)
			switch seg.Op {
			case sfnt.SegmentOpMoveTo:
				p.ClosePath()
				p.MoveTo(p0.X, p0.Y)
			case sfnt.SegmentOpLineTo:
				p.LineTo(p0.X, p0.Y)
			case sfnt.SegmentOpQuadTo:
				p.QuadraticCurveTo(p0.X, p0.Y, p1.X, p1.Y)
			case sfnt.SegmentOpCubeTo:
				p.CurveTo(p0.X, p0.Y, p1.X, p1.Y, p2.X, p2.Y)
			}
		}
		p.ClosePath()
		g.Outlines = p.Path
		simple.Glyphs[i] = g
	}

	return simple, nil
}

func importPDFFont(f model.Font) (Font, error) {
	switch f := f.(type) {
	case model.FontTrueType:
		file := f.FontDescriptor.FontFile
		if file == nil {
			return nil, fmt.Errorf("%v does not have embedded font data", f.FontName())
		}
		data, err := file.Decode()
		if err != nil {
			return nil, err
		}
		sf, err := sfnt.Parse(data)
		if err != nil {
			return nil, err
		}

		var byte2rune map[byte]rune
		switch f.Encoding {
		case model.MacRomanEncoding:
			byte2rune = simpleencodings.MacRoman.ByteToRune()
		case model.MacExpertEncoding:
			byte2rune = simpleencodings.MacExpert.ByteToRune()
		case model.WinAnsiEncoding:
			byte2rune = simpleencodings.WinAnsi.ByteToRune()
		}
		if byte2rune == nil {
			return nil, fmt.Errorf("%v: unknown encoding: %v", f.FontName, f.Encoding)
		}

		var encoding [256]rune
		for b, r := range byte2rune {
			encoding[b] = r
		}

		simple, err := SimpleFontFromSFNT(sf, encoding)
		if err != nil {
			return nil, err
		}
		return simple, nil

	default:
		return nil, fmt.Errorf("%v is an unsupported font type (%T)", f.FontName(), f)
	}
}