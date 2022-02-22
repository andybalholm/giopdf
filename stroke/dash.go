package stroke

import (
	"sort"
)

// Dash returns a dashed version of path, according to the pattern and phase
// specified.
func Dash(path [][]Segment, pattern []float32, phase float32) [][]Segment {
	var patternLen float32
	for _, d := range pattern {
		if d < 0 {
			// Invalid pattern; just return the original path.
			return path
		}
		patternLen += d
	}
	if patternLen == 0 {
		return path
	}

	for phase < 0 {
		// Multiply by two in case the pattern has an odd number of elements.
		phase += patternLen * 2
	}

	var result [][]Segment

	for _, contour := range path {
		ph := phase
		for i := 0; len(contour) > 0; i++ {
			dashLen := pattern[i%len(pattern)]
			if ph > dashLen {
				ph -= dashLen
				continue
			}
			dashLen -= ph
			ph = 0
			c1, c2 := splitContour(contour, dashLen)
			if i%2 == 0 {
				result = append(result, c1)
			}
			contour = c2
		}
	}

	return result
}

// length returns the approximate arc length of s, calculated with Gauss-Lobatto
// quadrature.
//
// It is based on the approximateCubicArcLengthC function from
// github.com/fonttools/fonttools.
func (s Segment) length() float32 {
	v0 := distance(s.Start, s.CP1) * 0.15
	v1 := hypot(
		s.Start.Mul(-0.558983582205757).
			Add(s.CP1.Mul(0.325650248872424)).
			Add(s.CP2.Mul(0.208983582205757)).
			Add(s.End.Mul(0.024349751127576)),
	)
	v2 := hypot(s.End.Sub(s.Start).Add(s.CP2).Sub(s.CP1)) * 0.26666666666666666
	v3 := hypot(
		s.Start.Mul(-0.024349751127576).
			Sub(s.CP1.Mul(0.208983582205757)).
			Sub(s.CP2.Mul(0.325650248872424)).
			Add(s.End.Mul(0.558983582205757)),
	)
	v4 := distance(s.End, s.CP2) * 0.15

	return v0 + v1 + v2 + v3 + v4
}

// splitAtLength splits s into two sections. The first one (s1) has the
// specified length, if possible, and s2 contains the remainder. If s is too
// short, s1 will be equal to s, and s2 will be empty.
func (s Segment) splitAtLength(length float32) (s1, s2 Segment) {
	const steps = 1 << 20
	n := sort.Search(steps, func(i int) bool {
		s1, _ := s.Split(float32(i) / steps)
		return s1.length() >= length
	})
	if n == steps {
		return s, Segment{}
	}
	return s.Split(float32(n) / steps)
}

// splitContour splits c into two sections. The first one (c1) has the
// specified length, if possible, and c2 contains the remainder. If c is too
// short, c1 will be equal to c, and c2 will be empty.
func splitContour(c []Segment, length float32) (c1, c2 []Segment) {
	for len(c) > 0 && length > 0 {
		segmentLength := c[0].length()
		if segmentLength > length {
			s1, s2 := c[0].splitAtLength(length)
			c1 = append(c1, s1)
			if s2 != (Segment{}) {
				c2 = []Segment{s2}
			}
			c2 = append(c2, c[1:]...)
			return c1, c2
		}
		length -= segmentLength
		c1 = append(c1, c[0])
		c = c[1:]
	}
	return c1, c
}
