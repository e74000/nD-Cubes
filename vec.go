package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"gonum.org/v1/gonum/mat"
	"math"
)

var (
	p2dMat *mat.Dense
	i2dMat *mat.Dense
)

func calculateP2dMat(n int) {
	p2dMat = mat.NewDense(3, n, nil)

	p2dMat.Apply(func(i, j int, v float64) float64 {
		if i == j%3 {
			rowNum := float64(n / 3)
			if n%3 > i%3 {
				rowNum++
			}
			return 1 / rowNum
		}
		return 0
	}, p2dMat)
}

func calculateI2dMat(n int) {
	i2dMat = mat.NewDense(2, n, nil)

	i2dMat.Apply(func(i, j int, v float64) float64 {
		theta := 2 * math.Pi * float64(j) / float64(n)
		if i == 0 {
			return math.Cos(theta)
		} else {
			return math.Sin(theta)
		}
	}, i2dMat)
}

func avgPersp2d(v *mat.VecDense) vec2 {
	temp := mat.NewDense(3, 1, nil)
	temp.Product(p2dMat, v)

	return vec2{
		x: temp.At(0, 0) / temp.At(2, 0),
		y: temp.At(1, 0) / temp.At(2, 0),
	}
}

func avgFlat2d(v *mat.VecDense) vec2 {
	temp := mat.NewDense(3, 1, nil)
	temp.Product(p2dMat, v)

	return vec2{
		x: temp.At(0, 0),
		y: temp.At(1, 0),
	}
}

func flatPersp2d(v *mat.VecDense) vec2 {
	return vec2{
		x: v.At(0, 0) / v.At(2, 0),
		y: v.At(1, 0) / v.At(2, 0),
	}
}

func flatFlat2d(v *mat.VecDense) vec2 {
	return vec2{
		x: v.At(0, 0),
		y: v.At(1, 0),
	}
}

func iso2d(v *mat.VecDense) vec2 {
	temp := mat.NewDense(2, 1, nil)
	temp.Product(i2dMat, v)

	return vec2{
		x: temp.At(0, 0),
		y: temp.At(1, 0),
	}
}

func prune2d(v *mat.VecDense) vec2 {
	return vec2{
		x: v.At(0, 0),
		y: v.At(0, 1),
	}
}

func rotate(points *mat.Dense, thetas *mat.VecDense, givens []*mat.Dense) *mat.Dense {
	r, _ := points.Dims()

	w := newIdent(r)

	for i := 0; i < len(givens); i++ {
		w.Product(w, applyRotation(givens[i], thetas.At(i, 0)))
	}

	out := mat.DenseCopyOf(points)

	out.Product(w, points)

	return out
}

type vec2 struct {
	x float64
	y float64
}

func (v vec2) MulScl(f float64) vec2 {
	return vec2{
		x: v.x * f,
		y: v.y * f,
	}
}

func (v vec2) AddVec(w vec2) vec2 {
	return vec2{
		x: v.x + w.x,
		y: v.y + w.y,
	}
}

func (v vec2) projectToScreenScale(screen *ebiten.Image, scale float64) vec2 {
	rx, ry := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())

	if rx <= ry {
		return vec2{
			x: (rx/scale)*v.x + rx/2,
			y: (rx/scale)*-v.y + ry/2,
		}
	} else {
		return vec2{
			x: (ry/scale)*v.x + rx/2,
			y: (ry/scale)*-v.y + ry/2,
		}
	}
}

func (v vec2) String() string {
	return fmt.Sprintf("(%.2f, %.2f)", v.x, v.y)
}

func applyRotation(m *mat.Dense, theta float64) *mat.Dense {
	r, c := m.Dims()
	if r != c {
		panic("Matrix must be square.")
	}

	res := newIdent(r)

	res.Apply(func(i, j int, v float64) float64 {
		switch m.At(i, j) {
		case -1:
			return math.Cos(theta)
		case 2:
			return math.Sin(theta)
		case -2:
			return -math.Sin(theta)
		default:
			return v
		}
	}, res)

	return res
}
