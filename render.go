package giopdf

import (
	"fmt"

	"gioui.org/op"
	"github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/pdf/reader/parser"
)

// RenderPage draws the contents of a PDF page to ops.
// The caller should do the appropriate transformation and scaling to ensure
// that the content is not rendered upside down. (The PDF coordinate system
// starts in the lower left, not in the upper left like Gio's.)
func RenderPage(ops *op.Ops, page *model.PageObject) error {
	r := newRenderer(ops)
	r.resources = page.Resources

	for _, stream := range page.Contents {
		decoded, err := stream.Decode()
		if err != nil {
			return err
		}
		operations, err := parser.ParseContent(decoded, page.Resources.ColorSpace)
		if err != nil {
			return err
		}
		r.do(operations)
	}

	return nil
}

type renderer struct {
	*Canvas

	resources *model.ResourcesDict
}

func newRenderer(ops *op.Ops) *renderer {
	return &renderer{
		Canvas: NewCanvas(ops),
	}
}

func (r *renderer) do(operations []contentstream.Operation) {
	for _, op := range operations {
		switch op := op.(type) {
		case contentstream.OpClosePath:
			r.ClosePath()
		case contentstream.OpConcat:
			m := op.Matrix
			r.Transform(m[0], m[1], m[2], m[3], m[4], m[5])
		case contentstream.OpCubicTo:
			r.CurveTo(op.X1, op.Y1, op.X2, op.Y2, op.X3, op.Y3)
		case contentstream.OpCurveTo1:
			r.CurveV(op.X2, op.Y2, op.X3, op.Y3)
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
			gs := r.resources.ExtGState[op.Dict]
			if gs.LW != 0 {
				r.SetLineWidth(gs.LW)
			}
			if lc, ok := gs.LC.(model.ObjInt); ok {
				r.SetLineCap(int(lc))
			}
			if lj, ok := gs.LJ.(model.ObjInt); ok {
				r.SetLineJoin(int(lj))
			}
			if gs.ML != 0 {
				r.SetMiterLimit(gs.ML)
			}
			if gs.D != nil {
				r.SetDash(gs.D.Array, gs.D.Phase)
			}
			if CA, ok := gs.CA.(model.ObjFloat); ok {
				r.SetStrokeAlpha(float32(CA))
			}
			if ca, ok := gs.Ca.(model.ObjFloat); ok {
				r.SetFillAlpha(float32(ca))
			}
		case contentstream.OpSetFillGray:
			r.SetFillGray(op.G)
		case contentstream.OpSetFillRGBColor:
			r.SetRGBFillColor(op.R, op.G, op.B)
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
		case contentstream.OpStroke:
			r.Stroke()
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
		default:
			fmt.Printf("%T: %v\n", op, op)
		}
	}
}
