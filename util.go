package main

import (
	"fmt"
	"image"
	"time"

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
