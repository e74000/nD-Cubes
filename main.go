package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/mazznoer/colorgrad"
	"math"
	"math/rand"
	"os"
	"time"
)

var (
	cg = colorgrad.PuBuGn()
)

type point struct {
	pos   vec
	id    uint
	edges []uint
	value []float64
}

type window struct {
	scale   float64
	viewPos vec
	points  map[uint]point
	dim     int
	ncr     int
	rot     []float64
	rMats   []mat
	counter time.Time
	dir     []float64
}

func (g *window) Update(_ *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0) // Quit if [esc] is pressed
	}

	if time.Since(g.counter).Seconds() > 5 {
		g.dir = randNOne(g.ncr) // Pick random rotation axis
		g.counter = time.Now()  // and reset timer
	} else {
		for i := 0; i < g.ncr; i++ {
			g.rot[i] += g.dir[i] / 50                // Add to rotation
			g.rot[i] = math.Mod(g.rot[i], math.Pi*2) // Map Rotation between 0-2π
		}
	}

	return nil
}

func (g *window) Draw(screen *ebiten.Image) {
	for _, p := range g.points {
		for i, edge := range p.edges {
			p0 := p.pos.
				rotate(g.rot, g.rMats).
				sub(g.viewPos).
				projectNDTo3D().project3DTo2D().
				projectToScreenScale(screen, g.scale)
			p1 := g.points[edge].pos.
				rotate(g.rot, g.rMats).
				sub(g.viewPos).
				projectNDTo3D().project3DTo2D().
				projectToScreenScale(screen, g.scale)

			ebitenutil.DrawLine(screen, p0.x, p0.y, p1.x, p1.y, cg.At(p.value[i]))
		}
	}
}

func (g *window) Layout(x, y int) (screenWidth, screenHeight int) {
	return x, y
}

func main() {
	rand.Seed(time.Now().UnixNano())

	g := &window{
		scale:   2,
		counter: time.Now(),
		dim:     7,
	}

	g.ncr = nCr(g.dim, 2)

	g.viewPos = make(vec, g.dim)
	g.viewPos[2] = -5

	g.points = makeNCube(g.dim)

	g.rot = make([]float64, g.ncr)
	g.dir = make([]float64, g.ncr)

	g.rMats = givenMats(g.dim)

	ebiten.SetWindowTitle("N-Dimensional shape rotator")
	ebiten.SetFullscreen(true)

	if err := ebiten.RunGame(g); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
