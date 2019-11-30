package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/OpenPeeDeeP/xdg"
)

const (
	TypeScale = "Scale"
	TypeChord = "Chord"
)

type BoardState struct {
	Name   string
	Type   string
	Root   string
	Tuning string
}

type State struct {
	Columns int
	Tuning  string

	Boards []BoardState
}

func getSaveFilePath() (string, error) {
	path := xdg.New("", "fretnoter").DataHome()
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(path, "config.json"), nil
}

func Save(s *State) error {
	data, err := json.Marshal(*s)
	if err != nil {
		return err
	}

	path, err := getSaveFilePath()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)
}

func Load() (*State, error) {
	path, err := getSaveFilePath()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ret State
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}

	fmt.Println("Loaded configuration from:", path)

	return &ret, nil
}
