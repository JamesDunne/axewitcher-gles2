package main

import (
	"github.com/JamesDunne/golang-nanovg/nvg"
)

var (
	// Palettes are ordered from darkest to lightest shades:
	// This palette from https://flatuicolors.com/
	palette = [...]nvg.Color{
		nvg.RGB(44, 62, 80),    // midnight blue
		nvg.RGB(52, 73, 94),    // wet asphalt
		nvg.RGB(127, 140, 141), // asbestos
		nvg.RGB(149, 165, 166), // concrete
	}
)

type Window struct {
	x float32
	y float32
	w float32
	h float32
}

func NewWindow(x, y, w, h float32) Window {
	return Window{x, y, w, h}
}

type UI struct {
	vg        *nvg.Context
	container Window
}

func NewUI(vg *nvg.Context) *UI {
	return &UI{vg: vg}
}

// TODO: use recursion to nest windows within windows.
func (u *UI) Draw(container Window) {
	u.container = container
}
