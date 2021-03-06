package main

import (
	"fmt"
	"image"
	"image/color"
	"regexp"
	"strings"
	"time"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/rect"
	"github.com/aarzilli/nucular/style"
	"golang.org/x/mobile/event/key"
)

type infoBoard struct {
	Type string

	FretBoard

	// List board contents
	Root       string
	Scale      string
	ScaleNotes []string
	Chords     ChordMap
}

type NewFretBoard struct {
	Tuning  []string
	Root    string
	Scale   string
	IsScale bool
}

type FretUI struct {
	boards []infoBoard

	scalechords []string

	columns int

	root    string
	scale   string
	isScale bool
	tuning  []string
	error   string

	width  int
	height int

	dirty     time.Time
	saveState State

	tuningEdit  nucular.TextEditor
	scalesearch string
	sclist      []string
	searchEdit  nucular.TextEditor

	// A fretboard to add in the next round
	newFretBoard *NewFretBoard
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

func addChordListBoard(tuning []string, root, scale string) (*infoBoard, error) {
	ret := &infoBoard{
		Type:  TypeList,
		Root:  root,
		Scale: scale,
	}
	var err error

	ret.ScaleNotes, err = GetScale(root, scale)
	if err != nil {
		return nil, err
	}

	ret.Chords, err = GetChordsInScale(root, scale)
	if err != nil {
		return nil, err
	}

	ret.Name = fmt.Sprintf("%s %s chords\nTuning: %s\nNotes: %s",
		root, scale, strings.Join(tuning, ""), strings.Join(ret.ScaleNotes, " "))

	return ret, nil
}

func (f *FretUI) setDirty() {
	f.dirty = time.Now()
}

func (f *FretUI) saveConfig() {
	if !f.dirty.IsZero() && time.Now().After(f.dirty.Add(time.Millisecond*500)) {
		fmt.Println("Saving!")
		err := Save(&f.saveState)
		if err != nil {
			f.error = fmt.Sprintf("Could not save configuration: %v", err)
		}
		f.dirty = time.Time{}
	}
}

func (f *FretUI) drawFretDiagram(w *nucular.Window, fb *infoBoard) {
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
		NoteUnvoiced: {white, black},
		NoteRoot:     {red, black},
		NoteBlack:    {black, white},
		NoteGrey:     {grey, white},
	}

	borderX := bounds.W * 10 / 100
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

	// Maximum width of the fret number box
	fretnumWidth := nucular.FontWidth(fnt, "00") + (bounds.W * 5 / 100)

	// Shift the board to the right a bit so it isn't on top of the numbers
	boardShiftX := bounds.W * 2 / 100

	x := boardBounds.X + fretwidth/2 + boardShiftX
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
		fS := fmt.Sprintf("%2d", i+fb.StartingFret)
		fH := nucular.FontHeight(fnt)
		box := rect.Rect{
			X: x - fretnumWidth - boardShiftX,
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
	var deleteidx = -1
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

func (f *FretUI) ChordListWidget(w *nucular.Window, title string, idx int) int {
	var deleteidx = -1
	if sw := w.GroupBegin(title, nucular.WindowBorder|nucular.WindowNoScrollbar); sw != nil {
		sw.Row(55).Ratio(0.90, 0.10)
		sw.Label(title, "LT")
		if sw.Button(label.T("Close"), false) {
			deleteidx = idx
		} else {
			for _, note := range f.boards[idx].ScaleNotes {
				ch := f.boards[idx].Chords[note]
				sw.Row(20).Dynamic(1)
				sw.Label(note, "LT")
				chordsperrow := 3
				for j := range ch {
					if (j % chordsperrow) == 0 {
						sw.Row(20).Dynamic(chordsperrow)
					}
					if sw.Button(label.T(ch[j]), false) {
						f.newFretBoard = &NewFretBoard{
							Tuning:  f.tuning,
							Root:    note,
							Scale:   ch[j],
							IsScale: false,
						}
					}
				}
			}
		}
		sw.GroupEnd()
	}
	return deleteidx
}

// Add a new fretboard data to display and save
func (f *FretUI) AddFretBoard(tuning []string, root, scale string, isScale bool) error {
	fb, err := addBoard(tuning, root, scale, isScale)
	if err != nil {
		return err
	}

	var ib infoBoard
	ib.FretBoard = *fb
	f.boards = append(f.boards, ib)
	tp := TypeScale
	if !isScale {
		tp = TypeChord
	}
	f.saveState.Boards = append(f.saveState.Boards, BoardState{
		Name:   scale,
		Type:   tp,
		Root:   root,
		Tuning: strings.Join(f.tuning, ""),
	})
	f.setDirty()
	return nil
}

func (f *FretUI) update(w *nucular.Window) {
	for _, e := range w.Input().Keyboard.Keys {
		switch {
		case (e.Modifiers == key.ModControl && e.Code == key.CodeQ):
			go w.Master().Close()
		}
	}

	ratios := []float64{0.1, 0.4, 0.2, 0.1, 0.1, 0.1}
	w.Row(30).Ratio(ratios...)
	w.Label("Root", "LC")
	w.Label("Scale or Chord", "LC")
	w.Label("Tuning", "LC")
	w.Label("", "LC")
	w.Label("", "LC")
	w.Label("Columns", "LC")

	w.Row(30).Ratio(ratios...)

	if w := w.Combo(label.T(f.root), 400, nil); w != nil {
		w.Row(30).Dynamic(1)
		for i := range Notes {
			if w.MenuItem(label.TA(Notes[i], "LC")) {
				f.root = Notes[i]
				f.saveState.Root = f.root
				f.setDirty()
			}
		}
	}

	if w := w.Combo(label.T(f.scale), 1200, nil); w != nil {
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
			f.saveState.ScaleChord = f.scale
			f.setDirty()
			w.Close()
		}

		for i := range f.sclist {
			if w.MenuItem(label.TA(f.sclist[i], "LC")) {
				ret := strings.Replace(f.sclist[i], "Scale: ", "", 1)
				f.isScale = (ret != f.sclist[i])

				ret = strings.Replace(ret, "Chord: ", "", 1)
				f.scale = ret
				f.saveState.ScaleChord = f.scale
				f.setDirty()
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

	if f.newFretBoard != nil {
		err = f.AddFretBoard(f.newFretBoard.Tuning, f.newFretBoard.Root, f.newFretBoard.Scale, f.newFretBoard.IsScale)
		if err != nil {
			f.error = fmt.Sprintf("Error: %v", err)
		}
		f.newFretBoard = nil
	}

	if w.Button(label.T("Frets"), false) {
		f.tuning, err = parseTuning(string(f.tuningEdit.Buffer))
		if err != nil {
			f.error = fmt.Sprintf("Error: %v", err)
		} else {
			f.saveState.Tuning = strings.Join(f.tuning, "")
			f.setDirty()

			err = f.AddFretBoard(f.tuning, f.root, f.scale, f.isScale)
			if err != nil {
				f.error = fmt.Sprintf("Error: %v", err)
			}
		}
	}
	if w.Button(label.T("Chords"), false) {
		if !f.isScale {
			f.error = fmt.Sprintf("Given scale is not a scale: %s", f.scale)
		} else {
			f.tuning, err = parseTuning(string(f.tuningEdit.Buffer))
			if err != nil {
				f.error = fmt.Sprintf("Error: %v", err)
			} else {
				f.error = ""
				ib, err := addChordListBoard(f.tuning, f.root, f.scale)
				if err != nil {
					f.error = fmt.Sprintf("Error: %v", err)
				} else {
					f.boards = append(f.boards, *ib)
					f.saveState.Tuning = strings.Join(f.tuning, "")
					f.saveState.Boards = append(f.saveState.Boards, BoardState{
						Name:   f.scale,
						Type:   TypeList,
						Root:   f.root,
						Tuning: strings.Join(f.tuning, ""),
					})
					f.setDirty()
				}
			}
		}
	}

	w.PropertyInt("", 1, &f.columns, 5, 1, 1)

	w.Row(30).Dynamic(1)
	w.Label(f.error, "LC")

	if f.columns != f.saveState.Columns {
		f.saveState.Columns = f.columns
		f.setDirty()
	}

	var deleteidx = -1
	for i := range f.boards {
		if i%f.columns == 0 {
			w.Row(700).Dynamic(f.columns)
		}
		var di int
		if f.boards[i].Type == TypeList {
			di = f.ChordListWidget(w, f.boards[i].Name, i)
		} else {
			di = f.FretWidget(w, f.boards[i].Name, i)
		}
		if di >= 0 {
			deleteidx = di
		}
	}

	// Remove the fretboard if user wanted to close one of them
	if deleteidx >= 0 {
		f.boards = append(f.boards[:deleteidx], f.boards[deleteidx+1:]...)
		f.saveState.Boards = append(f.saveState.Boards[:deleteidx], f.saveState.Boards[deleteidx+1:]...)
		f.setDirty()
	}

	if w.Bounds.H != f.height || w.Bounds.W != f.width {
		fmt.Printf("old [%dx%d] new [%dx%d]\n",
			f.width, f.height,
			w.Bounds.W, w.Bounds.H)
		f.height = w.Bounds.H
		f.width = w.Bounds.W
		f.saveState.Height = f.height
		f.saveState.Width = f.width
		f.setDirty()
	}

	f.saveConfig()
}

func NewFretUI() *FretUI {
	fu := &FretUI{
		columns: 4,
		root:    "E",
		scale:   "Major (Ionian)",
		isScale: true,
		tuning:  []string{"E", "A", "D", "G", "B", "E"},
		width:   700,
		height:  830,
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
		for i := range Notes {
			if ss.Root == Notes[i] {
				fu.root = ss.Root
			}
		}

		if _, ok := Scales[ss.ScaleChord]; ok {
			fu.scale = ss.ScaleChord
			fu.isScale = true
		} else if _, ok := Chords[ss.ScaleChord]; ok {
			fu.scale = ss.ScaleChord
			fu.isScale = false
		}

		if ss.Width != 0 {
			fu.width = ss.Width
		}
		if ss.Height != 0 {
			fu.height = ss.Height
		}

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

			tp := ss.Boards[i].Type
			if tp == TypeScale || tp == TypeChord {
				isscale := true
				if tp == TypeChord {
					isscale = false
				}
				var fb *FretBoard
				fb, err = addBoard(tuning, ss.Boards[i].Root, ss.Boards[i].Name, isscale)
				if err == nil {
					var ib infoBoard
					ib.FretBoard = *fb
					fu.boards = append(fu.boards, ib)
				}
			} else {
				ib, err := addChordListBoard(tuning, ss.Boards[i].Root, ss.Boards[i].Name)
				if err == nil {
					fu.boards = append(fu.boards, *ib)
				}
			}

		}
	}

	return fu
}

func (f *FretUI) FilterScaleChords(filter string) []string {
	if filter == "" {
		return f.scalechords
	}
	re, err := regexp.Compile(`(?i)` + filter)
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
	w := nucular.NewMasterWindowSize(0, title, image.Point{fu.width, fu.height}, fu.update)

	w.SetStyle(style.FromTheme(style.DarkTheme, 1.0))

	w.Main()
	return nil
}
