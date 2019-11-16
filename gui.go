package main

import (
	"fmt"
	"image"
	"image/color"
	"regexp"
	"strings"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/rect"
	"github.com/aarzilli/nucular/style"
)

type FretUI struct {
	boards []FretBoard

	scalechords []string

	columns int

	root    string
	scale   string
	isScale bool
	tuning  []string
	error   string

	saveState State

	tuningEdit nucular.TextEditor
	scalesearch string
	sclist []string
	searchEdit nucular.TextEditor
}

var (
	tuningRe = regexp.MustCompile(`[ABCDEFGabcdefg]#?`)
)

func parseTuning(tuning string) ([]string, error) {
	splits := tuningRe.FindAllString(tuning, -1)
	if splits == nil {
		return nil, fmt.Errorf("invalid tuning given")
	}

	for i := range splits {
		splits[i] = strings.TrimSpace(splits[i])
		splits[i] = strings.ToUpper(splits[i])

		found := false
		for j := range Notes {
			if splits[i] == Notes[j] {
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid note '%s' given", splits[i])
		}
	}

	return splits, nil
}

func addBoard(tuning []string, root, scale string, isScale bool) (*FretBoard, error) {
	ret := &FretBoard{
		Strings:      len(tuning),
		Frets:        11,
		StartingFret: 0,
		Tuning:       tuning,
	}

	var notes []string
	var err error
	var boardtype string

	if isScale {
		notes, err = GetScale(root, scale)
		boardtype = "scale"
	} else {
		notes, err = GetChord(root, scale)
		boardtype = "chord"
	}
	if err != nil {
		return nil, err
	}

	err = ret.SetNotes(notes[1:], NoteBlack)
	if err != nil {
		return nil, err
	}

	err = ret.SetNotes([]string{root}, NoteRoot)
	if err != nil {
		return nil, err
	}

	ret.Name = fmt.Sprintf("%s %s %s\nTuning: %s\nNotes: %s",
		root, scale, boardtype,
		strings.Join(tuning, ""),
		strings.Join(notes, " "))

	return ret, nil
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

	x := boardBounds.X + fretwidth/2
	y := boardBounds.Y + fretheight

	// Draw the background
	out.FillRect(bounds, 0, white)

	// there is some rounding error between this and boardBounds.Max()
	maxy := y + fretheight*(fb.Frets)
	maxx := x + fretwidth*(fb.Strings-1)

	// Print fret grid
	for i := 0; i < fb.Strings; i++ {
		xpos := x + fretwidth*i
		start := image.Point{xpos, y}
		stop := image.Point{xpos, maxy}
		out.StrokeLine(start, stop, 2, black)
	}
	for i := 0; i < fb.Frets+1; i++ {
		ypos := y + fretheight*i
		start := image.Point{x, ypos}
		stop := image.Point{maxx, ypos}
		out.StrokeLine(start, stop, 2, black)
	}

	// Print fret numbers
	for i := 0; i < fb.Frets+1; i++ {
		fS := fmt.Sprintf("%d", i+fb.StartingFret)
		fH := nucular.FontHeight(fnt)
		box := rect.Rect{
			X: x - fretwidth/2,
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

func (f *FretUI) FretWidget(w *nucular.Window, title string, idx int) int {
	var deleteidx int = -1
	if sw := w.GroupBegin(title, nucular.WindowBorder|nucular.WindowNoScrollbar); sw != nil {
		sw.Row(55).Ratio(0.90, 0.10)
		sw.Label(title, "LT")
		if sw.Button(label.T("Close"), false) {
			deleteidx = idx
		} else {
			sw.Row(0).Dynamic(1)
			f.drawFretDiagram(sw, &f.boards[idx])
		}
		sw.GroupEnd()
	}
	return deleteidx
}

func (f *FretUI) update(w *nucular.Window) {
	ratios := []float64{0.1, 0.4, 0.3, 0.1, 0.1}
	w.Row(30).Ratio(ratios...)
	w.Label("Root", "LC")
	w.Label("Scale or Chord", "LC")
	w.Label("Tuning", "LC")
	w.Label("", "LC")
	w.Label("Columns", "LC")

	w.Row(30).Ratio(ratios...)

	if w := w.Combo(label.T(f.root), 400, nil); w != nil {
		w.Row(30).Dynamic(1)
		for i := range Notes {
			if w.MenuItem(label.TA(Notes[i], "LC")) {
				f.root = Notes[i]
			}
		}
	}

	if w := w.Combo(label.T(f.scale), 400, nil); w != nil {
		w.Row(30).Dynamic(1)
		f.searchEdit.Active = true
		a := f.searchEdit.Edit(w)
		if f.scalesearch != string(f.searchEdit.Buffer) {
			f.scalesearch = string(f.searchEdit.Buffer)
			f.sclist = f.FilterScaleChords(f.scalesearch)
		}
		if a&nucular.EditCommitted != 0 {
			if len(f.sclist) > 0 {
				f.scale = f.sclist[0]
			} else {
				f.scale = f.scalechords[0]
			}
			w.Close()
		}

		for i := range f.sclist {
			if w.MenuItem(label.TA(f.sclist[i], "LC")) {
				ret := strings.Replace(f.sclist[i], "Scale: ", "", 1)
				f.isScale = (ret != f.sclist[i])

				ret = strings.Replace(ret, "Chord: ", "", 1)
				f.scale = ret
			}
		}
	}

	var err error
	a := f.tuningEdit.Edit(w)
	if a&nucular.EditCommitted != 0 {
		f.tuning, err = parseTuning(string(f.tuningEdit.Buffer))
		if err != nil {
			f.error = fmt.Sprintf("Error: %v", err)
		} else {
			f.error = ""
			f.tuningEdit.Buffer = []rune(strings.Join(f.tuning, ""))
		}
	}
	if w.Button(label.T("Add"), false) {
		f.tuning, err = parseTuning(string(f.tuningEdit.Buffer))
		if err != nil {
			f.error = fmt.Sprintf("Error: %v", err)
		} else {
			f.error = ""
			f.tuningEdit.Buffer = []rune(strings.Join(f.tuning, ""))
			var fb *FretBoard
			fb, err = addBoard(f.tuning, f.root, f.scale, f.isScale)
			if err != nil {
				f.error = fmt.Sprintf("Error: %v", err)
			} else {
				f.boards = append(f.boards, *fb)
				f.saveState.Tuning = strings.Join(f.tuning, "")
				f.saveState.Boards = append(f.saveState.Boards, BoardState{
					Root:    f.root,
					Scale:   f.scale,
					IsScale: f.isScale,
					Tuning:  strings.Join(f.tuning, ""),
				})
				_ = Save(&f.saveState)
			}
		}
	}
	w.PropertyInt("", 1, &f.columns, 5, 1, 1)
	w.Row(30).Dynamic(1)
	w.Label(f.error, "LC")

	if f.columns != f.saveState.Columns {
		f.saveState.Columns = f.columns
		_ = Save(&f.saveState)
	}

	var deleteidx int = -1
	for i := range f.boards {
		if i%f.columns == 0 {
			w.Row(700).Dynamic(f.columns)
		}
		di := f.FretWidget(w, f.boards[i].Name, i)
		if di >= 0 {
			deleteidx = di
		}
	}

	// Remove the fretboard if user wanted to close one of them
	if deleteidx >= 0 {
		f.boards = append(f.boards[:deleteidx], f.boards[deleteidx+1:]...)
		f.saveState.Boards = append(f.saveState.Boards[:deleteidx], f.saveState.Boards[deleteidx+1:]...)
		_ = Save(&f.saveState)
	}
}

func NewFretUI() *FretUI {
	fu := &FretUI{
		columns: 4,
		root:    "D",
		scale:   "Natural Minor (Aeolian)",
		isScale: true,
		tuning:  []string{"D", "A", "D", "G", "B", "E"},
	}

	fu.searchEdit.Flags = nucular.EditField
	fu.searchEdit.Maxlen = 64

	fu.tuningEdit.Flags = nucular.EditField
	fu.tuningEdit.Maxlen = 64
	fu.tuningEdit.Buffer = []rune(strings.Join(fu.tuning, ""))

	fu.scalechords = make([]string, 0, len(Scales)+len(Chords))
	for s := range Scales {
		fu.scalechords = append(fu.scalechords, "Scale: "+s)
	}

	for c := range Chords {
		fu.scalechords = append(fu.scalechords, "Chord: "+c)
	}
	fu.sclist = fu.scalechords

	ss, err := Load()
	if err == nil {
		fu.saveState = *ss
		fu.columns = ss.Columns
		var tuning []string
		tuning, err = parseTuning(ss.Tuning)
		if err == nil {
			fu.tuning = tuning
			fu.tuningEdit.Buffer = []rune(strings.Join(fu.tuning, ""))
		}

		for i := range ss.Boards {
			tuning, err = parseTuning(ss.Boards[i].Tuning)
			if err != nil {
				continue
			}

			var fb *FretBoard
			fb, err = addBoard(tuning, ss.Boards[i].Root, ss.Boards[i].Scale, ss.Boards[i].IsScale)
			if err == nil {
				fu.boards = append(fu.boards, *fb)
			}
		}
	}

	return fu
}

func (f *FretUI) FilterScaleChords(filter string) []string {
	if filter == "" {
		return f.scalechords
	}
	re, err := regexp.Compile(`(?i)` +filter)
	if err != nil {
		return f.scalechords
	}

	var ret []string
	for i := range f.scalechords {
		if re.FindStringIndex(f.scalechords[i]) != nil {
			ret = append(ret, f.scalechords[i])
		}
	}

	return ret
}

func GUIMain(version string) error {
	fu := NewFretUI()

	title := fmt.Sprintf("Fretnoter %s", version)
	w := nucular.NewMasterWindowSize(0, title, image.Point{700, 830}, fu.update)

	w.SetStyle(style.FromTheme(style.DarkTheme, 1.0))

	w.Main()
	return nil
}
