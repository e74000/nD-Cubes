package main

import (
	"fmt"
	"github.com/e74000/wshim"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"gonum.org/v1/gonum/mat"
	"image/color"
	"math"
	"os"
	"time"
)

var (
	projectionStart    string
	projectionTarget   string
	projectionDuration float64
	projectionTime     time.Time
	projectionState    bool

	debugView bool

	scaleStart  float64
	scaleTarget float64
	distStart   float64
	distTarget  float64

	dimensions int

	dimensionLock bool

	wind *window

	rotationType string
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

	projectionStart = projectionTarget
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
	dimensions = 3

	wshim.Run(
		mainFunc,
		wshim.Radio("Perspective", []string{"Isometric", "Perspective - Avg", "Perspective - Trim", "Orthographic"}, &projectionTarget).
			OnChange(func(oldVal, newVal string) {
				projectionStart = oldVal
				projectionTime = time.Now()
				projectionState = true

				switch newVal {
				case "Isometric":
					fmt.Println("Changing projection to Isometric")
					scaleStart = scaleTarget
					scaleTarget = math.Sqrt2 * math.Sqrt(float64(dimensions))
				case "Perspective - Avg", "Perspective - Trim":
					fmt.Println("Changing projection to perspective")
					distStart = distTarget
					scaleStart = scaleTarget
					distTarget = 10 + math.Sqrt2*float64(dimensions)
					scaleTarget = 4 * math.Sqrt2 * math.Sqrt(float64(dimensions)) / distTarget
				case "Orthographic":
					fmt.Println("Changing projection to orthographic")
					scaleStart = scaleTarget
					scaleTarget = math.Sqrt(float64(dimensions))
				}

			}),
		wshim.IntSlider("Dimensions", 3, 9, 1, &dimensions).
			OnChange(func(oldVal, newVal int) {
				if oldVal == newVal {
					return
				}

				dimensionLock = true
				dimensions = newVal
				wind.init(dimensions)
				dimensionLock = false
			}),
		wshim.Radio("Rotation type", []string{"Unit", "Axis"}, &rotationType),
		wshim.Toggle("Debug View", &debugView),
	)
}

func mainFunc() {
	wind = &window{}
	wind.init(dimensions)

	if err := ebiten.RunGame(wind); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
