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

type Point struct {
	X float32
	Y float32
}

type Window struct {
	X float32
	Y float32
	W float32
	H float32
}

func NewWindow(x, y, w, h float32) Window {
	return Window{x, y, w, h}
}

func (n Window) Inner(l, t, r, b float32) Window {
	return Window{n.X + l, n.Y + t, n.W - r - l, n.H - t - b}
}

func (n Window) SplitH(t float32) (top Window, bottom Window) {
	top = Window{n.X, n.Y, n.W, t - 1}
	bottom = Window{n.X, n.Y + t, n.W, n.H - t}
	return
}

func (n Window) SplitV(l float32) (left Window, right Window) {
	left = Window{n.X, n.Y, l - 1, n.H}
	right = Window{l, n.Y, n.W - l, n.H}
	return
}

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

func (u *UI) Fill() {
	nvg.Fill(u.vg)
}

func (u *UI) Stroke() {
	nvg.Stroke(u.vg)
}

func (u *UI) Rect(w Window) {
	nvg.BeginPath(u.vg)
	nvg.Rect(u.vg, w.X, w.Y, w.W, w.H)
}

func (u *UI) RoundedRect(w Window, radius float32) {
	nvg.BeginPath(u.vg)
	nvg.RoundedRect(u.vg, w.X, w.Y, w.W, w.H, radius)
}

func (u *UI) Circle(cx, cy, r float32) {
	nvg.BeginPath(u.vg)
	nvg.Circle(u.vg, cx, cy, r)
}

func (u *UI) Text(w Window, size float32, align int32, string string) {
	nvg.FontSize(u.vg, size)
	nvg.TextAlign(u.vg, align)
	nvg.Text(u.vg, w.X, w.Y, string)
}
