package main

import (
	"errors"
	"math"
	"math/rand"
)

var (
	dimensionError         = errors.New("invalid dimensionality")
	rotationParameterError = errors.New("incorrect number of rotation parameters for given dimensionality")
)

func randNNorm(n int) []float64 {
	norm := make([]float64, n)

	var mag float64

	for i := 0; i < n; i++ {
		norm[i] = rand.Float64()*2 - 1
		mag += norm[i] * norm[i]
	}

	mag = math.Sqrt(mag)

	if mag == 0 {
		return randNNorm(n)
	}

	for i := 0; i < n; i++ {
		norm[i] /= mag
	}

	return norm
}

func randNOne(n int) []float64 {
	r := make([]float64, n)

	i := rand.Int() % n

	if rand.Int()%2 == 1 {
		r[i] = -1
	} else {
		r[i] = 1
	}

	return r
}

func makeNCube(dim int) map[uint]point {
	points := make(map[uint]point)

	for i := 0; i < 1<<dim; i++ {
		pos := make(vec, dim)
		edges := make([]uint, dim)

		for j := 0; j < dim; j++ {
			pos[j] = float64(((i>>j)&1)*2 - 1)
			edges[j] = uint(i) ^ (1 << uint(dim-j-1))
		}

		points[uint(i)] = point{
			pos:   pos,
			id:    uint(i),
			edges: edges,
			value: make([]float64, dim),
		}
	}

	for i := 0; i < 1<<dim; i++ {
		p1 := points[uint(i)].pos

		for j := 0; j < dim; j++ {
			p2 := points[points[uint(i)].edges[j]].pos

			var p uint

			for k := 0; k < dim; k++ {
				p |= (uint((p1[k]+1)/2) | uint((p2[k]+1)/2)) << k
			}

			v := float64(p) / float64(int(1<<dim))

			points[uint(i)].value[j] = v
		}
	}

	return points
}

func nCr(n int, r int) int {
	num, den := 1, 1

	if r > n-r {
		r = n - r
	}

	for i := 0; i < r; i++ {
		num *= n - i
		den *= r - i
	}

	return num / den
}

func newIdent(n int) mat {
	m := make(mat, n)

	for i := 0; i < n; i++ {
		m[i] = make([]float64, n)
		m[i][i] = 1
	}

	return m
}

func givenMats(dim int) []mat {
	n := nCr(dim, 2)

	mats := make([]mat, n)

	for i := 0; i < n; i++ {
		mats[i] = newIdent(dim)
	}

	count := 0

	for i := 0; i < dim; i++ {
		for j := 0; j < dim; j++ {
			if i <= j {
				continue
			}

			mats[count][i][i] = -1
			mats[count][j][j] = -1
			mats[count][i][j] = 2
			mats[count][j][i] = -2

			count++
		}
	}

	return mats
}
