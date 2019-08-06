package main

const (
	NoteUnvoiced = iota
	NoteRoot
	NoteBlack
	NoteGrey
)

type NoteType int

type Note struct {
	String int
	Fret   int
	Name   string
	Type   NoteType
}

type FretBoard struct {
	Name         string
	Tuning       []string
	Strings      int
	Frets        int
	StartingFret int
	Notes        []Note
}

func (f *FretBoard) SetNotes(notes []string, ntype NoteType) error {
	var err error
	notemap := map[string]bool{}
	for i := range notes {
		_, err = NotePosition(notes[i])
		if err != nil {
			return err
		}
		notemap[notes[i]] = true
	}

	for s := 0; s < f.Strings; s++ {
		for fr := 0; fr < f.Frets+1; fr++ {
			note, _ := GetNote(f.Tuning[s], fr)
			if _, ok := notemap[note]; ok {
				f.Notes = append(f.Notes, Note{
					String: s,
					Fret:   fr,
					Name:   note,
					Type:   ntype,
				})
			}
		}
	}

	return nil
}

func (f *FretBoard) Print() {
}

func (f *FretBoard) Clear() {
	f.Notes = nil
}
