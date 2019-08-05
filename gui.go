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

	white := color.RGBA{0xff, 0xff, 0xff, 0xff}
	black := color.RGBA{0, 0, 0, 0xff}
	red := color.RGBA{0xff, 0, 0, 0xff}
	grey := color.RGBA{0x80, 0x80, 0x80, 0xff}

	type noteColor struct {
		back color.RGBA
		fore color.RGBA
	}

	circleColors := map[NoteType]noteColor{
		NoteUnvoiced: noteColor{white, black},
		NoteRoot:     noteColor{red, black},
		NoteBlack:    noteColor{black, white},
		NoteGrey:     noteColor{grey, white},
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

	// Get a font that is relatively scaled (the 12.0 is from Style.DefaultFont)
	fontscaling := (float64(fretheight) * 0.4) / 12.0
	s.DefaultFont(fontscaling)
	fnt := s.Font
	s.DefaultFont(s.Scaling) // Get the default font back

	x := boardBounds.X
	y := boardBounds.Y + fretheight

	// fmt.Println("fretwidth", fretwidth)

	out.FillRect(bounds, 0, white)

	// there is some rounding error between this and boardBounds.Max().Y
	maxy := y + fretheight*fb.Frets

	// Print fret grid
	for i := 0; i < fb.Strings+1; i++ {
		xpos := x + fretwidth*i
		start := image.Point{xpos, y}
		stop := image.Point{xpos, maxy}
		out.StrokeLine(start, stop, 2, black)
	}
	for i := 0; i < fb.Frets+1; i++ {
		ypos := y + fretheight*i
		start := image.Point{x, ypos}
		stop := image.Point{boardBounds.Max().X, ypos}
		out.StrokeLine(start, stop, 2, black)
	}

	// fmt.Println("maxy", boardBounds.Max().Y, "max line y", y + fretheight * fb.Frets)

	// Print fret numbers
	for i := 0; i < fb.Frets+1; i++ {
		fS := fmt.Sprintf("%d", i+fb.StartingFret)
		fH := nucular.FontHeight(fnt)
		box := rect.Rect{
			X: x - borderX/2,
			Y: y + fretheight*i - fH/2,
			W: borderX,
			H: fretheight,
		}
		out.DrawText(box, fS, fnt, black)
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

		fW := nucular.FontWidth(fnt, note.Name)
		fH := nucular.FontHeight(fnt)
		fbox := rect.Rect{
			X: x + note.String*fretwidth - fW/2,
			Y: y + (note.Fret-1)*fretheight + (fretheight-fH)/2,
			W: fW,
			H: fH,
		}
		out.DrawText(fbox, note.Name, fnt, circleColors[note.Type].fore)
	}
}

func (f *FretUI) update(w *nucular.Window) {
	w.Row(25).Dynamic(1)
	w.Label("helloworld", "LC")

	w.Row(700).Dynamic(2)
	f.drawFretDiagram(w, &f.frets)
	f.drawFretDiagram(w, &f.frets)
}

func GUIMain(version string) error {
	fu := &FretUI{
		frets: FretBoard{
			Strings:      6,
			Frets:        12,
			StartingFret: 0,
			Notes: []Note{
				Note{2, 0, "X", NoteUnvoiced},
				Note{1, 1, "Z", NoteGrey},
				Note{3, 2, "Z", NoteRoot},
				Note{4, 7, "Z", NoteBlack},
			},
		},
	}

	// Print a scale
	scl, err := GetScale("D", "Natural Minor")
	if err != nil {
		return err
	}
	fu.frets.InitGuitar()
	fu.frets.Clear()
	fu.frets.SetNotes(scl, NoteBlack)
	fu.frets.SetNotes([]string{"D"}, NoteRoot)

	title := fmt.Sprintf("Fretnoter %s", version)
	w := nucular.NewMasterWindowSize(0, title, image.Point{640, 630}, fu.update)

	w.SetStyle(style.FromTheme(style.DarkTheme, 1.0))

	w.Main()
	return nil
}
