package main

const (
	NoteUnvoiced = iota
	NoteRoot
	NoteBlack
	NoteGrey
)

type Note struct {
	String int
	Fret   int
	Name   string
	Type   int
}

type FretBoard struct {
	Strings      int
	Frets        int
	StartingFret int
	Notes        []Note
}

func (f *FretBoard) Print() {
}
