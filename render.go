package giopdf

import (
	"fmt"
	"io"

	"gioui.org/op"
	"github.com/andybalholm/giopdf/pdf"
	"github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/reader/parser"
)

// RenderPage draws the contents of a PDF page to ops.
// The caller should do the appropriate transformation and scaling to ensure
// that the content is not rendered upside down. (The PDF coordinate system
// starts in the lower left, not in the upper left like Gio's.)
func RenderPage(ops *op.Ops, page pdf.Page) error {
	r := newRenderer(ops)
	r.page = page

	stream := page.V.Key("Contents")

	decoded, err := io.ReadAll(stream.Reader())
	if err != nil {
		return err
	}
	operations, err := parser.ParseContent(decoded, nil)
	if err != nil {
		return err
	}
	r.do(operations)

	return nil
}

type renderer struct {
	*Canvas

	page pdf.Page
}

func newRenderer(ops *op.Ops) *renderer {
	return &renderer{
		Canvas: NewCanvas(ops),
	}
}

func (r *renderer) do(operations []contentstream.Operation) {
	for _, op := range operations {
		switch op := op.(type) {
		case contentstream.OpBeginText:
			r.BeginText()
		case contentstream.OpClosePath:
			r.ClosePath()
		case contentstream.OpConcat:
			m := op.Matrix
			r.Transform(m[0], m[1], m[2], m[3], m[4], m[5])
		case contentstream.OpCubicTo:
			r.CurveTo(op.X1, op.Y1, op.X2, op.Y2, op.X3, op.Y3)
		case contentstream.OpCurveTo1:
			r.CurveV(op.X2, op.Y2, op.X3, op.Y3)
		case contentstream.OpEndText:
			r.EndText()
		case contentstream.OpFill, contentstream.OpEOFill:
			r.Fill()
		case contentstream.OpFillStroke, contentstream.OpEOFillStroke:
			r.FillAndStroke()
		case contentstream.OpLineTo:
			r.LineTo(op.X, op.Y)
		case contentstream.OpMoveTo:
			r.MoveTo(op.X, op.Y)
		case contentstream.OpRestore:
			r.Restore()
		case contentstream.OpSave:
			r.Save()
		case contentstream.OpSetDash:
			r.SetDash(op.Dash.Array, op.Dash.Phase)
		case contentstream.OpSetExtGState:
			gs := r.page.Resources().Key("ExtGState").Key(string(op.Dict))
			if gs.IsNull() {
				fmt.Printf("ExtGState resource missing: %v", op.Dict)
				continue
			}
			for _, k := range gs.Keys() {
				v := gs.Key(k)
				switch k {
				case "Type":
					// ignore
				case "LW":
					r.SetLineWidth(v.Float32())
				case "LC":
					r.SetLineCap(v.Int())
				case "LJ":
					r.SetLineJoin(v.Int())
				case "ML":
					r.SetMiterLimit(v.Float32())
				case "D":
					array := v.Index(0)
					phase := v.Index(1).Float32()
					dashes := make([]float32, array.Len())
					for i := range dashes {
						dashes[i] = array.Index(i).Float32()
					}
					r.SetDash(dashes, phase)
				case "CA":
					r.SetStrokeAlpha(v.Float32())
				case "ca":
					r.SetFillAlpha(v.Float32())
				case "BM":
					if v.Name() != "Normal" {
						fmt.Printf("Unsupported blend mode: %v\n", v)
					}
				default:
					fmt.Printf("Unsupported graphics state parameter %v = %v\n", k, v)
				}
			}
		case contentstream.OpSetFillGray:
			r.SetFillGray(op.G)
		case contentstream.OpSetFillRGBColor:
			r.SetRGBFillColor(op.R, op.G, op.B)
		case contentstream.OpSetFont:
			fd := r.page.Font(string(op.Font))
			if fd.V.IsNull() {
				fmt.Printf("Font resource missing: $v", op.Font)
				continue
			}
			f, err := importPDFFont(fd)
			if err != nil {
				fmt.Println("Error importing font:", err)
				continue
			}
			r.SetFont(f, op.Size)
		case contentstream.OpSetLineCap:
			r.SetLineCap(int(op.Style))
		case contentstream.OpSetLineJoin:
			r.SetLineJoin(int(op.Style))
		case contentstream.OpSetLineWidth:
			r.SetLineWidth(op.W)
		case contentstream.OpSetStrokeGray:
			r.SetStrokeGray(op.G)
		case contentstream.OpSetStrokeRGBColor:
			r.SetRGBStrokeColor(op.R, op.G, op.B)
		case contentstream.OpShowText:
			r.ShowText(op.Text)
		case contentstream.OpStroke:
			r.Stroke()
		case contentstream.OpTextMove:
			r.TextMove(op.X, op.Y)

			/* TODO:
			case contentstream.OpXObject:
				x := r.resources.XObject[op.XObject]
				switch x := x.(type) {
				case *model.XObjectImage:
					img, err := decodeImage(x)
					if err != nil {
						fmt.Println(err)
						continue
					}
					r.Image(img)
				default:
					fmt.Printf("XObject (%T): %v\n", x, x)
				}
			*/
		default:
			fmt.Printf("%T: %v\n", op, op)
		}
	}
}
