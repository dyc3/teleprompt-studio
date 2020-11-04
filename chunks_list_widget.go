package main

import (
	"image"
	"sync"

	"github.com/mum4k/termdash/cell"

	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/buffer"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

type ChunkListWidget struct {
	mu sync.Mutex
}

func (w *ChunkListWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	cur := image.Point{
		X: 0,
		Y: 0,
	}

	width := cvs.Area().Dx()
	for _, header := range doc {
		cells := buffer.NewCells(header.Text)
		lim := clamp(width, 0, len(cells))
		for _, cell := range cells[:lim] {
			cvs.SetCell(cur, cell.Rune, cell.Opts)
			cur.X += 1
		}

		cur.Y += 1
		cur.X = 0

		for i, chunk := range header.Chunks {
			color := cell.ColorWhite
			if uint(i) == selectedChunk {
				color = cell.ColorYellow
			}

			cells := buffer.NewCells(chunk.Content, cell.FgColor(color))

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
	}
	return nil
}

func (w *ChunkListWidget) Keyboard(k *terminalapi.Keyboard) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return nil
}

func (w *ChunkListWidget) Mouse(m *terminalapi.Mouse) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return nil
}

func (w *ChunkListWidget) Options() widgetapi.Options {
	w.mu.Lock()
	defer w.mu.Unlock()

	return widgetapi.Options{}
}
