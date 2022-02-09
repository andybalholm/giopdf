package pdf

import "io"

// A ContentStream is a sequence of instructions (operators and operands)
// describing the content of a page.
type ContentStream struct {
	b *buffer
}

// NewContentStream creates a ContentStream with the contents of r.
func NewContentStream(r io.Reader) *ContentStream {
	b := newBuffer(r, 0)
	b.allowEOF = true
	b.allowObjptr = false
	b.allowStream = false

	return &ContentStream{b: b}
}

// ReadInstruction returns the next instruction.
// If the end of the stream is reached, the operator will be the empty string.
func (cs *ContentStream) ReadInstruction() (operands []Value, operator string) {
	for {
		tok := cs.b.readToken()
		if tok == io.EOF {
			return operands, ""
		}
		if kw, ok := tok.(keyword); ok {
			switch kw {
			case "null", "[", "]", "<<", ">>":
				break
			default:
				return operands, string(kw)
			}
		}
		cs.b.unreadToken(tok)
		obj := cs.b.readObject()
		operands = append(operands, Value{nil, objptr{}, obj})
	}
}
