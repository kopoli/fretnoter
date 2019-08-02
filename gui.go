package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/aarzilli/nucular"
	// "github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/command"
	"github.com/aarzilli/nucular/rect"
	"github.com/aarzilli/nucular/style"
)

type FretUI struct {
	frets FretBoard
}

func (f *FretUI) drawFrets(fb *FretBoard, bounds rect.Rect, out *command.Buffer) {
	background := color.RGBA{0xff, 0xff, 0xff, 0xff}
	foreground := color.RGBA{0, 0, 0, 0xff}
	out.FillRect(bounds, 0, background)

	borderX := bounds.W * 5 / 100
	borderY := bounds.H * 5 / 100

	boardBounds := rect.Rect{
		X: bounds.X + borderX,
		Y: bounds.Y + borderY,
		W: bounds.W - (borderX * 2),
		H: bounds.H - (borderY * 2),
	}

	x := boardBounds.Min().X
	y := boardBounds.Min().Y

	fretwidth := boardBounds.W / fb.Strings
	fretheight := boardBounds.H / fb.Frets

	fmt.Println("fretwidth", fretwidth)

	for i := 0; i < fb.Strings + 1; i++ {
		xpos := x + fretwidth*i
		start := image.Point{xpos, y}
		stop := image.Point{xpos, boardBounds.Max().Y}
		out.StrokeLine(start, stop, 2, foreground)
	}

	for i := 0; i < fb.Frets + 1; i++ {
		ypos := y + fretheight*i
		start := image.Point{x, ypos}
		stop := image.Point{boardBounds.Max().X, ypos}
		out.StrokeLine(start, stop, 2, foreground)
	}

	clr := color.RGBA{0xff, 0x00, 0xff, 0xff}
	bl := image.Point{bounds.X, bounds.Y + bounds.H - 1}
	tr := image.Point{bounds.X + bounds.W - 1, bounds.Y}
	br := image.Point{bounds.X + bounds.W - 1, bounds.Y + bounds.H - 1}
	out.StrokeLine(bounds.Min(), bl, 1, clr)
	out.StrokeLine(bl, br, 1, clr)
	out.StrokeLine(br, tr, 1, clr)
	out.StrokeLine(tr, bounds.Min(), 1, clr)
	out.StrokeLine(bounds.Min().Add(image.Point{50, 50}), bounds.Max().Add(image.Point{-50, -50}), 5, clr)
}

func (f *FretUI) update(w *nucular.Window) {
	w.Row(25).Dynamic(1)
	w.Label("helloworld", "LC")

	w.Row(300).Dynamic(1)
	bounds, out := w.Custom(style.WidgetStateInactive)
	if out != nil {

		f.drawFrets(&f.frets, bounds, out)
	}
}

func GUIMain(version string) error {
	fu := &FretUI{
		frets: FretBoard{
			Strings:      6,
			Frets:        6,
			StartingFret: 0,
		},
	}

	title := fmt.Sprintf("Fretnoter %s", version)
	w := nucular.NewMasterWindowSize(0, title, image.Point{640, 630}, fu.update)

	w.SetStyle(style.FromTheme(style.DarkTheme, 1.0))

	w.Main()
	return nil
}
