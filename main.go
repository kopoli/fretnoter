package main

import (
	"fmt"
	"os"
	"strings"
)

func fault(err error, message string, arg ...string) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s%s: %s\n", message, strings.Join(arg, " "), err)
		os.Exit(1)
	}
}

func main() {
	var err error

	// scale, err := GetScale("D", "Natural Minor")
	// fault(err, "Getting the scale failed")
	// fmt.Println(scale)
	// chord, err := GetChord("C", "maj7")
	// fault(err, "Getting the chord failed")
	// fmt.Println(chord)

	err = GUIMain("v0")
	fault(err, "Running GUI failed")

	os.Exit(0)
}
