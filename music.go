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
	"Major (Ionian)":           {2, 2, 1, 2, 2, 2, 1},
	"Dorian Mode":              {2, 1, 2, 2, 2, 1, 2},
	"Phrygian Mode":            {1, 2, 2, 2, 1, 2, 2},
	"Lydian Mode":              {2, 2, 2, 1, 2, 2, 1},
	"Mixolydian Mode":          {2, 2, 1, 2, 2, 1, 2},
	"Natural Minor (Aeolian)":  {2, 1, 2, 2, 1, 2, 2},
	"Locrian Mode":             {1, 2, 2, 1, 2, 2, 2},
	"Harmonic Minor":           {2, 1, 2, 2, 1, 3, 1},
	"Locrian nat6":             {1, 2, 2, 1, 3, 1, 2},
	"Ionian #5":                {2, 2, 1, 3, 1, 2, 1},
	"Ukranian minor":           {2, 1, 3, 1, 2, 1, 2},
	"Phrygian dominant":        {1, 3, 1, 2, 1, 2, 2},
	"Lydian #2":                {3, 1, 2, 1, 2, 2, 1},
	"Super Locrian diminished": {1, 2, 1, 2, 2, 1, 3},
	"Diminished":               {2, 1, 2, 1, 2, 1, 2, 1},
	"Dominant Diminished":      {1, 2, 1, 2, 1, 2, 1, 2},
	"Pentatonic Major":         {2, 2, 3, 2, 3},
	"Pentatonic Minor":         {3, 2, 2, 3, 2},
	"Metallica":                {1, 1, 1, 2, 1, 1, 1, 2, 2},
}

// Chords as the distance from the root note
var Chords = map[string][]int{
	"Major":      {0, 4, 7},
	"Minor":      {0, 3, 7},
	"Augmented":  {0, 4, 8},
	"Diminished": {0, 4, 6},
	"sus2":       {0, 2, 7},
	"sus4":       {0, 5, 7},
	"Power":      {0, 7},
	"7":          {0, 4, 7, 10},
	"m7":         {0, 3, 7, 10},
	"maj7":       {0, 4, 7, 11},
	"dom7":       {0, 4, 7, 10},
	"dim7":       {0, 3, 6, 9},
	"dom7f5":     {0, 4, 6, 10},
	"halfdim7":   {0, 3, 6, 10},
	"majdim7":    {0, 3, 6, 11},
	"minmaj7":    {0, 3, 7, 11},
	"augmaj7":    {0, 4, 8, 11},
	"aug7":       {0, 4, 8, 10},
	"7sus2":      {0, 5, 7, 10},
	"9":          {0, 4, 7, 10, 14},
	"m9":         {0, 3, 7, 10, 14},
	"maj9":       {0, 4, 7, 11, 14},
	"11":         {0, 4, 7, 10, 14, 17},
	"m11":        {0, 3, 7, 10, 14, 17},
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

	return scalechords, err
}

func DetectChord(chord string) ([]string, error) {
	notes := strings.Fields(chord)

	if len(notes) == 0 {
		return []string{}, fmt.Errorf("No notes detected")
	}

	// Convert notes into positions
	pos := make([]int, len(notes))
	for i := range notes {
		p, err := NotePosition(notes[i])
		if err != nil {
			return []string{}, err
		}

		pos[i] = p
	}

	// Modify the positions relative to root
	root := pos[0]
	for i := range pos {
		pos[i] -= root
		if pos[i] < 0 {
			pos[i] += len(Notes)
		}
	}

	// Skip root note
	pos = pos[1:]

	// Get all the chords the notes are present in
	chords := []string{}
	for k, v := range Chords {
		if len(pos) > len(v) {
			continue
		}

		// Check that all given notes are present in the chord
		found := false
		for i := range pos {
			found = false
			for j := range v {
				if pos[i] == v[j] {
					found = true
					break
				}
			}

			if !found {
				break
			}
		}
		if !found {
			continue
		}

		// Print out all the chords
		if found {
			msg := notes[0] + k
			if len(pos) < len(v) {
				n, _ := GetChord(notes[0], k)
				desc := ""
				if len(n) > len(notes) {
					desc = "partial "
				}
				msg = fmt.Sprintf("%s (%s%s)", msg, desc, strings.Join(n, " "))
			}
			chords = append(chords, msg)
		}
	}

	sort.Strings(chords)

	return chords, nil
}
