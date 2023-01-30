package main

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"math"
	"math/bits"
	"math/rand"
)

func randUnit(n int, l float64) *mat.VecDense {
	norm := mat.NewVecDense(n, nil)

	for i := 0; i < n; i++ {
		norm.SetVec(i, rand.Float64())
	}

	norm.ScaleVec(l/math.Sqrt(mat.Dot(norm, norm)), norm)

	return norm
}

func randAxis(n int, l float64) *mat.VecDense {
	norm := mat.NewVecDense(n, nil)

	if rand.Int()%2 == 0 {
		norm.SetVec(rand.Int()%n, -l)
	} else {
		norm.SetVec(rand.Int()%n, l)
	}

	return norm
}

type nCube struct {
	points *mat.Dense
	edges  *mat.Dense
}

func makeNCube(dim int) *nCube {
	c := &nCube{
		points: mat.NewDense(dim, 1<<dim, nil),
		edges:  mat.NewDense(1<<dim, 1<<dim, nil),
	}

	c.points.Apply(func(i, j int, v float64) float64 {
		return float64(((j>>i)&1)*2 - 1)
	}, c.points)

	c.edges.Apply(func(i, j int, v float64) float64 {
		if bits.OnesCount(uint(i^j)) == 1 && i > j {
			return 1
		}

		return 0
	}, c.edges)

	return c
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

func newIdent(n int) *mat.Dense {
	m := mat.NewDense(n, n, nil)

	m.Apply(func(i, j int, v float64) float64 {
		if i == j {
			return 1
		} else {
			return 0
		}
	}, m)

	return m
}

func givenMats(dim int) []*mat.Dense {
	n := nCr(dim, 2)

	mats := make([]*mat.Dense, n)

	for i := 0; i < n; i++ {
		mats[i] = newIdent(dim)
	}

	count := 0

	for i := 0; i < dim; i++ {
		for j := 0; j < dim; j++ {
			if i <= j {
				continue
			}

			mats[count].Set(i, i, -1)
			mats[count].Set(j, j, -1)
			mats[count].Set(i, j, 2)
			mats[count].Set(j, i, -2)

			count++
		}
	}

	return mats
}

func vecString(v *mat.VecDense) string {
	s := "("
	r, _ := v.Dims()

	for i := 0; i < r; i++ {
		s += fmt.Sprintf("%.2f", v.AtVec(i))
		if i != r-1 {
			s += ", "
		}
	}

	s += ")"
	return s
}
