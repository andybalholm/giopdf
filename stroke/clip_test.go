// SPDX-License-Identifier: Unlicense OR MIT

// Adapted from gioui.org/x/stroke

package stroke

import (
	"math"
	"testing"

	"golang.org/x/image/colornames"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func TestStrokedPathBevelFlat(t *testing.T) {
	run(t, func(o *op.Ops) {
		s := Stroke(fruit, Options{
			Width: 2.5,
			Cap:   FlatCap,
			Join:  BevelJoin,
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(10, 50, colornames.Red)
	})
}

func TestStrokedPathBevelRound(t *testing.T) {
	run(t, func(o *op.Ops) {
		s := Stroke(fruit, Options{
			Width: 2.5,
			Cap:   RoundCap,
			Join:  BevelJoin,
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(10, 50, colornames.Red)
	})
}

func TestStrokedPathBevelSquare(t *testing.T) {
	run(t, func(o *op.Ops) {
		s := Stroke(fruit, Options{
			Width: 2.5,
			Cap:   SquareCap,
			Join:  BevelJoin,
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(10, 50, colornames.Red)
	})
}

func TestStrokedPathFlatMiter(t *testing.T) {
	run(t, func(o *op.Ops) {
		s := Stroke(zigzag, Options{
			Width:      10,
			Cap:        FlatCap,
			Join:       MiterJoin,
			MiterLimit: 5,
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, s)}.Op())

		s = Stroke(zigzag, Options{
			Width: 2,
			Cap:   FlatCap,
			Join:  BevelJoin,
		})
		paint.FillShape(o, black, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(40, 10, colornames.Black)
		r.expect(40, 12, colornames.Red)
	})
}

func TestStrokedPathFlatMiterInf(t *testing.T) {
	run(t, func(o *op.Ops) {
		s := Stroke(zigzag, Options{
			Width:      10,
			Cap:        FlatCap,
			Join:       MiterJoin,
			MiterLimit: float32(math.Inf(+1)),
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, s)}.Op())

		s = Stroke(zigzag, Options{
			Width: 2,
			Cap:   FlatCap,
			Join:  BevelJoin,
		})
		paint.FillShape(o, black, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(40, 10, colornames.Black)
		r.expect(40, 12, colornames.Red)
	})
}

func TestStrokedPathZeroWidth(t *testing.T) {
	run(t, func(o *op.Ops) {
		p := [][]Segment{
			{
				LinearSegment(f32.Pt(10, 50), f32.Pt(60, 50)),
			},
		}
		st := Stroke(p, Options{
			Width: 2,
			Cap:   FlatCap,
			Join:  BevelJoin,
		})
		paint.FillShape(o, black, clip.Outline{ToPathSpec(o, st)}.Op())

		p = [][]Segment{
			{
				LinearSegment(f32.Pt(10, 50), f32.Pt(40, 50)),
				LinearSegment(f32.Pt(40, 50), f32.Pt(10, 50)),
			},
		}
		st = Stroke(p, Options{
			Width: 0,
			Cap:   RoundCap,
			Join:  RoundJoin,
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, st)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(10, 50, colornames.Black)
		r.expect(30, 50, colornames.Black)
		r.expect(65, 50, transparent)
	})
}

func TestDashedPathFlatCapEllipse(t *testing.T) {
	run(t, func(o *op.Ops) {
		dashed := Dash(ellipse, []float32{5, 3}, 0)
		s := Stroke(dashed, Options{
			Width:      10,
			Cap:        FlatCap,
			Join:       MiterJoin,
			MiterLimit: float32(math.Inf(+1)),
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, s)}.Op())

		s = Stroke(ellipse, Options{
			Width: 2,
		})
		paint.FillShape(o, black, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(0, 62, colornames.Red)
		r.expect(0, 65, colornames.Black)
	})
}

func TestDashedPathFlatCapZ(t *testing.T) {
	run(t, func(o *op.Ops) {
		dashed := Dash(zigzag, []float32{5, 3}, 0)
		s := Stroke(dashed, Options{
			Width:      10,
			Cap:        FlatCap,
			Join:       MiterJoin,
			MiterLimit: float32(math.Inf(+1)),
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, s)}.Op())

		s = Stroke(zigzag, Options{
			Width: 2,
			Cap:   FlatCap,
			Join:  BevelJoin,
		})
		paint.FillShape(o, black, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(40, 10, colornames.Black)
		r.expect(40, 12, colornames.Red)
		r.expect(46, 12, transparent)
	})
}

func TestDashedPathFlatCapZNoDash(t *testing.T) {
	run(t, func(o *op.Ops) {
		dashed := Dash(zigzag, nil, 1)
		s := Stroke(dashed, Options{
			Width:      10,
			Cap:        FlatCap,
			Join:       MiterJoin,
			MiterLimit: float32(math.Inf(+1)),
		})
		paint.FillShape(o, red, clip.Outline{ToPathSpec(o, s)}.Op())

		s = Stroke(zigzag, Options{
			Width: 2,
			Cap:   FlatCap,
			Join:  BevelJoin,
		})
		paint.FillShape(o, black, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(40, 10, colornames.Black)
		r.expect(40, 12, colornames.Red)
		r.expect(46, 12, colornames.Red)
	})
}

func TestStrokedPathCoincidentControlPoint(t *testing.T) {
	run(t, func(o *op.Ops) {
		p := [][]Segment{
			{
				Segment{f32.Pt(70, 20), f32.Pt(70, 20), f32.Pt(70, 110), f32.Pt(120, 120)},
				LinearSegment(f32.Pt(120, 120), f32.Pt(20, 120)),
				LinearSegment(f32.Pt(20, 120), f32.Pt(70, 20)),
			},
		}
		s := Stroke(p, Options{
			Width: 20,
			Cap:   RoundCap,
			Join:  RoundJoin,
		})
		paint.FillShape(o, black, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(70, 20, colornames.Black)
		r.expect(70, 90, transparent)
	})
}

func TestStrokedPathBalloon(t *testing.T) {
	run(t, func(o *op.Ops) {
		// This shape is based on the one drawn by the Bubble function in
		// github.com/llgcode/draw2d/samples/geometry/geometry.go.
		p := [][]Segment{
			{
				Segment{f32.Pt(42.69375, 10.5), f32.Pt(42.69375, 10.5), f32.Pt(14.85, 10.5), f32.Pt(14.85, 31.5)},
				Segment{f32.Pt(14.85, 31.5), f32.Pt(14.85, 31.5), f32.Pt(14.85, 52.5), f32.Pt(28.771875, 52.5)},
				Segment{f32.Pt(28.771875, 52.5), f32.Pt(28.771875, 52.5), f32.Pt(28.771875, 63.7), f32.Pt(17.634375, 66.5)},
				Segment{f32.Pt(17.634375, 66.5), f32.Pt(17.634375, 66.5), f32.Pt(34.340626, 63.7), f32.Pt(37.125, 52.5)},
				Segment{f32.Pt(37.125, 52.5), f32.Pt(37.125, 52.5), f32.Pt(70.5375, 52.5), f32.Pt(70.5375, 31.5)},
				Segment{f32.Pt(70.5375, 31.5), f32.Pt(70.5375, 31.5), f32.Pt(70.5375, 10.5), f32.Pt(42.69375, 10.5)},
			},
		}
		s := Stroke(p, Options{
			Width: 2.83,
			Cap:   RoundCap,
			Join:  RoundJoin,
		})
		paint.FillShape(o, black, clip.Outline{ToPathSpec(o, s)}.Op())
	}, func(r result) {
		r.expect(0, 0, transparent)
		r.expect(70, 52, colornames.Black)
		r.expect(70, 90, transparent)
	})
}

var fruit = [][]Segment{
	{
		LinearSegment(f32.Pt(10, 50), f32.Pt(20, 50)),
		QuadraticSegment(f32.Pt(20, 50), f32.Point{X: 20.00035, Y: 48.607147}, f32.Point{X: 20.288229, Y: 47.240997}),
		QuadraticSegment(f32.Point{X: 20.288229, Y: 47.240997}, f32.Point{X: 20.57679, Y: 45.874977}, f32.Point{X: 21.141825, Y: 44.588024}),
		QuadraticSegment(f32.Point{X: 21.141825, Y: 44.588024}, f32.Point{X: 21.707504, Y: 43.301327}, f32.Point{X: 22.527983, Y: 42.143032}),
		QuadraticSegment(f32.Point{X: 22.527983, Y: 42.143032}, f32.Point{X: 23.349041, Y: 40.985104}, f32.Point{X: 24.393435, Y: 39.99998}),
		QuadraticSegment(f32.Point{X: 24.393435, Y: 39.99998}, f32.Point{X: 25.43832, Y: 39.01532}, f32.Point{X: 26.666492, Y: 38.241226}),
		QuadraticSegment(f32.Point{X: 26.666492, Y: 38.241226}, f32.Point{X: 27.89505, Y: 37.467674}, f32.Point{X: 29.259802, Y: 36.934353}),
		QuadraticSegment(f32.Point{X: 29.259802, Y: 36.934353}, f32.Point{X: 30.62482, Y: 36.401638}, f32.Point{X: 32.073708, Y: 36.12959}),
		QuadraticSegment(f32.Point{X: 32.073708, Y: 36.12959}, f32.Point{X: 33.522728, Y: 35.858185}, f32.Point{X: 35.00007, Y: 35.857857}),
		QuadraticSegment(f32.Point{X: 35.00007, Y: 35.857857}, f32.Point{X: 36.47741, Y: 35.858192}, f32.Point{X: 37.92643, Y: 36.129604}),
		QuadraticSegment(f32.Point{X: 37.92643, Y: 36.129604}, f32.Point{X: 39.375313, Y: 36.401665}, f32.Point{X: 40.740334, Y: 36.93438}),
		QuadraticSegment(f32.Point{X: 40.740334, Y: 36.93438}, f32.Point{X: 42.105087, Y: 37.467705}, f32.Point{X: 43.33364, Y: 38.241264}),
		QuadraticSegment(f32.Point{X: 43.33364, Y: 38.241264}, f32.Point{X: 44.561806, Y: 39.01536}, f32.Point{X: 45.60668, Y: 40.00003}),
		QuadraticSegment(f32.Point{X: 45.60668, Y: 40.00003}, f32.Point{X: 46.651073, Y: 40.985157}, f32.Point{X: 47.472122, Y: 42.14309}),
		QuadraticSegment(f32.Point{X: 47.472122, Y: 42.14309}, f32.Point{X: 48.292587, Y: 43.301384}, f32.Point{X: 48.85826, Y: 44.58808}),
		QuadraticSegment(f32.Point{X: 48.85826, Y: 44.58808}, f32.Point{X: 49.42329, Y: 45.87504}, f32.Point{X: 49.711845, Y: 47.241055}),
		QuadraticSegment(f32.Point{X: 49.711845, Y: 47.241055}, f32.Point{X: 49.999718, Y: 48.60721}, f32.Point{X: 50.000053, Y: 50.000053}),
		LinearSegment(f32.Pt(50.000053, 50.00053), f32.Pt(60, 50)),
		LinearSegment(f32.Pt(60, 50), f32.Pt(70, 60)),
		QuadraticSegment(f32.Pt(70, 60), f32.Point{X: 75.96515, Y: 60.01108}, f32.Point{X: 81.48046, Y: 62.283623}),
		QuadraticSegment(f32.Point{X: 81.48046, Y: 62.283623}, f32.Point{X: 86.987305, Y: 64.57663}, f32.Point{X: 91.21312, Y: 68.78679}),
		QuadraticSegment(f32.Point{X: 91.21312, Y: 68.78679}, f32.Point{X: 95.423294, Y: 73.01262}, f32.Point{X: 97.71627, Y: 78.519455}),
		QuadraticSegment(f32.Point{X: 97.71627, Y: 78.519455}, f32.Point{X: 99.98879, Y: 84.034775}, f32.Point{X: 99.99987, Y: 89.999916}),
		QuadraticSegment(f32.Point{X: 99.99987, Y: 89.999916}, f32.Point{X: 99.988785, Y: 95.96506}, f32.Point{X: 97.716255, Y: 101.48037}),
		QuadraticSegment(f32.Point{X: 97.716255, Y: 101.48037}, f32.Point{X: 95.42325, Y: 106.9872}, f32.Point{X: 91.21309, Y: 111.21302}),
		QuadraticSegment(f32.Point{X: 91.21309, Y: 111.21302}, f32.Point{X: 86.987274, Y: 115.42317}, f32.Point{X: 81.48043, Y: 117.71617}),
		QuadraticSegment(f32.Point{X: 81.48043, Y: 117.71617}, f32.Point{X: 75.96512, Y: 119.9887}, f32.Point{X: 69.99997, Y: 119.99979}),
		QuadraticSegment(f32.Point{X: 69.99997, Y: 119.99979}, f32.Point{X: 64.03482, Y: 119.9887}, f32.Point{X: 58.51951, Y: 117.71617}),
		QuadraticSegment(f32.Point{X: 58.51951, Y: 117.71617}, f32.Point{X: 53.01267, Y: 115.42317}, f32.Point{X: 48.78685, Y: 111.213005}),
		QuadraticSegment(f32.Point{X: 48.78685, Y: 111.213005}, f32.Point{X: 44.57669, Y: 106.98717}, f32.Point{X: 42.283707, Y: 101.48033}),
		QuadraticSegment(f32.Point{X: 42.283707, Y: 101.48033}, f32.Point{X: 40.011185, Y: 95.96501}, f32.Point{X: 40.000122, Y: 89.99987}),
		QuadraticSegment(f32.Point{X: 40.000122, Y: 89.99987}, f32.Point{X: 40.01121, Y: 84.03473}, f32.Point{X: 42.283745, Y: 78.519424}),
		QuadraticSegment(f32.Point{X: 42.283745, Y: 78.519424}, f32.Point{X: 44.576748, Y: 73.01259}, f32.Point{X: 48.78691, Y: 68.78678}),
		QuadraticSegment(f32.Point{X: 48.78691, Y: 68.78678}, f32.Point{X: 53.012737, Y: 64.57663}, f32.Point{X: 58.51957, Y: 62.283646}),
		QuadraticSegment(f32.Point{X: 58.51957, Y: 62.283646}, f32.Point{X: 64.03488, Y: 60.01113}, f32.Point{X: 70.000015, Y: 60.00006}),
		LinearSegment(f32.Pt(70.000015, 60.00006), f32.Pt(50, 60)),
		QuadraticSegment(f32.Pt(50, 60), f32.Pt(40, 50), f32.Pt(20, 90)),
	},
}

var zigzag = [][]Segment{
	{
		LinearSegment(f32.Pt(40, 10), f32.Pt(90, 10)),
		LinearSegment(f32.Pt(90, 10), f32.Pt(40, 60)),
		LinearSegment(f32.Pt(40, 60), f32.Pt(90, 60)),
		QuadraticSegment(f32.Pt(90, 60), f32.Pt(40, 80), f32.Pt(40, 110)),
		LinearSegment(f32.Pt(40, 110), f32.Pt(90, 110)),
	},
}

var ellipse = [][]Segment{
	{
		LinearSegment(f32.Pt(0, 65), f32.Pt(20, 65)),
		QuadraticSegment(f32.Pt(20, 65), f32.Point{X: 20.016611, Y: 57.560127}, f32.Point{X: 23.425419, Y: 50.681286}),
		QuadraticSegment(f32.Point{X: 23.425419, Y: 50.681286}, f32.Point{X: 26.864927, Y: 43.81302}, f32.Point{X: 33.1802, Y: 38.542465}),
		QuadraticSegment(f32.Point{X: 33.1802, Y: 38.542465}, f32.Point{X: 39.51897, Y: 33.291443}, f32.Point{X: 47.779266, Y: 30.431564}),
		QuadraticSegment(f32.Point{X: 47.779266, Y: 30.431564}, f32.Point{X: 56.052277, Y: 27.59721}, f32.Point{X: 65.00003, Y: 27.583397}),
		QuadraticSegment(f32.Point{X: 65.00003, Y: 27.583397}, f32.Point{X: 73.947784, Y: 27.59721}, f32.Point{X: 82.2208, Y: 30.431564}),
		QuadraticSegment(f32.Point{X: 82.2208, Y: 30.431564}, f32.Point{X: 90.4811, Y: 33.291443}, f32.Point{X: 96.81986, Y: 38.542465}),
		QuadraticSegment(f32.Point{X: 96.81986, Y: 38.542465}, f32.Point{X: 103.13513, Y: 43.813015}, f32.Point{X: 106.574646, Y: 50.681282}),
		QuadraticSegment(f32.Point{X: 106.574646, Y: 50.681282}, f32.Point{X: 109.98345, Y: 57.56012}, f32.Point{X: 110.00008, Y: 64.99999}),
		QuadraticSegment(f32.Point{X: 110.00008, Y: 64.99999}, f32.Point{X: 109.98346, Y: 72.439865}, f32.Point{X: 106.57466, Y: 79.3187}),
		QuadraticSegment(f32.Point{X: 106.57466, Y: 79.3187}, f32.Point{X: 103.135155, Y: 86.18697}, f32.Point{X: 96.819885, Y: 91.45753}),
		QuadraticSegment(f32.Point{X: 96.819885, Y: 91.45753}, f32.Point{X: 90.48111, Y: 96.70854}, f32.Point{X: 82.22082, Y: 99.568436}),
		QuadraticSegment(f32.Point{X: 82.22082, Y: 99.568436}, f32.Point{X: 73.9478, Y: 102.40279}, f32.Point{X: 65.000046, Y: 102.4166}),
		QuadraticSegment(f32.Point{X: 65.000046, Y: 102.4166}, f32.Point{X: 56.052288, Y: 102.40279}, f32.Point{X: 47.779274, Y: 99.568436}),
		QuadraticSegment(f32.Point{X: 47.779274, Y: 99.568436}, f32.Point{X: 39.51898, Y: 96.70856}, f32.Point{X: 33.180206, Y: 91.45754}),
		QuadraticSegment(f32.Point{X: 33.180206, Y: 91.45754}, f32.Point{X: 26.86493, Y: 86.18698}, f32.Point{X: 23.425415, Y: 79.318726}),
		QuadraticSegment(f32.Point{X: 23.425415, Y: 79.318726}, f32.Point{X: 20.016602, Y: 72.439896}, f32.Point{X: 19.999983, Y: 65.00001}),
	},
}
