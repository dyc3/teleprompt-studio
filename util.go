package main

import (
	"fmt"
	"image"
	"time"

	"github.com/mum4k/termdash/cell"

	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/buffer"
)

func clamp(value int, min int, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

func valmap(x, in_min, in_max, out_min, out_max int) int {
	return (x-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
}

// TODO: alias time.Duration to our own type, and make this the String() function for that type
func Timestamp(t *time.Duration) string {
	return fmt.Sprintf("%02d:%02d:%02d.%03d", int32(t.Hours()), int32(t.Minutes())%60, int32(t.Seconds())%60, t.Milliseconds()%1000)
}

func DrawCells(cvs *canvas.Canvas, cells []*buffer.Cell, x, y int) {
	for i, c := range cells {
		cvs.SetCell(image.Point{
			X: x + i,
			Y: y,
		}, c.Rune, c.Opts)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func markdownFontModifiers(cells []*buffer.Cell) []*buffer.Cell {
	var mdcells []*buffer.Cell
	mode := 0
	stars := 0
	for _, c := range cells {
		if c.Rune == '`' {
			if mode == 0 {
				mode = 3
			} else {
				mode = 0
			}
			continue
		} else if c.Rune == '*' {
			stars++
			continue
		} else {
			if mode == 0 {
				mode = stars
			} else if stars == mode {
				mode = 0
			}
			stars = 0
		}

		if mode == 1 {
			c.Opts.Italic = true
		} else if mode == 2 {
			c.Opts.Bold = true
		} else if mode == 3 {
			c.Opts.Bold = true
			c.Opts.FgColor = cell.ColorNumber(57)
		}

		mdcells = append(mdcells, c)
	}
	return mdcells
}

func indexOfMaxInt32(arr []int32) int {
	var idx int
	var m int32
	for i, e := range arr {
		if i == 0 || e > m {
			idx = i
			m = e
		}
	}
	return idx
}
