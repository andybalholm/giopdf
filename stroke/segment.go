// The stroke package provides functions for stroking cubic bezier paths.
//
// Many of the algorithms come from https://pomax.github.io/bezierinfo/
package stroke

import (
	"math"

	"gioui.org/f32"
	"golang.org/x/exp/slices"
)

// A Segment is a cubic bezier curve (or a line segment that has been converted
// into a bezier curve).
type Segment struct {
	Start    f32.Point
	CP1, CP2 f32.Point
	End      f32.Point
}

// LinearSegment returns a line segment connecting a and b, in the form of a
// cubic bezier curve with collinear control points.
func LinearSegment(a, b f32.Point) Segment {
	diff := b.Sub(a)
	spacing := diff.Div(3)
	return Segment{
		Start: a,
		CP1:   a.Add(spacing),
		CP2:   b.Sub(spacing),
		End:   b,
	}
}

// unitVector returns p scaled to that it lies on the unit circle (one unit
// away from the origin, in the same direction. If p is (0, 0), it is returned
// unchanged.
func unitVector(p f32.Point) f32.Point {
	if p == f32.Pt(0, 0) {
		return p
	}
	length := float32(math.Hypot(float64(p.X), float64(p.Y)))
	return p.Div(length)
}

// tangents returns the tangent directions at the start and end of s, as unit
// vectors (points with a magnitude of one unit).
func (s Segment) tangents() (t0, t1 f32.Point) {
	if s.CP1 != s.Start {
		t0 = unitVector(s.CP1.Sub(s.Start))
	} else if s.CP2 != s.Start {
		t0 = unitVector(s.CP2.Sub(s.Start))
	} else {
		t0 = unitVector(s.End.Sub(s.Start))
	}

	if s.CP2 != s.End {
		t1 = unitVector(s.End.Sub(s.CP2))
	} else if s.CP1 != s.End {
		t1 = unitVector(s.End.Sub(s.CP1))
	} else {
		t1 = unitVector(s.End.Sub(s.Start))
	}

	return t0, t1
}

// quadraticRoots appends the values of t for which a one-dimensional quadratic
// bezier function (with endpoints a and c, and control point b) returns zero.
func quadraticRoots(dst []float32, a, b, c float32) []float32 {
	d := a - 2*b + c
	switch {
	case d != 0:
		m1 := float32(-math.Sqrt(float64(b*b - a*c)))
		m2 := -a + b
		v1 := -(m1 + m2) / d
		v2 := -(-m1 + m2) / d
		return append(dst, v1, v2)
	case b != c && d == 0:
		return append(dst, (2*b-c)/(2*(b-c)))
	default:
		return dst
	}
}

// linearRoot returns the value of t for which a one-dimensional linear
// bezeir function (with endpoints a and b) returns zero.
func linearRoot(a, b float32) (root float32, ok bool) {
	if a != b {
		return a / (a - b), true
	}
	return 0, false
}

// extrema returns a sorted slice of t values of the extreme points of s,
// including the start and end points (t = 0 and t = 1).
func (s Segment) extrema() []float32 {
	var storage [8]float32
	result := storage[:0]
	a, b, c := s.CP1.X-s.Start.X, s.CP2.X-s.CP1.X, s.End.X-s.CP2.X
	result = quadraticRoots(result, a, b, c)
	if r, ok := linearRoot(b-a, c-b); ok {
		result = append(result, r)
	}
	a, b, c = s.CP1.Y-s.Start.Y, s.CP2.Y-s.CP1.Y, s.End.Y-s.CP2.Y
	result = quadraticRoots(result, a, b, c)
	if r, ok := linearRoot(b-a, c-b); ok {
		result = append(result, r)
	}
	// Make sure the endpoints are included.
	result = append(result, 0, 1)
	// Filter out results that are outside the range 0 to 1, or NaN.
	for i, v := range result {
		if v < 0 || v > 1 || v != v {
			result[i] = 0
		}
	}
	slices.Sort(result)
	return slices.Compact(result)
}

// interpolate returns a point between a and b, with the ratio specified by t.
func interpolate(t float32, a, b f32.Point) f32.Point {
	return a.Mul(1 - t).Add(b.Mul(t))
}

// Split splits s into two segments with de Casteljau's algorithm, at t.
func (s Segment) Split(t float32) (Segment, Segment) {
	a1 := interpolate(t, s.Start, s.CP1)
	a2 := interpolate(t, s.CP1, s.CP2)
	a3 := interpolate(t, s.CP2, s.End)

	b1 := interpolate(t, a1, a2)
	b2 := interpolate(t, a2, a3)

	c := interpolate(t, b1, b2)

	return Segment{s.Start, a1, b1, c}, Segment{c, b2, a3, s.End}
}

// Split2 returns the section of s that lies between t1 and t2.
func (s Segment) Split2(t1, t2 float32) Segment {
	if t1 == 0 {
		r, _ := s.Split(t2)
		return r
	}
	if t2 == 1 {
		_, r := s.Split(t1)
		return r
	}

	a, _ := s.Split(t2)
	_, b := a.Split(t1 / t2)

	return b
}

// splitAtExtrema returns a slice of sub-segments of s that start and end at
// the extrema (points with maximum or minumum coordinates or slope).
func (s Segment) splitAtExtrema() []Segment {
	extrema := s.extrema()
	result := make([]Segment, len(extrema)-1)
	for i := range result {
		result[i] = s.Split2(extrema[i], extrema[i+1])
	}
	return result
}

func (s Segment) reverse() Segment {
	return Segment{s.End, s.CP2, s.CP1, s.Start}
}

func reversePath(path []Segment) []Segment {
	result := make([]Segment, len(path))
	for i, s := range path {
		result[len(result)-i-1] = s.reverse()
	}
	return result
}
