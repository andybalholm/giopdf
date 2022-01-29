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
		case contentstream.OpConcat:
			m := op.Matrix
			r.Transform(m[0], m[1], m[2], m[3], m[4], m[5])
		case contentstream.OpRestore:
			r.Restore()
		case contentstream.OpSave:
			r.Save()
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
