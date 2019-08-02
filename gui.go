package main

import (
	"fmt"
	"image"

	"github.com/aarzilli/nucular"
	// "github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/style"
)

type FretUI struct {
}

func (f *FretUI) update(w *nucular.Window) {
	w.Row(25).Static(100)
	w.Label("helloworld", "LC")
}

func GUIMain(version string) error {

	fu := &FretUI{}

	title := fmt.Sprintf("Fretnoter %s", version)
	w := nucular.NewMasterWindowSize(0, title, image.Point{640, 630}, fu.update)

	w.SetStyle(style.FromTheme(style.DarkTheme, 1.0))

	w.Main()
	return nil
}
