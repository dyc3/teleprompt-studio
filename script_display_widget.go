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

	var b [][]*buffer.Cell

	// cell y positions and heights of chunks
	type cellmeta struct {
		y      int
		height int
	}
	chunkPos := map[int]cellmeta{}

	renderable := currentSession.Doc.GetRenderable()
	width := cvs.Area().Dx()
	chunkIdx := 0
	expandBuffer := func() {
		for cur.Y >= len(b) {
			row := make([]*buffer.Cell, width)
			for i := range row {
				row[i] = &buffer.Cell{
					Rune: ' ',
					Opts: cell.NewOptions(),
				}
			}
			b = append(b, row)
		}
		// log.Printf("expanded to: %d", len(b))
	}

	for _, r := range renderable {
		expandBuffer()

		switch t := r.(type) {
		case Header:
			cur.X = 0
			header := t
			cells := buffer.NewCells(
				header.Text,
				cell.FgColor(cell.ColorNumber(33)),
				cell.Bold(),
			)
			lim := clamp(width, 0, len(cells))
			for _, cell := range cells[:lim] {
				b[cur.Y][cur.X] = cell
				cur.X += 1
			}
			cur.Y += 1
			cur.X = 0
		case Chunk:
			chunk := t
			color := cell.ColorWhite
			if uint(chunkIdx) == selectedChunk {
				color = SELECT_COLOR
			}
			if len(chunk.Content) == 0 {
				continue
			}
			cells := buffer.NewCells(chunk.Content, cell.FgColor(color))
			cells = markdownFontModifiers(cells)
			wr, err := wrap.Cells(cells, width, wrap.AtWords)
			if err != nil {
				log.Printf("failed to word wrap chunk content: %s", err)
			}

			meta := cellmeta{
				y:      cur.Y,
				height: len(wr),
			}
			chunkPos[chunkIdx] = meta
			for _, line := range wr {
				expandBuffer()
				cur.X = 0
				for _, cell := range line {
					b[cur.Y][cur.X] = cell
					cur.X += 1
				}
				cur.Y += 1
			}
			cur.Y += 1

			chunkIdx++
		default:
			log.Printf("Unknown type %T", t)
		}
	}

	m := chunkPos[int(selectedChunk)]
	offsetY := clamp(m.y-(cvs.Area().Dy()/2)+(m.height/2), 0, len(b)-cvs.Area().Dy())
	c := image.Point{}
	cut := b[offsetY : offsetY+cvs.Area().Dy()]
	for _, line := range cut {
		c.X = 0
		for _, cell := range line {
			_, err := cvs.SetCell(c, cell.Rune, cell.Opts)
			if err != nil {
				log.Print(err)
			}
			c.X += 1
		}
		c.Y += 1
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
