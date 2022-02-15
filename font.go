package giopdf

import (
	"bytes"
	"fmt"
	"io"

	"gioui.org/f32"
	"github.com/andybalholm/giopdf/pdf"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
	"github.com/benoitkugler/textlayout/fonts/type1"
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

func getEncoding(e pdf.Value) ([256]string, error) {
	switch e.Kind() {
	case pdf.Null:
		return [256]string{}, nil

	case pdf.Name:
		switch e.Name() {
		case "MacRomanEncoding":
			return [256]string(simpleencodings.MacRoman), nil
		case "MaxExpertEncoding":
			return [256]string(simpleencodings.MacExpert), nil
		case "WinAnsiEncoding":
			return [256]string(simpleencodings.WinAnsi), nil
		default:
			return [256]string{}, fmt.Errorf("unknown encoding: %v", e)
		}

	case pdf.Dict:
		enc, err := getEncoding(e.Key("BaseEncoding"))
		if err != nil {
			return enc, err
		}
		diff := e.Key("Differences")
		code := 0
		for i := 0; i < diff.Len(); i++ {
			item := diff.Index(i)
			switch item.Kind() {
			case pdf.Integer:
				code = item.Int() - 1
			case pdf.Name:
				code++
				enc[code] = item.Name()
			}
		}
		return enc, nil

	default:
		return [256]string{}, fmt.Errorf("invalid encoding: %v", e)
	}
}

// SimpleFontFromType1 converts a parsed Type 1 font to a SimpleFont, using the
// encoding provided (normally taken from the font dictionary).
func SimpleFontFromType1(f *type1.Font, encoding pdf.Value) (*SimpleFont, error) {
	nameToGID := map[string]fonts.GID{}
	for i := fonts.GID(0); ; i++ {
		name := f.GlyphName(i)
		if name == "" {
			break
		}
		nameToGID[name] = i
	}

	enc, err := getEncoding(encoding)
	if err != nil {
		return nil, err
	}
	for i, name := range enc {
		// Fill in the blanks with the font's builtin encoding.
		if name == "" {
			enc[i] = f.Encoding[i]
		}
	}

	fm := f32.NewAffine2D(f.FontMatrix[0], f.FontMatrix[2], f.FontMatrix[4], f.FontMatrix[1], f.FontMatrix[3], f.FontMatrix[5])

	simple := new(SimpleFont)
	for i, name := range enc {
		var g Glyph
		gi, ok := nameToGID[name]
		if !ok {
			continue
		}

		width := f.HorizontalAdvance(gi)
		g.Width = fm.Transform(f32.Pt(width, 0)).X

		gd := f.GlyphData(gi, 0, 0)
		if gd == nil {
			continue
		}
		outline := gd.(fonts.GlyphOutline)
		var p PathBuilder
		for _, seg := range outline.Segments {
			p0 := fm.Transform(f32.Point(seg.Args[0]))
			p1 := fm.Transform(f32.Point(seg.Args[1]))
			p2 := fm.Transform(f32.Point(seg.Args[2]))
			switch seg.Op {
			case fonts.SegmentOpMoveTo:
				p.ClosePath()
				p.MoveTo(p0.X, p0.Y)
			case fonts.SegmentOpLineTo:
				p.LineTo(p0.X, p0.Y)
			case fonts.SegmentOpQuadTo:
				p.QuadraticCurveTo(p0.X, p0.Y, p1.X, p1.Y)
			case fonts.SegmentOpCubeTo:
				p.CurveTo(p0.X, p0.Y, p1.X, p1.Y, p2.X, p2.Y)
			}
		}
		p.ClosePath()
		g.Outlines = p.Path
		simple.Glyphs[i] = g
	}

	return simple, nil
}

func importPDFFont(f pdf.Font) (Font, error) {
	switch f.V.Key("Subtype").Name() {
	case "TrueType":
		file := f.V.Key("FontDescriptor").Key("FontFile2")
		if file.IsNull() {
			return nil, fmt.Errorf("%v does not have embedded font data", f.V.Key("BaseFont"))
		}
		data, err := io.ReadAll(file.Reader())
		if err != nil {
			return nil, err
		}
		sf, err := sfnt.Parse(data)
		if err != nil {
			return nil, err
		}

		switch f.V.Key("Encoding").Name() {
		case "WinAnsiEncoding":
			return SimpleFontFromSFNT(sf, pdf.WinAnsiEncoding)
		case "MacRomanEncoding":
			return SimpleFontFromSFNT(sf, pdf.MacRomanEncoding)
		default:
			return nil, fmt.Errorf("%v: unknown encoding: %v", f.V.Key("BaseFont"), f.V.Key("Encoding"))
		}

	case "Type1":
		file := f.V.Key("FontDescriptor").Key("FontFile")
		if file.IsNull() {
			return nil, fmt.Errorf("%v does not have embedded font data", f.V.Key("BaseFont"))
		}
		data, err := io.ReadAll(file.Reader())
		if err != nil {
			return nil, err
		}
		t1f, err := type1.Parse(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return SimpleFontFromType1(t1f, f.V.Key("Encoding"))

	default:
		return nil, fmt.Errorf("%v is an unsupported font type (%v)", f.V.Key("BaseFont"), f.V.Key("Subtype"))
	}
}
