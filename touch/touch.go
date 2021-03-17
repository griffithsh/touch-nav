package touch

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input is a "manager" for touch input and provides logical encapsulations of
// touch interactions like panning, pinching, and tapping.
type Input struct {
	touches map[ebiten.TouchID]*touch

	Pinch *pinch
	Pan   *pan
	Taps  []tap
}

func NewInput() *Input {
	return &Input{
		touches: map[ebiten.TouchID]*touch{},
	}
}

// hypotenuse of a right triangle.
func hypotenuse(xa, ya, xb, yb int) float64 {
	x := math.Abs(float64(xa - xb))
	y := math.Abs(float64(ya - yb))
	return math.Sqrt(x*x + y*y)
}

// Update the touch input manager. After a call to Update, if we have two
// touches, then we have a pinch, if we have one touch, held longer than (N)ms,
// then we have a pan. If we have quickly released presses, then they are
// represented in taps.
func (in *Input) Update() error {
	in.Taps = []tap{}

	// What's gone?
	for id, t := range in.touches {
		if inpututil.IsTouchJustReleased(id) {
			if in.Pinch != nil && (id == in.Pinch.id1 || id == in.Pinch.id2) {
				// FIXME: what about this frame's movement?
				in.Pinch = nil
			}
			if in.Pan != nil && id == in.Pan.id {
				// FIXME: what about this frame's movement?
				in.Pan = nil
			}

			// If this one has not been touched long, or moved far, then it's a
			// tap.
			diff := hypotenuse(t.originX, t.originY, t.currX, t.currY)
			if !t.wasPinch && !t.isPan && (t.duration <= 75 || diff < 5) {
				in.Taps = append(in.Taps, tap{
					X: t.currX,
					Y: t.currY,
				})
			}

			delete(in.touches, id)
		}
	}

	// What's new?
	for _, id := range inpututil.JustPressedTouchIDs() {
		x, y := ebiten.TouchPosition(id)
		in.touches[ebiten.TouchID(id)] = &touch{
			originX: x, originY: y,
			currX: x, currY: y,
		}
	}

	// What's going on?
	for _, id := range ebiten.TouchIDs() {
		t := in.touches[id]
		t.duration = inpututil.TouchPressDuration(id)
		t.currX, t.currY = ebiten.TouchPosition(id)
	}

	if len(in.touches) == 2 {
		// Potential pinch?
		// If the diff between their origins is different to the diff between
		// their currents and if these two are not already a pinch, then this is
		// a new pinch!
		id1, id2 := ebiten.TouchIDs()[0], ebiten.TouchIDs()[1]
		t1, t2 := in.touches[id1], in.touches[id2]
		originDiff := hypotenuse(t1.originX, t1.originY, t2.originX, t2.originY)
		currDiff := hypotenuse(t1.currX, t1.currY, t2.currX, t2.currY)
		if in.Pinch == nil && in.Pan == nil && math.Abs(originDiff-currDiff) > 3 {
			t1.wasPinch = true
			t2.wasPinch = true
			in.Pinch = &pinch{
				id1:     id1,
				id2:     id2,
				originH: originDiff,
				prevH:   originDiff,
			}
		}
	} else if len(in.touches) == 1 {
		// Potential pan.
		id := ebiten.TouchIDs()[0]
		t := in.touches[id]
		if !t.wasPinch && in.Pan == nil && in.Pinch == nil {
			diff := math.Abs(hypotenuse(t.originX, t.originY, t.currX, t.currY))
			if diff > 0.25 {
				t.isPan = true
				in.Pan = &pan{
					id:      id,
					originX: t.originX,
					originY: t.originY,
					prevX:   t.originX,
					prevY:   t.originY,
				}
			}
		}
	}

	return nil
}

type touch struct {
	originX, originY int
	currX, currY     int
	duration         int
	wasPinch, isPan  bool
}

type pinch struct {
	id1, id2 ebiten.TouchID
	originH  float64
	prevH    float64
}

func (p *pinch) currentH() float64 {
	x1, y1 := ebiten.TouchPosition(p.id1)
	x2, y2 := ebiten.TouchPosition(p.id2)
	return hypotenuse(x1, y1, x2, y2)
}

func (p *pinch) Total() float64 {
	return -(p.currentH() - p.originH)
}

// Incremental zoom that has occurred in this frame.
func (p *pinch) Incremental() float64 {
	curr := p.currentH()
	delta := curr - p.prevH
	p.prevH = curr
	return -delta
}

type pan struct {
	id ebiten.TouchID

	prevX, prevY     int
	originX, originY int
}

// Total panning that has occurred since this pan started.
func (p *pan) Total() (float64, float64) {
	currX, currY := ebiten.TouchPosition(p.id)
	return float64(currX - p.originX), float64(currY - p.originY)
}

// Incremental panning that has occurred in this frame.
func (p *pan) Incremental() (float64, float64) {
	currX, currY := ebiten.TouchPosition(p.id)
	deltaX, deltaY := currX-p.prevX, currY-p.prevY

	p.prevX, p.prevY = currX, currY

	return float64(deltaX), float64(deltaY)
}

type tap struct {
	X, Y int
}
