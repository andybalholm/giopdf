package stroke

// StrokeContour strokes a single contour (connected series of segments). If c
// is closed, it returns both the outer contour of the stroke and the inner
// one (the outline of the hole in the middle). Otherwise inner is nil.
func StrokeContour(c []Segment, width float32) (outer, inner []Segment) {
	halfWidth := width / 2
	for _, s := range c {
		right, left := OffsetCurves(s, halfWidth)
		if len(outer) > 0 && outer[len(outer)-1].End != right[0].Start {
			// TODO: other join styles
			outer = append(outer, LinearSegment(outer[len(outer)-1].End, right[0].Start))
		}
		outer = append(outer, right...)
		if len(inner) > 0 && inner[len(inner)-1].End != left[0].Start {
			inner = append(inner, LinearSegment(inner[len(inner)-1].End, left[0].Start))
		}
		inner = append(inner, left...)
	}

	if c[0].Start == c[len(c)-1].End {
		// The path was closed, so we'll return two separate contours.
		// TODO: check for countrclockwise direction.
		return outer, reversePath(inner)
	} else {
		// Cap the ends and combine into one contour.
		// TODO: other cap styles.
		outer = append(outer, LinearSegment(outer[len(outer)-1].End, inner[len(inner)-1].End))
		outer = append(outer, reversePath(inner)...)
		outer = append(outer, LinearSegment(inner[0].Start, outer[0].Start))
		return outer, nil
	}
}
