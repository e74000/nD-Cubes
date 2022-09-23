package main

import (
	"github.com/hajimehoshi/ebiten"
	"math"
)

func (v vec) project3DTo2D() vec2 {
	if len(v) != 3 {
		panic(dimensionError)
	}

	return vec2{
		x: v[0] / v[2],
		y: v[1] / v[2],
	}
}

type vec []float64

func (v vec) projectNDTo3D() vec {
	switch len(v) {
	case 0:
		return vec{ // Since 0x0 matrices cannot exist, return (0,0,0)
			0, 0, 0,
		}
	case 1:
		return mat{{1}, {0}, {0}}.prod(v)
	case 2:
		return mat{{1, 0}, {0, 1}, {0, 0}}.prod(v)
	case 3:
		return mat{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}.prod(v)
	case 4:
		return mat{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}}.prod(v)
	case 5:
		return mat{{1, 0, 0, 0, 0}, {0, 1, 0, 0, 0}, {0, 0, 0.5, 0.5, 0}}.prod(v)
	case 6:
		return mat{{0.5, 0, 0, 0.5, 0, 0}, {0, 0.5, 0, 0, 0.5, 0}, {0, 0, 0.5, 0, 0, 0.5}}.prod(v)
	case 7:
		return mat{{0.5, 0, 0, 0.5, 0, 0, 0}, {0, 0.5, 0, 0, 0.5, 0, 0}, {0, 0, 0.5, 0, 0, 0.5, 0}}.prod(v)
	}

	return vec{ // Return truncated projection if matrix doesn't exist
		v[0],
		v[1],
		v[2],
	}
}

func (v vec) rotate(thetas []float64, givens []mat) vec {
	if len(thetas) != len(givens) {
		panic(rotationParameterError)
	}

	w := make(vec, len(v))
	copy(w, v)

	for i := 0; i < len(thetas); i++ {
		rMat := givens[i].applyRotation(thetas[i], len(v))
		w = rMat.prod(w)
	}

	return w
}

func (v vec) sub(w vec) vec {
	if len(v) != len(w) {
		panic(dimensionError)
	}

	r := make(vec, len(v))

	for i := 0; i < len(v); i++ {
		r[i] = v[i] - w[i]
	}

	return r
}

func (v vec) dp(w vec) float64 {
	if len(v) != len(w) {
		panic(dimensionError)
	}

	r := 0.0

	for i := 0; i < len(v); i++ {
		r += v[i] * w[i]
	}

	return r
}

type vec2 struct {
	x float64
	y float64
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

type mat [][]float64

func (m mat) prod(v vec) vec {
	if len(m[0]) != len(v) {
		panic(dimensionError)
	}

	r := make(vec, len(m))

	for i := 0; i < len(m); i++ {
		r[i] = vec(m[i]).dp(v)
	}

	return r
}

func (m mat) applyRotation(theta float64, dim int) mat {
	res := newIdent(dim)

	for i := 0; i < dim; i++ {
		for j := 0; j < dim; j++ {
			switch m[i][j] {
			case -1:
				res[i][j] = math.Cos(theta)
			case 2:
				res[i][j] = math.Sin(theta)
			case -2:
				res[i][j] = -math.Sin(theta)
			}
		}
	}

	return res
}
