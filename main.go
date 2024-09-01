package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"gonum.org/v1/gonum/mat"
	"image/color"
	"math"
	"os"
	"time"
)

var (
	projectionStart    string = "Isometric"
	projectionTarget   string = "Isometric"
	projectionDuration float64
	projectionTime     time.Time
	projectionState    bool

	debugView bool

	scaleStart  float64
	scaleTarget float64
	distStart   float64
	distTarget  float64

	dimensions int = 3

	dimensionLock bool

	wind *window

	rotationType string = "Axis"
)

type window struct {
	cube *nCube

	viewPos *mat.VecDense

	dim int
	ncr int

	dir   *mat.VecDense
	rot   *mat.VecDense
	rMats []*mat.Dense

	counter time.Time
}

func (w *window) Update() error {
	if dimensionLock {
		return nil
	}

	// Handle key inputs
	handleInputs(w)

	if projectionState && time.Since(projectionTime).Seconds() > projectionDuration {
		w.viewPos.SetVec(2, distTarget)
		projectionState = false
	} else if projectionState {
		v := time.Since(projectionTime).Seconds() / projectionDuration
		w.viewPos.SetVec(2, lerp(v, distStart, distTarget))
	}

	if time.Since(w.counter).Seconds() > 1 {
		if rotationType == "Axis" {
			w.dir = randAxis(w.ncr, 1.0/50.0) // Pick random rotation axis
		} else {
			w.dir = randUnit(w.ncr, 1.0/50.0) // Pick random unit vector
		}

		w.counter = time.Now() // and reset timer

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
	if dimensionLock {
		return
	}

	rotated := rotate(w.cube.points, w.rot, w.rMats)

	var scale float64

	if projectionState {
		scale = lerp(time.Since(projectionTime).Seconds()/projectionDuration, scaleStart, scaleTarget)
	} else {
		scale = scaleTarget
	}

	for j := 0; j < 1<<w.dim; j++ {
		p1 := mat.VecDenseCopyOf(rotated.ColView(j))
		ps1 := getPerspective(p1, w.viewPos).projectToScreenScale(screen, scale)

		for i := 0; i < 1<<w.dim; i++ {
			if w.cube.edges.At(i, j) != 0 {
				p2 := mat.VecDenseCopyOf(rotated.ColView(i))
				ps2 := getPerspective(p2, w.viewPos).projectToScreenScale(screen, scale)
				ebitenutil.DrawLine(screen, ps1.x, ps1.y, ps2.x, ps2.y, color.White)

				if debugView {
					ebitenutil.DebugPrintAt(screen, vecString(p1), int(ps1.x), int(ps1.y))
					ebitenutil.DebugPrintAt(screen, vecString(p2), int(ps2.x), int(ps2.y))
				}
			}
		}
	}

	// Display keybinds information
	displayKeybinds(screen)
}

func (w *window) Layout(x, y int) (screenWidth, screenHeight int) {
	return x, y
}

func (w *window) init(dim int) {
	ncr := nCr(dim, 2)

	*w = window{
		cube:    makeNCube(dim),
		viewPos: mat.NewVecDense(dim, nil),
		dim:     dim,
		ncr:     ncr,
		dir:     mat.NewVecDense(ncr, nil),
		rot:     mat.NewVecDense(ncr, nil),
		rMats:   givenMats(dim),
		counter: time.Now(),
	}

	dimensions = dim

	switch projectionTarget {
	case "Isometric":
		scaleTarget = math.Sqrt2 * math.Sqrt(float64(dimensions))
		distTarget = 10 + math.Sqrt2*float64(dimensions)
	case "Perspective - Avg", "Perspective - Trim":
		distTarget = 10 + math.Sqrt2*float64(dimensions)
		scaleTarget = 4 * math.Sqrt2 * math.Sqrt(float64(dimensions)) / distTarget
	case "Orthographic":
		scaleTarget = math.Sqrt(float64(dimensions))
		distTarget = 10 + math.Sqrt2*float64(dimensions)
	}

	scaleStart = scaleTarget
	distStart = distTarget

	projectionState = false
	projectionDuration = 0.5

	w.viewPos.SetVec(2, distTarget)

	calculateP2dMat(dim)
	calculateI2dMat(dim)
}

func main() {
	wind = &window{}
	wind.init(dimensions)
	
	if err := ebiten.RunGame(wind); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// Keybind handling
func handleInputs(w *window) {
	if ebiten.IsKeyPressed(ebiten.Key1) {
		changeProjection("Isometric", w)
	}
	if ebiten.IsKeyPressed(ebiten.Key2) {
		changeProjection("Perspective - Avg", w)
	}
	if ebiten.IsKeyPressed(ebiten.Key3) {
		changeProjection("Perspective - Trim", w)
	}
	if ebiten.IsKeyPressed(ebiten.Key4) {
		changeProjection("Orthographic", w)
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		changeDimensions(w, dimensions+1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		changeDimensions(w, dimensions-1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		toggleRotationType()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		toggleDebugView()
	}
}

// Change projection type
func changeProjection(projection string, w *window) {
	if projectionTarget == projection {
		return
	}

	projectionStart = projectionTarget
	projectionTarget = projection
	projectionTime = time.Now()
	projectionState = true

	switch projection {
	case "Isometric":
		scaleStart = scaleTarget
		scaleTarget = math.Sqrt2 * math.Sqrt(float64(dimensions))
	case "Perspective - Avg", "Perspective - Trim":
		distStart = distTarget
		scaleStart = scaleTarget
		distTarget = 10 + math.Sqrt2*float64(dimensions)
		scaleTarget = 4 * math.Sqrt2 * math.Sqrt(float64(dimensions)) / distTarget
	case "Orthographic":
		scaleStart = scaleTarget
		scaleTarget = math.Sqrt(float64(dimensions))
	}
}

// Change the number of dimensions
func changeDimensions(w *window, newDim int) {
	if newDim < 3 || newDim > 9 {
		return
	}

	dimensionLock = true
	dimensions = newDim
	wind.init(dimensions)
	dimensionLock = false
}

// Toggle rotation type between "Axis" and "Unit"
func toggleRotationType() {
	if rotationType == "Axis" {
		rotationType = "Unit"
	} else {
		rotationType = "Axis"
	}
}

// Toggle debug view
func toggleDebugView() {
	debugView = !debugView
}

// Display the keybinds information
func displayKeybinds(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "Keybinds:", 10, 10)
	ebitenutil.DebugPrintAt(screen, "1: Isometric", 10, 30)
	ebitenutil.DebugPrintAt(screen, "2: Perspective - Avg", 10, 50)
	ebitenutil.DebugPrintAt(screen, "3: Perspective - Trim", 10, 70)
	ebitenutil.DebugPrintAt(screen, "4: Orthographic", 10, 90)
	ebitenutil.DebugPrintAt(screen, "Up Arrow: Increase Dimensions", 10, 110)
	ebitenutil.DebugPrintAt(screen, "Down Arrow: Decrease Dimensions", 10, 130)
	ebitenutil.DebugPrintAt(screen, "R: Toggle Rotation Type", 10, 150)
	ebitenutil.DebugPrintAt(screen, "D: Toggle Debug View", 10, 170)
}
