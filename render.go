package giopdf

import (
	"fmt"

	"gioui.org/op"
	"github.com/andybalholm/giopdf/pdf"
)

// RenderPage draws the contents of a PDF page to ops.
// The caller should do the appropriate transformation and scaling to ensure
// that the content is not rendered upside down. (The PDF coordinate system
// starts in the lower left, not in the upper left like Gio's.)
func RenderPage(ops *op.Ops, page pdf.Page) error {
	c := NewCanvas(ops)

	stream := page.V.Key("Contents")
	cs := pdf.NewContentStream(stream.Reader())

	for {
		args, op := cs.ReadInstruction()
		switch op {
		case "":
			return nil
		default:
			fmt.Println(args, op)

		case "B", "B*":
			c.FillAndStroke()
		case "BT":
			c.BeginText()
		case "c":
			c.CurveTo(args[0].Float32(), args[1].Float32(), args[2].Float32(), args[3].Float32(), args[4].Float32(), args[5].Float32())
		case "cm":
			c.Transform(args[0].Float32(), args[1].Float32(), args[2].Float32(), args[3].Float32(), args[4].Float32(), args[5].Float32())
		case "d":
			array := args[0]
			phase := args[1].Float32()
			dashes := make([]float32, array.Len())
			for i := range dashes {
				dashes[i] = array.Index(i).Float32()
			}
			c.SetDash(dashes, phase)
		case "Do":
			x := page.Resources().Key("XObject").Key(args[0].Name())
			if x.IsNull() {
				fmt.Printf("XObject resource missing: %v", args[0])
				continue
			}
			switch x.Key("Subtype").Name() {
			case "Image":
				img, err := decodeImage(x)
				if err != nil {
					fmt.Println(err)
					continue
				}
				c.Image(img)
			default:
				fmt.Printf("Unsupported XObject: %v\n", x)
			}
		case "ET":
			c.EndText()
		case "f", "f*":
			c.Fill()
		case "G":
			c.SetStrokeGray(args[0].Float32())
		case "g":
			c.SetFillGray(args[0].Float32())
		case "gs":
			gs := page.Resources().Key("ExtGState").Key(args[0].Name())
			if gs.IsNull() {
				fmt.Printf("ExtGState resource missing: %v", args[0])
				continue
			}
			for _, k := range gs.Keys() {
				v := gs.Key(k)
				switch k {
				case "Type":
					// ignore
				case "LW":
					c.SetLineWidth(v.Float32())
				case "LC":
					c.SetLineCap(v.Int())
				case "LJ":
					c.SetLineJoin(v.Int())
				case "ML":
					c.SetMiterLimit(v.Float32())
				case "D":
					array := v.Index(0)
					phase := v.Index(1).Float32()
					dashes := make([]float32, array.Len())
					for i := range dashes {
						dashes[i] = array.Index(i).Float32()
					}
					c.SetDash(dashes, phase)
				case "CA":
					c.SetStrokeAlpha(v.Float32())
				case "ca":
					c.SetFillAlpha(v.Float32())
				case "BM":
					if v.Name() != "Normal" {
						fmt.Printf("Unsupported blend mode: %v\n", v)
					}
				default:
					fmt.Printf("Unsupported graphics state parameter %v = %v\n", k, v)
				}
			}
		case "h":
			c.ClosePath()
		case "J":
			c.SetLineCap(args[0].Int())
		case "j":
			c.SetLineJoin(args[0].Int())
		case "l":
			c.LineTo(args[0].Float32(), args[1].Float32())
		case "m":
			c.MoveTo(args[0].Float32(), args[1].Float32())
		case "n":
			c.NoOpPaint()
		case "Q":
			c.Restore()
		case "q":
			c.Save()
		case "re":
			x := args[0].Float32()
			y := args[1].Float32()
			width := args[2].Float32()
			height := args[3].Float32()
			c.Rectangle(x, y, width, height)
		case "RG":
			c.SetRGBStrokeColor(args[0].Float32(), args[1].Float32(), args[2].Float32())
		case "rg":
			c.SetRGBFillColor(args[0].Float32(), args[1].Float32(), args[2].Float32())
		case "S":
			c.Stroke()
		case "Td":
			c.TextMove(args[0].Float32(), args[1].Float32())
		case "Tf":
			fd := page.Font(args[0].Name())
			if fd.V.IsNull() {
				fmt.Printf("Font resource missing: %v\n", args[0])
				continue
			}
			f, err := importPDFFont(fd)
			if err != nil {
				fmt.Println("Error importing font:", err)
				continue
			}
			c.SetFont(f, args[1].Float32())
		case "TJ":
			if c.font == nil {
				// TODO: remove
				continue
			}
			a := args[0]
			for i := 0; i < a.Len(); i++ {
				v := a.Index(i)
				switch v.Kind() {
				case pdf.Real, pdf.Integer:
					c.Kern(v.Float32())
				case pdf.String:
					c.ShowText(v.RawString())
				}
			}
		case "Tj":
			c.ShowText(args[0].RawString())
		case "Tm":
			c.SetTextMatrix(args[0].Float32(), args[1].Float32(), args[2].Float32(), args[3].Float32(), args[4].Float32(), args[5].Float32())
		case "Tr":
			c.SetTextRendering(args[0].Int())
		case "Tz":
			c.SetHScale(args[0].Float32())
		case "v":
			c.CurveV(args[0].Float32(), args[1].Float32(), args[2].Float32(), args[3].Float32())
		case "W", "W*":
			c.Clip()
		case "w":
			c.SetLineWidth(args[0].Float32())
		}
	}

	return nil
}
