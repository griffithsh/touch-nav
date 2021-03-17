package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/griffithsh/touch-nav/touch"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type demo struct {
	touchManager *touch.Input

	zoom       float64
	panX, panY float64
}

func (d *demo) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{127, 0, 127, 255})

	msg := fmt.Sprintf("pan: %f,%f\nzoom: %f\ntaps: %v", d.panX, d.panY, d.zoom, d.touchManager.Taps)

	x := screen.Bounds().Max.X / 2
	y := screen.Bounds().Max.Y / 2
	ebitenutil.DebugPrintAt(screen, msg, x, y)
}

func (d *demo) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (d *demo) Update() error {
	if err := d.touchManager.Update(); err != nil {
		return err
	}
	for _, tap := range d.touchManager.Taps {
		fmt.Printf("tap! (%d,%d)\n", tap.X, tap.Y)
	}
	if d.touchManager.Pan != nil {
		x, y := d.touchManager.Pan.Incremental()
		d.panX += x
		d.panY += y
	}

	if d.touchManager.Pinch != nil {
		d.zoom += d.touchManager.Pinch.Incremental()
	}

	return nil
}

func main() {
	d := demo{
		touchManager: touch.NewInput(),
	}
	ebiten.SetWindowSize(240, 320)
	if err := ebiten.RunGame(&d); err != nil {
		log.Fatal(err)
	}
}
