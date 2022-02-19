package stroke

import (
	"testing"

	"gioui.org/f32"
	"golang.org/x/exp/slices"
)

var tangentTests = []struct {
	segment Segment
	t0, t1  f32.Point
}{
	{
		segment: Segment{
			Start: f32.Pt(119, 100),
			CP1:   f32.Pt(25, 190),
			CP2:   f32.Pt(210, 250),
			End:   f32.Pt(210, 30),
		},
		t0: f32.Pt(-0.72230804, 0.69157153),
		t1: f32.Pt(0, -1),
	},
	{
		segment: Segment{
			Start: f32.Pt(25, 190),
			CP1:   f32.Pt(25, 190),
			CP2:   f32.Pt(210, 250),
			End:   f32.Pt(210, 30),
		},
		t0: f32.Pt(0.95122284, 0.3085047),
		t1: f32.Pt(0, -1),
	},
}

func TestTangents(t *testing.T) {
	for _, c := range tangentTests {
		t0, t1 := c.segment.tangents()
		if t0 != c.t0 {
			t.Errorf("unexpected t0 for %v: got %v, want %v", c.segment, t0, c.t0)
		}
		if t1 != c.t1 {
			t.Errorf("unexpected t1 for %v: got %v, want %v", c.segment, t1, c.t1)
		}
	}
}

var extremaTests = []struct {
	segment Segment
	extrema []float32
}{
	{
		segment: Segment{
			Start: f32.Pt(110, 150),
			CP1:   f32.Pt(25, 190),
			CP2:   f32.Pt(210, 250),
			End:   f32.Pt(210, 30),
		},
		extrema: []float32{0, 0.06666667, 0.18681319, 0.43785095, 0.5934066, 1},
	},
}

func TestExtrema(t *testing.T) {
	for _, c := range extremaTests {
		extrema := c.segment.extrema()
		if !slices.Equal(extrema, c.extrema) {
			t.Errorf("extrema for %v: got %v, want %v", c.segment, extrema, c.extrema)
		}
	}
}
