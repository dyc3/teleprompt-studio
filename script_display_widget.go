package main

import (
	"image"
	"log"
	"sync"

	"github.com/mum4k/termdash/cell"

	"github.com/mum4k/termdash/private/canvas/buffer"

	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/wrap"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

type ScriptDisplayWidget struct {
	mu sync.Mutex
}

func (w *ScriptDisplayWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	cur := image.Point{
		X: 0,
		Y: 0,
	}

	width := cvs.Area().Dx()
	for i, chunk := range loadedChunks {
		color := cell.ColorWhite
		if uint(i) == selectedChunk {
			color = cell.ColorYellow
		}
		wr, err := wrap.Cells(buffer.NewCells(chunk.Content, cell.FgColor(color)), width, wrap.AtWords)
		if err != nil {
			log.Printf("failed to word wrap chunk content: %s", err)
		}

		for _, line := range wr {
			cur.X = 0
			for _, cell := range line {
				cvs.SetCell(cur, cell.Rune, cell.Opts)
				cur.X += 1
			}
			cur.Y += 1
		}
		cur.Y += 1
	}
	return nil
}

func (w *ScriptDisplayWidget) Keyboard(k *terminalapi.Keyboard) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return nil
}

func (w *ScriptDisplayWidget) Mouse(m *terminalapi.Mouse) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return nil
}

func (w *ScriptDisplayWidget) Options() widgetapi.Options {
	w.mu.Lock()
	defer w.mu.Unlock()

	return widgetapi.Options{}
}
