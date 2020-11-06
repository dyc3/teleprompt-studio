package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"

	"github.com/mum4k/termdash/mouse"

	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/canvas/buffer"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

type AudioDisplayWidget struct {
	mu sync.Mutex

	// The time interval in which to display the waveform
	window          TimeSpan
	stickToEnd      bool
	selected        TimeSpan
	selectionActive bool
	lastClickStart  image.Point

	area         image.Rectangle
	waitingFrame int
}

func (w *AudioDisplayWidget) animateWaiting() {
	for len(recordedAudio) == 0 {
		w.waitingFrame++
		if w.waitingFrame >= len(waitingAnimation) {
			w.waitingFrame = 0
		}
		time.Sleep(300 * time.Millisecond)
	}
}

var waitingAnimation = []rune("▞▚")

func (w *AudioDisplayWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.area = cvs.Area()

	if len(recordedAudio) == 0 {
		cells := buffer.NewCells(string(waitingAnimation[w.waitingFrame]) + " Waiting for audio...")
		x, y := (w.area.Dx()/2)-(len(cells)/2), w.area.Dy()/2
		for i, c := range cells {
			cvs.SetCell(image.Point{
				X: x + i,
				Y: y,
			}, c.Rune, c.Opts)
		}

		return nil
	}

	if w.stickToEnd {
		recorded := samplesToDuration(sampleRate, len(recordedAudio))
		diff := recorded - w.window.End
		w.window.End += diff
		w.window.Start += diff
		if w.window.Duration() < time.Second {
			if recorded > time.Second {
				w.window.Start = w.window.End - time.Second
			} else {
				w.window.Start = w.window.End - recorded
			}
		}
	}

	start := durationToSamples(sampleRate, w.window.Start)
	start = clamp(start, 0, len(recordedAudio)-1)
	end := durationToSamples(sampleRate, w.window.End)
	end = clamp(end, 0, len(recordedAudio)-1)
	samples := recordedAudio[start:end]
	bc, err := braille.New(w.area)
	if err != nil {
		return err
	}

	for i := 1; i < len(samples); i++ {
		a := bc.Area()
		startX := valmap(i-1, 0, len(samples), 0, a.Dx())
		startY := valmap(int(samples[i-1]), math.MinInt32, math.MaxInt32, a.Dy(), 0)
		endX := valmap(i, 0, len(samples), 0, a.Dx())
		endY := valmap(int(samples[i]), math.MinInt32, math.MaxInt32, a.Dy(), 0)
		color := cell.ColorWhite
		if w.selectionActive {
			if start+i >= durationToSamples(sampleRate, w.selected.Start) && start+i <= durationToSamples(sampleRate, w.selected.End) {
				color = cell.ColorYellow
			}
		}
		err := draw.BrailleLine(bc,
			image.Point{startX, startY},
			image.Point{endX, endY},
			draw.BrailleLineCellOpts(cell.FgColor(color)),
		)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	if err := bc.CopyTo(cvs); err != nil {
		log.Printf("Copy failed: %s", err)
		return err
	}

	cells := buffer.NewCells(fmt.Sprintf("%s", Timestamp(&w.window.Start)))
	x, y := 0, w.area.Dy()-1
	DrawCells(cvs, cells, x, y)

	cells = buffer.NewCells(fmt.Sprintf("%s", Timestamp(&w.window.End)))
	x, y = w.area.Dx()-len(cells), w.area.Dy()-1
	DrawCells(cvs, cells, x, y)

	cells = buffer.NewCells(fmt.Sprintf("%v (%d) [%d:%d] %v", w.selected, len(recordedAudio), start, end, w.area))
	x, y = (w.area.Dx()/2)-(len(cells)/2), 0
	DrawCells(cvs, cells, x, y)

	cells = buffer.NewCells(fmt.Sprintf("%d <= %d <= %d", durationToSamples(sampleRate, w.selected.Start), start, durationToSamples(sampleRate, w.selected.End)))
	x, y = (w.area.Dx()/2)-(len(cells)/2), 1
	DrawCells(cvs, cells, x, y)

	return nil
}

func (w *AudioDisplayWidget) Keyboard(k *terminalapi.Keyboard) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return nil
}

func (w *AudioDisplayWidget) Mouse(m *terminalapi.Mouse) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if m.Button == mouse.ButtonRight {
		w.selectionActive = false
	} else if m.Button == mouse.ButtonLeft {
		if w.selectionActive {
			w.selected.End = mousePointToTimestampOffset(m.Position, w.area, w.window)
		} else {
			w.selectionActive = true
			w.selected = TimeSpan{
				Start: mousePointToTimestampOffset(m.Position, w.area, w.window),
			}
			log.Printf("drag select start %s", w.selected.Start)
			w.lastClickStart = m.Position
		}
	}

	if w.selectionActive && m.Button == mouse.ButtonRelease {
		if m.Position == w.lastClickStart {
			w.selectionActive = false
		}
		w.selected.End = mousePointToTimestampOffset(m.Position, w.area, w.window)
		log.Printf("drag select end %s", w.selected.End)
	}

	return nil
}

func (w *AudioDisplayWidget) Options() widgetapi.Options {
	w.mu.Lock()
	defer w.mu.Unlock()

	return widgetapi.Options{
		WantMouse: widgetapi.MouseScopeWidget,
	}
}

func mousePointToTimestampOffset(p image.Point, area image.Rectangle, window TimeSpan) time.Duration {
	return window.Start + window.Duration()*time.Duration(p.X)/time.Duration(area.Dx())
}
