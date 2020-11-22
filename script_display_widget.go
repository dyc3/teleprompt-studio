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

// cell y positions and heights of chunks
type cellmeta struct {
	y      int
	height int
}

func scriptRenderBuffer(width int) ([][]*buffer.Cell, map[int]cellmeta) {
	var b [][]*buffer.Cell

	cur := image.Point{
		X: 0,
		Y: 0,
	}

	chunkPos := map[int]cellmeta{}

	renderable := currentSession.Doc.GetRenderable()
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

	return b, chunkPos
}

func (w *ScriptDisplayWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	width := cvs.Area().Dx() - 2
	b, chunkPos := scriptRenderBuffer(width)

	m := chunkPos[int(selectedChunk)]
	absY := (cvs.Area().Dy() / 2) - (m.height / 2)
	offsetY := clamp(m.y-absY, 0, len(b)-cvs.Area().Dy())
	c := image.Point{}
	cut := b[offsetY : offsetY+cvs.Area().Dy()]
	cvs.SetCell(image.Point{X: 0, Y: m.y - offsetY}, '>', cell.FgColor(SELECT_COLOR))

	if m.height > 1 {
		for i := 0; i < m.height; i++ {
			var r rune
			switch i {
			case 0:
				r = '┌'
			case m.height - 1:
				r = '└'
			default:
				r = '│'
			}
			cvs.SetCell(image.Point{X: 1, Y: m.y + i - offsetY}, r, cell.FgColor(SELECT_COLOR))
		}
	} else {
		cvs.SetCell(image.Point{X: 1, Y: m.y - offsetY}, '═', cell.FgColor(SELECT_COLOR))
	}

	for _, line := range cut {
		c.X = 2
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
