package stroke

import (
	"math"

	"gioui.org/f32"
)

// strokeContour strokes a single contour (connected series of segments). If c
// is closed, it returns both the outer contour of the stroke and the inner
// one (the outline of the hole in the middle). Otherwise inner is nil.
func strokeContour(c []Segment, opt Options) (outer, inner []Segment) {
	halfWidth := opt.Width / 2
	if !isCCW(c) {
		c = reversePath(c)
	}
	for _, s := range c {
		// Skip segments that don't do anything.
		if s.CP1 == s.Start && s.CP2 == s.Start && s.End == s.Start {
			continue
		}
		right, left := OffsetCurves(s, halfWidth)
		if len(outer) > 0 && outer[len(outer)-1].End != right[0].Start {
			j := join(outer[len(outer)-1].End, right[0].Start, s.Start, opt)
			outer = append(outer, j...)
		}
		outer = append(outer, right...)
		if len(inner) > 0 && inner[len(inner)-1].End != left[0].Start {
			j := join(left[0].Start, inner[len(inner)-1].End, s.Start, opt)
			inner = append(inner, reversePath(j)...)
		}
		inner = append(inner, left...)
	}

	if c[0].Start == c[len(c)-1].End {
		// The path was closed, so we'll return two separate contours.
		// But we need to draw a join first.
		j := join(outer[len(outer)-1].End, outer[0].Start, c[0].Start, opt)
		outer = append(outer, j...)
		j = join(inner[0].Start, inner[len(inner)-1].End, c[0].Start, opt)
		inner = append(inner, reversePath(j)...)
		return outer, reversePath(inner)
	} else {
		// Cap the ends and combine into one contour.
		switch opt.Cap {
		default:
			// FlatCap or invalid value
			outer = append(outer, LinearSegment(outer[len(outer)-1].End, inner[len(inner)-1].End))
			outer = append(outer, reversePath(inner)...)
			outer = append(outer, LinearSegment(inner[0].Start, outer[0].Start))
		case RoundCap:
			cp := roundCap(outer[len(outer)-1].End, inner[len(inner)-1].End)
			outer = append(outer, cp[:]...)
			outer = append(outer, reversePath(inner)...)
			cp = roundCap(inner[0].Start, outer[0].Start)
			outer = append(outer, cp[:]...)
		case SquareCap:
			cp := squareCap(outer[len(outer)-1].End, inner[len(inner)-1].End)
			outer = append(outer, cp[:]...)
			outer = append(outer, reversePath(inner)...)
			cp = squareCap(inner[0].Start, outer[0].Start)
			outer = append(outer, cp[:]...)
		}
		return outer, nil
	}
}

func roundCap(p1, p2 f32.Point) [2]Segment {
	const k = 0.551784777779014
	half := p2.Sub(p1).Mul(0.5)
	tip := p1.Add(half).Add(rot90CW(half))
	return [2]Segment{
		{p1, p1.Add(rot90CW(half).Mul(k)), tip.Sub(half.Mul(k)), tip},
		{tip, tip.Add(half.Mul(k)), p2.Add(rot90CW(half).Mul(k)), p2},
	}
}

func squareCap(p1, p2 f32.Point) [3]Segment {
	half := p2.Sub(p1).Mul(0.5)
	offset := rot90CW(half)
	return [3]Segment{
		LinearSegment(p1, p1.Add(offset)),
		LinearSegment(p1.Add(offset), p2.Add(offset)),
		LinearSegment(p2.Add(offset), p2),
	}
}

// Stroke returns outlines for the contours in path. Both in the parameter and
// in the return value, each element of the slice is a contour (a connected
// series of segments).
func Stroke(path [][]Segment, opt Options) [][]Segment {
	var result [][]Segment
	for _, c := range path {
		outer, inner := strokeContour(c, opt)
		result = append(result, outer)
		if inner != nil {
			result = append(result, inner)
		}
	}
	return result
}

type CapStyle int

const (
	FlatCap   CapStyle = 0
	RoundCap           = 1
	SquareCap          = 2
)

type JoinStyle int

const (
	MiterJoin JoinStyle = 0
	RoundJoin           = 1
	BevelJoin           = 2
)

type Options struct {
	Width      float32
	Cap        CapStyle
	Join       JoinStyle
	MiterLimit float32
}

// isCCW returns whether c is counter-clockwise.
func isCCW(c []Segment) bool {
	// Use the shoelace formula:
	//  https://en.wikipedia.org/wiki/Shoelace_formula
	var area float32
	for _, s := range c {
		area += (s.End.X - s.Start.X) * (s.End.Y + s.Start.Y)
	}
	return area < 0
}

// join draws a corner join from start to end, with the style specified by opt.
// (center is the center of the corner; i.e. the corner of the path being
// stroked.)
func join(start, end, center f32.Point, opt Options) []Segment {
	style := opt.Join

	angle := math.Atan2(float64(start.X-center.X), float64(start.Y-center.Y)) -
		math.Atan2(float64(end.X-center.X), float64(end.Y-center.Y))
	switch {
	case angle > math.Pi:
		angle -= 2 * math.Pi
	case angle < -math.Pi:
		angle += 2 * math.Pi
	}

	// If it's an inside corner, always do a bevel join, since it's the simplest,
	// and it won't show anyway.
	if angle < 0 {
		style = BevelJoin
	}

	if style == MiterJoin {
		phi := math.Pi - angle
		miterRatio := float32(1 / math.Sin(phi/2))
		if miterRatio > opt.MiterLimit {
			style = BevelJoin
		} else {
			direction := rot90CW(end.Sub(start).Div(distance(end, start)))
			dist := distance(start, center) * miterRatio
			tip := center.Add(direction.Mul(dist))
			return []Segment{
				LinearSegment(start, tip),
				LinearSegment(tip, end),
			}
		}
	}

	if style == RoundJoin {
		k := float32(math.Tan(angle/4)) * 4 / 3
		cp1 := start.Add(rot90CW(center.Sub(start)).Mul(k))
		cp2 := end.Add(rot90CW(end.Sub(center)).Mul(k))
		return []Segment{
			{start, cp1, cp2, end},
		}
	}

	return []Segment{
		LinearSegment(start, end),
	}
}
