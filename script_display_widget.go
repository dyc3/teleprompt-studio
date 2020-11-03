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
	mu            sync.Mutex
	chunks        *[]Chunk
	selectedChunk uint
}

func (w *ScriptDisplayWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	cur := image.Point{
		X: 0,
		Y: 0,
	}

	width := cvs.Area().Dx()
	for i, chunk := range *w.chunks {
		wr, err := wrap.Cells(chunk.GetCellBuffer(uint(i) == selectedChunk), width, wrap.AtWords)
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

func (w *ScriptDisplayWidget) SetChunks(chunks []Chunk) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.chunks = &chunks
}

func (w *ScriptDisplayWidget) SelectChunk(index uint) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if index >= uint(len(*w.chunks)) {
		index = uint(len(*w.chunks) - 1)
	}
	w.selectedChunk = index
}

func (c *Chunk) GetCellBuffer(isSelected bool) []*buffer.Cell {
	buf := []*buffer.Cell{}
	for _, r := range c.Content {
		color := cell.ColorWhite
		if isSelected {
			color = cell.ColorYellow
		}
		cell := buffer.NewCell(r, cell.FgColor(color))
		buf = append(buf, cell)
	}
	return buf
}
