package main

import (
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/cell"

	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/buffer"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

type TakeListWidget struct {
	mu sync.Mutex
}

func (w *TakeListWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	cur := image.Point{
		X: 0,
		Y: 0,
	}

	width := cvs.Area().Dx()

	for i, Take := range currentSession.Doc.GetChunk(int(selectedChunk)).Takes {
		color := cell.ColorWhite
		symbolRune := ' '

		if Take.Mark == Good {
			color = cell.ColorGreen
			symbolRune = '✓'
		} else if Take.Mark == Bad {
			color = cell.ColorRed
			symbolRune = '✗'
		}

		symbol := buffer.NewCell(symbolRune, cell.FgColor(color))

		if i == selectedTake {
			color = cell.ColorYellow
		}

		cells := buffer.NewCells(fmt.Sprintf("Take %d", i), cell.FgColor(color))

		header := []*buffer.Cell{
			buffer.NewCell('[', cell.FgColor(cell.ColorWhite)),
			symbol,
			buffer.NewCell(']', cell.FgColor(cell.ColorWhite)),
		}

		for _, cell := range header {
			cvs.SetCell(cur, cell.Rune, cell.Opts)
			cur.X += 1
		}

		cur.X += 1
		lim := clamp(width-cur.X, cur.X, len(cells))
		if lim < 0 {
			lim = 0
		}
		for _, cell := range cells[:lim] {
			cvs.SetCell(cur, cell.Rune, cell.Opts)
			cur.X += 1
		}
		cur.Y += 1
		cur.X = 0
	}
	return nil
}

func (w *TakeListWidget) Keyboard(k *terminalapi.Keyboard) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return nil
}

func (w *TakeListWidget) Mouse(m *terminalapi.Mouse) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return nil
}

func (w *TakeListWidget) Options() widgetapi.Options {
	w.mu.Lock()
	defer w.mu.Unlock()

	return widgetapi.Options{}
}
