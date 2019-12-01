package main

import (
	"fmt"
	"sort"
	"strings"
)

// https://gist.github.com/inky/3188870

var Notes = []string{"A", "A#", "B", "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#"}

// Scales as steps from the previous note
var Scales = map[string][]int{
	"Major (Ionian)":           []int{2, 2, 1, 2, 2, 2, 1},
	"Dorian Mode":              []int{2, 1, 2, 2, 2, 1, 2},
	"Phrygian Mode":            []int{1, 2, 2, 2, 1, 2, 2},
	"Lydian Mode":              []int{2, 2, 2, 1, 2, 2, 1},
	"Mixolydian Mode":          []int{2, 2, 1, 2, 2, 1, 2},
	"Natural Minor (Aeolian)":  []int{2, 1, 2, 2, 1, 2, 2},
	"Locrian Mode":             []int{1, 2, 2, 1, 2, 2, 2},
	"Harmonic Minor":           []int{2, 1, 2, 2, 1, 3, 1},
	"Locrian nat6":             []int{1, 2, 2, 1, 3, 1, 2},
	"Ionian #5":                []int{2, 2, 1, 3, 1, 2, 1},
	"Ukranian minor":           []int{2, 1, 3, 1, 2, 1, 2},
	"Phrygian dominant":        []int{1, 3, 1, 2, 1, 2, 2},
	"Lydian #2":                []int{3, 1, 2, 1, 2, 2, 1},
	"Super Locrian diminished": []int{1, 2, 1, 2, 2, 1, 3},
	"Diminished":               []int{2, 1, 2, 1, 2, 1, 2, 1},
	"Dominant Diminished":      []int{1, 2, 1, 2, 1, 2, 1, 2},
	"Pentatonic Major":         []int{2, 2, 3, 2, 3},
	"Pentatonic Minor":         []int{3, 2, 2, 3, 2},
}

// Chords as the distance from the root note
var Chords = map[string][]int{
	"Major":      []int{0, 4, 7},
	"Minor":      []int{0, 3, 7},
	"Augmented":  []int{0, 4, 8},
	"Diminished": []int{0, 4, 6},
	"sus2":       []int{0, 2, 7},
	"sus4":       []int{0, 5, 7},
	"Power":      []int{0, 7},
	"maj7":       []int{0, 4, 7, 11},
	"dom7":       []int{0, 4, 7, 10},
	"dim7":       []int{0, 3, 6, 9},
	"dom7f5":     []int{0, 4, 6, 10},
	"halfdim7":   []int{0, 3, 6, 10},
	"majdim7":    []int{0, 3, 6, 11},
	"minmaj7":    []int{0, 3, 7, 11},
	"augmaj7":    []int{0, 4, 8, 11},
	"aug7":       []int{0, 4, 8, 10},
	"7sus2":      []int{0, 5, 7, 10},
}

func NotePosition(note string) (int, error) {
	for i := range Notes {
		if Notes[i] == note {
			return i, nil
		}
	}
	return -1, fmt.Errorf("note '%s' doesn't exist", note)
}

func GetNote(note string, steps int) (string, error) {
	pos, err := NotePosition(note)
	if err != nil {
		return "", err
	}

	return Notes[(pos+steps)%len(Notes)], nil
}

func GetScale(note, scale string) ([]string, error) {
	pos, err := NotePosition(note)
	if err != nil {
		return nil, err
	}

	if _, ok := Scales[scale]; !ok {
		return nil, fmt.Errorf("scale '%s' doesn't exist", scale)
	}

	ret := make([]string, 0, len(Scales[scale]))

	for _, steps := range Scales[scale] {
		ret = append(ret, Notes[pos])
		pos = (pos + steps) % len(Notes)
	}

	return ret, nil
}

func GetChord(note, chord string) ([]string, error) {
	pos, err := NotePosition(note)
	if err != nil {
		return nil, err
	}

	if _, ok := Chords[chord]; !ok {
		return nil, fmt.Errorf("chord '%s' doesn't exist", chord)
	}

	ret := make([]string, 0, len(Chords[chord]))

	for _, distance := range Chords[chord] {
		ret = append(ret, Notes[(pos+distance)%len(Notes)])
	}

	return ret, nil
}

func IsChordInScale(chordNotes, scaleNotes []string) bool {
	for i := range chordNotes {
		found := false
		for j := range scaleNotes {
			if chordNotes[i] == scaleNotes[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type ChordMap map[string][]string

// Get the chords that can be played with the notes in the given scale
func GetChordsInScale(root, scale string) (ChordMap, error) {
	scalenotes, err := GetScale(root, scale)
	if err != nil {
		return nil, err
	}

	scalechords := ChordMap{}

	// Get the chords in a map with the Note name as the key
	for i := range Notes {
		for j := range Chords {
			notes, err := GetChord(Notes[i], j)
			if err != nil {
				return nil, err
			}
			if IsChordInScale(notes, scalenotes) {
				scalechords[Notes[i]] = append(scalechords[Notes[i]], j)
			}
		}
		sort.Strings(scalechords[Notes[i]])
	}

	fmt.Println(scalenotes)
	fmt.Println(scalechords)

	pos, err := NotePosition(root)
	if err != nil {
		return nil, err
	}

	var chords []string

	for i := range Notes {
		note := Notes[(pos+i)%len(Notes)]
		if len(scalechords[note]) > 0 {
			chords = append(chords, fmt.Sprintf("%2s: %v", note, strings.Join(scalechords[note], ", ")))
		}
	}

	fmt.Println(chords)

	return scalechords, err
}
