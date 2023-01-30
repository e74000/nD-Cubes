package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"gonum.org/v1/gonum/mat"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"
)

type window struct {
	cube *nCube

	scale   float64
	viewPos *mat.VecDense

	dim int
	ncr int

	dir   *mat.VecDense
	rot   *mat.VecDense
	rMats []*mat.Dense

	counter time.Time
}

func (w *window) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0) // Quit if [esc] is pressed
	}

	if time.Since(w.counter).Seconds() > .5 {
		w.dir = randAxis(w.ncr, 1.0/50.0) // Pick random rotation axis
		w.counter = time.Now()            // and reset timer

		// Clean up rot by mapping to 0 <-> 2pi
		for i := 0; i < w.ncr; i++ {
			w.rot.SetVec(i, math.Mod(w.rot.AtVec(i), math.Pi*2))
		}
	} else {
		w.rot.AddVec(w.rot, w.dir)
	}

	return nil
}

func (w *window) Draw(screen *ebiten.Image) {
	rotated := rotate(w.cube.points, w.rot, w.rMats)

	for j := 0; j < 1<<w.dim; j++ {
		p1 := mat.VecDenseCopyOf(rotated.ColView(j))
		p1.SubVec(w.viewPos, p1)
		ps1 := avgPersp2d(p1).projectToScreenScale(screen, w.scale)

		for i := 0; i < 1<<w.dim; i++ {
			if w.cube.edges.At(i, j) != 0 {
				p2 := mat.VecDenseCopyOf(rotated.ColView(i))
				p2.SubVec(w.viewPos, p2)
				ps2 := avgPersp2d(p2).projectToScreenScale(screen, w.scale)
				ebitenutil.DrawLine(screen, ps1.x, ps1.y, ps2.x, ps2.y, color.White)

				// Debug mode woop woop
				//ebitenutil.DebugPrintAt(screen, vecString(p1), int(ps1.x), int(ps1.y))
				//ebitenutil.DebugPrintAt(screen, vecString(p2), int(ps2.x), int(ps2.y))
			}
		}
	}
}

func (w *window) Layout(x, y int) (screenWidth, screenHeight int) {
	return x, y
}

func (w *window) init(dim int, scale float64) {
	ncr := nCr(dim, 2)

	*w = window{
		cube:    makeNCube(dim),
		scale:   scale,
		viewPos: mat.NewVecDense(dim, nil),
		dim:     dim,
		ncr:     ncr,
		dir:     mat.NewVecDense(ncr, nil),
		rot:     mat.NewVecDense(ncr, nil),
		rMats:   givenMats(dim),
		counter: time.Now(),
	}

	w.viewPos.SetVec(2, 1+2*math.Sqrt(float64(dim)))

	calculateP2dMat(dim)
	calculateI2dMat(dim)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	w := &window{}
	w.init(9, 3)

	ebiten.SetWindowTitle("N-Dimensional shape rotator")
	ebiten.SetFullscreen(true)

	if err := ebiten.RunGame(w); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
