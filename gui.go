package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/aarzilli/nucular"
	// "github.com/aarzilli/nucular/label"

	"github.com/aarzilli/nucular/rect"
	"github.com/aarzilli/nucular/style"
)

type FretUI struct {
	frets FretBoard
}

func (f *FretUI) drawFretDiagram(w *nucular.Window, fb *FretBoard) {
	bounds, out := w.Custom(style.WidgetStateInactive)
	if out == nil {
		return
	}

	mw := w.Master()
	s := mw.Style()

	background := color.RGBA{0xff, 0xff, 0xff, 0xff}
	foreground := color.RGBA{0, 0, 0, 0xff}
	red := color.RGBA{0xff, 0, 0, 0xff}
	grey := color.RGBA{0x80, 0x80, 0x80, 0xff}

	type noteColor struct {
		back color.RGBA
		fore color.RGBA
	}

	circleColors := map[int]noteColor{
		NoteUnvoiced: noteColor{background, foreground},
		NoteRoot:     noteColor{red, foreground},
		NoteBlack:    noteColor{foreground, background},
		NoteGrey:     noteColor{grey, background},
	}

	borderX := bounds.W * 5 / 100
	borderY := bounds.H * 5 / 100

	boardBounds := rect.Rect{
		X: bounds.X + borderX,
		Y: bounds.Y + borderY,
		W: bounds.W - (borderX * 2),
		H: bounds.H - (borderY * 2),
	}

	fretwidth := boardBounds.W / fb.Strings
	fretheight := boardBounds.H / (fb.Frets + 1)

	x := boardBounds.X
	y := boardBounds.Y + fretheight

	// fmt.Println("fretwidth", fretwidth)

	out.FillRect(bounds, 0, background)

	// there is some rounding error between this and boardBounds.Max().Y
	maxy := y + fretheight*fb.Frets

	// Print fret grid
	for i := 0; i < fb.Strings+1; i++ {
		xpos := x + fretwidth*i
		start := image.Point{xpos, y}
		stop := image.Point{xpos, maxy}
		out.StrokeLine(start, stop, 2, foreground)
	}
	for i := 0; i < fb.Frets+1; i++ {
		ypos := y + fretheight*i
		start := image.Point{x, ypos}
		stop := image.Point{boardBounds.Max().X, ypos}
		out.StrokeLine(start, stop, 2, foreground)
	}

	// fmt.Println("maxy", boardBounds.Max().Y, "max line y", y + fretheight * fb.Frets)

	// Print fret numbers
	for i := 0; i < fb.Frets+1; i++ {
		fS := fmt.Sprintf("%d", i+fb.StartingFret)
		fH := nucular.FontHeight(s.Font)
		box := rect.Rect{
			X: x - borderX/2,
			Y: y + fretheight*i - fH/2,
			W: borderX,
			H: fretheight,
		}
		out.DrawText(box, fS, s.Font, foreground)
	}

	circleW := fretheight
	if fretheight > fretwidth {
		circleW = fretwidth
	}
	circleW = circleW * 95 / 100

	// Print note circles and texts
	for _, note := range fb.Notes {
		box := rect.Rect{
			X: x + note.String*fretwidth - circleW/2,
			Y: y + (note.Fret-1)*fretheight + (fretheight-circleW)/2,
			W: circleW,
			H: circleW,
		}
		out.FillCircle(box, circleColors[note.Type].back)

		fW := nucular.FontWidth(s.Font, note.Name)
		fH := nucular.FontHeight(s.Font)
		fbox := rect.Rect{
			X: x + note.String*fretwidth - fW/2,
			Y: y + (note.Fret-1)*fretheight + (fretheight-fH)/2,
			W: fW,
			H: fH,
		}
		out.DrawText(fbox, note.Name, s.Font, circleColors[note.Type].fore)
	}
}

func (f *FretUI) update(w *nucular.Window) {
	w.Row(25).Dynamic(1)
	w.Label("helloworld", "LC")

	w.Row(500).Dynamic(2)
	f.drawFretDiagram(w, &f.frets)
	f.drawFretDiagram(w, &f.frets)
}

func GUIMain(version string) error {
	fu := &FretUI{
		frets: FretBoard{
			Strings:      6,
			Frets:        6,
			StartingFret: 0,
			Notes: []Note{
				Note{2, 0, "X", NoteUnvoiced},
				Note{1, 1, "Z", NoteUnvoiced},
				Note{3, 2, "Z", NoteRoot},
				Note{4, 7, "Z", NoteBlack},
			},
		},
	}

	title := fmt.Sprintf("Fretnoter %s", version)
	w := nucular.NewMasterWindowSize(0, title, image.Point{640, 630}, fu.update)

	w.SetStyle(style.FromTheme(style.DarkTheme, 1.0))

	w.Main()
	return nil
}
