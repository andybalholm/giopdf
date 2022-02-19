package stroke

import (
	"math"
	"sort"

	"gioui.org/f32"
)

// simpleOffset returns a Segment that is approximately parallel to s, d units
// to the right. It just offsets the endpoints and control points; it doesn't
// subdivide the curve into smaller segments.
func simpleOffset(s Segment, d float32) Segment {
	t0, t1 := s.tangents()
	delta0 := rot90CW(t0).Mul(d)
	delta1 := rot90CW(t1).Mul(d)
	start := s.Start.Add(delta0)
	end := s.End.Add(delta1)

	// Scale the distances to the control points proportionately.
	scale := float32(1)
	originalDistance := distance(s.Start, s.End)
	if originalDistance != 0 {
		scale = distance(start, end) / originalDistance
	}

	return Segment{
		Start: start,
		CP1:   start.Add(s.CP1.Sub(s.Start).Mul(scale)),
		CP2:   end.Add(s.CP2.Sub(s.End).Mul(scale)),
		End:   end,
	}
}

func distance(a, b f32.Point) float32 {
	d := b.Sub(a)
	return float32(math.Hypot(float64(d.X), float64(d.Y)))
}

func rot90CW(p f32.Point) f32.Point { return f32.Pt(+p.Y, -p.X) }

// OffsetCurves returns the offset curves d units to the right and left of s.
func OffsetCurves(s Segment, d float32) (right, left []Segment) {
	for _, piece := range s.splitAtExtrema() {
		if simpleEnough(piece) {
			right = append(right, simpleOffset(piece, d))
			left = append(left, simpleOffset(piece, -d))
		} else {
			// Split the piece into simple-enough sections.
			const steps = 100
			pos := 0
			for pos < steps {
				next := pos + sort.Search(steps-pos, func(n int) bool {
					return !simpleEnough(piece.Split2(float32(pos)/steps, float32(pos+n+1)/steps))
				})
				if next == pos {
					next = pos + 1
				}
				seg := piece.Split2(float32(pos)/steps, float32(next)/steps)
				right = append(right, simpleOffset(seg, d))
				left = append(left, simpleOffset(seg, -d))
				pos = next
			}
		}
	}
	return right, left
}

func angle(origin, v1, v2 f32.Point) float64 {
	d1 := v1.Sub(origin)
	d2 := v2.Sub(origin)
	cross := d1.X*d2.Y - d1.Y*d2.X
	dot := d1.X*d2.X + d1.Y*d2.Y
	return math.Atan2(float64(cross), float64(dot))
}

func simpleEnough(s Segment) bool {
	a1 := angle(s.Start, s.End, s.CP1)
	a2 := angle(s.Start, s.End, s.CP2)
	if a1 > 0 && a2 < 0 || a1 < 0 && a2 > 0 {
		return false
	}
	t1, t2 := s.tangents()
	ss := t1.X*t2.X + t1.Y*t2.Y
	return math.Abs(math.Acos(float64(ss))) < math.Pi/3
}
