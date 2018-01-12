package main

import (
	"github.com/JamesDunne/golang-nanovg/nvg"
)

type UIPalette [4]nvg.Color
type PaletteIndex int

var (
	// Palettes are ordered from darkest to lightest shades:
	// This palette from https://flatuicolors.com/
	palette = UIPalette{
		nvg.RGB(44, 62, 80),    // midnight blue
		nvg.RGB(52, 73, 94),    // wet asphalt
		nvg.RGB(127, 140, 141), // asbestos
		nvg.RGB(149, 165, 166), // concrete
	}
)

type UI struct {
	vg *nvg.Context
	p  [4]nvg.Color
}

func NewUI(vg *nvg.Context) *UI {
	return &UI{vg: vg, p: palette}
}

func (u *UI) Palette(p PaletteIndex) nvg.Color {
	return u.p[p]
}

func (u *UI) FillColor(c nvg.Color) {
	nvg.FillColor(u.vg, c)
}

func (u *UI) StrokeColor(c nvg.Color) {
	nvg.StrokeColor(u.vg, c)
}

func (u *UI) StrokeWidth(size float32) {
	nvg.StrokeWidth(u.vg, size)
}

func (u *UI) BeginPath() {
	nvg.BeginPath(u.vg)
}

func (u *UI) Fill() {
	nvg.Fill(u.vg)
}

func (u *UI) Stroke() {
	nvg.Stroke(u.vg)
}

func (u *UI) Rect(w Window) {
	nvg.Rect(u.vg, w.X, w.Y, w.W, w.H)
}

func (u *UI) RoundedRect(w Window, radius float32) {
	nvg.RoundedRect(u.vg, w.X, w.Y, w.W, w.H, radius)
}

func (u *UI) Circle(cx, cy, r float32) {
	nvg.Circle(u.vg, cx, cy, r)
}

func (u *UI) TextPoint(p Point, size float32, align int32, string string) {
	nvg.FontSize(u.vg, size)
	nvg.TextAlign(u.vg, align)
	nvg.Text(u.vg, p.X, p.Y, string)
}

func (u *UI) Text(w Window, size float32, align int32, string string) {
	u.TextPoint(w.AlignedPoint(align), size, align, string)
}
