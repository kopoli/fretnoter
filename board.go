package main

type Note struct {
	String int
	Fret   int
	Name   string
}

type FretBoard struct {
	Strings      int
	Frets        int
	StartingFret int
	Notes        []Note
}

func (f *FretBoard) Print() {
}
