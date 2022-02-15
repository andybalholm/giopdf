package giopdf

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gioui.org/f32"
	"github.com/andybalholm/giopdf/pdf"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/fonts/type1"
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

// SimpleFontFromSFNT converts an SFNT (TrueType or OpenType) font to a
// SimpleFont with the specified encoding.
func SimpleFontFromSFNT(data []byte, encoding pdf.Value) (*SimpleFont, error) {
	f, err := truetype.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	ppem := f.Upem()
	scale := 1 / float32(ppem)
	fm := f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(scale, scale))

	var GIDEncoding [256]fonts.GID

	if !encoding.IsNull() {
		enc, err := getEncoding(encoding)
		if err != nil {
			return nil, err
		}
		for b, r := range enc.ByteToRune() {
			gi, ok := f.NominalGlyph(r)
			if ok {
				GIDEncoding[b] = gi
			}
		}
	} else {
		// The encoding in the font dictionary was missing, so use the font's
		// builtin encoding.
		fp, err := truetype.NewFontParser(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		ct, err := fp.CmapTable()
		if err != nil {
			return nil, err
		}
		cmap3_0 := ct.FindSubtable(truetype.CmapID{3, 0})
		if cmap3_0 != nil {
			// If the font contains a (3, 0) subtable, the range of character codes must be one
			// of the following: 0x0000 - 0x00FF, 0xF000 - 0xF0FF, 0xF100 - 0xF1FF, or
			// 0xF200 - 0xF2FF. Depending on the range of codes, each byte from the string is
			// prepended with the high byte of the range, to form a two-byte character, which
			// is used to select the associated glyph description from the subtable.
			iter := cmap3_0.Iter()
			for iter.Next() {
				code, gid := iter.Char()
				switch code & 0xFF00 {
				case 0x0000, 0xF000, 0xF100, 0xF200:
					b := code & 0xFF
					if GIDEncoding[b] == 0 {
						GIDEncoding[b] = gid
					}
				}
			}
		}
		cmap1_0 := ct.FindSubtable(truetype.CmapID{1, 0})
		if cmap1_0 != nil {
			// Otherwise, if the font contains a (1, 0) subtable, single bytes from the string are
			// used to look up the associated glyph descriptions from the subtable.
			iter := cmap1_0.Iter()
			for iter.Next() {
				code, gid := iter.Char()
				if code > 255 {
					continue
				}
				if GIDEncoding[code] == 0 {
					GIDEncoding[code] = gid
				}
			}
		}
	}

	simple := new(SimpleFont)

	for i, gi := range GIDEncoding {
		var g Glyph
		g.Width = f.HorizontalAdvance(gi) * scale

		gd := f.GlyphData(gi, ppem, ppem)
		switch gd := gd.(type) {
		case fonts.GlyphOutline:
			g.Outlines = glyphOutline(gd, fm)
		case fonts.GlyphSVG:
			g.Outlines = glyphOutline(gd.Outline, fm)
		case fonts.GlyphBitmap:
			return nil, errors.New("bitmap fonts not supported")
		}
		simple.Glyphs[i] = g
	}

	return simple, nil
}

func getEncoding(e pdf.Value) (simpleencodings.Encoding, error) {
	switch e.Kind() {
	case pdf.Null:
		return simpleencodings.Encoding{}, nil

	case pdf.Name:
		switch e.Name() {
		case "MacRomanEncoding":
			return simpleencodings.MacRoman, nil
		case "MaxExpertEncoding":
			return simpleencodings.MacExpert, nil
		case "WinAnsiEncoding":
			return simpleencodings.WinAnsi, nil
		default:
			return simpleencodings.Encoding{}, fmt.Errorf("unknown encoding: %v", e)
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
		return simpleencodings.Encoding{}, fmt.Errorf("invalid encoding: %v", e)
	}
}

// SimpleFontFromType1 converts a Type 1 font to a SimpleFont, using the
// encoding provided (normally taken from the font dictionary).
func SimpleFontFromType1(data []byte, encoding pdf.Value) (*SimpleFont, error) {
	f, err := type1.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

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
		g.Outlines = glyphOutline(gd.(fonts.GlyphOutline), fm)
		simple.Glyphs[i] = g
	}

	return simple, nil
}

// glyphOutline converts outline to our path format, transforming the points
// with fm.
func glyphOutline(outline fonts.GlyphOutline, fm f32.Affine2D) []PathElement {
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
	return p.Path
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
		return SimpleFontFromSFNT(data, f.V.Key("Encoding"))

	case "Type1":
		file := f.V.Key("FontDescriptor").Key("FontFile")
		if file.IsNull() {
			return nil, fmt.Errorf("%v does not have embedded font data", f.V.Key("BaseFont"))
		}
		data, err := io.ReadAll(file.Reader())
		if err != nil {
			return nil, err
		}
		return SimpleFontFromType1(data, f.V.Key("Encoding"))

	default:
		return nil, fmt.Errorf("%v is an unsupported font type (%v)", f.V.Key("BaseFont"), f.V.Key("Subtype"))
	}
}
