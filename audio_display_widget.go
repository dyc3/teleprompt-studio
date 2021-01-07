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
	dragging        bool
	lastClickStart  image.Point
	showDebug       bool
	recordStart     time.Time

	area         image.Rectangle
	waitingFrame int
}

func (w *AudioDisplayWidget) animateWaiting() {
	for len(currentSession.Audio) == 0 {
		w.waitingFrame++
		if w.waitingFrame >= len(waitingAnimation) {
			w.waitingFrame = 0
		}
		time.Sleep(300 * time.Millisecond)
	}
	w.recordStart = time.Now()
}

var waitingAnimation = []rune("▞▚")

func (w *AudioDisplayWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.area = cvs.Area()

	if len(currentSession.Audio) == 0 {
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
		recorded := samplesToDuration(sampleRate, len(currentSession.Audio))
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
	start = clamp(start, 0, len(currentSession.Audio)-1)
	end := durationToSamples(sampleRate, w.window.End)
	end = clamp(end, 0, len(currentSession.Audio)-1)
	samples := currentSession.Audio[start:end]
	bc, err := braille.New(w.area)
	if err != nil {
		return err
	}
	var takes []Take
	for _, t := range currentSession.Doc.GetAllTakes() {
		if (t.End < w.window.End && t.End > w.window.Start) || (t.Start > w.window.Start && t.Start < w.window.End) {
			takes = append(takes, t)
		}
	}

	chunk_length := len(samples) / bc.Area().Dx()
	for x := 0; x < bc.Area().Dx(); x++ {
		cStart, cEnd := x*chunk_length, (x+1)*chunk_length
		chunk := samples[cStart:cEnd]
		max, min := int32(0), int32(0)
		for _, s := range chunk {
			if s > max {
				max = s
			} else if s < min {
				min = s
			}
		}

		color := cell.ColorWhite
		for _, t := range takes {
			if start+cStart >= durationToSamples(sampleRate, t.Start) && start+cEnd <= durationToSamples(sampleRate, t.End) {
				if t.Mark == Good {
					color = GOOD_COLOR
				} else if t.Mark == Bad {
					color = BAD_COLOR
				} else if t.Mark == Sync {
					color = SYNC_COLOR
				} else {
					color = cell.ColorNumber(33)
				}
			}
		}
		if w.selectionActive {
			if start+cStart >= durationToSamples(sampleRate, w.selected.Start) && start+cEnd <= durationToSamples(sampleRate, w.selected.End) {
				color = SELECT_COLOR
			}
		}

		maxY := valmap(int(max), math.MinInt32, math.MaxInt32, bc.Area().Dy(), 0)
		minY := valmap(int(min), math.MinInt32, math.MaxInt32, bc.Area().Dy(), 0)

		err := draw.BrailleLine(bc,
			image.Point{x, maxY},
			image.Point{x, minY},
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

	d := samplesToDuration(sampleRate, len(currentSession.Audio))
	cells = buffer.NewCells(fmt.Sprintf("Recorded: %s", Timestamp(&d)))
	x, y = 0, 0
	DrawCells(cvs, cells, x, y)

	if w.showDebug {
		real := time.Now().Sub(w.recordStart)
		cells = buffer.NewCells(fmt.Sprintf("Real time: %s", Timestamp(&real)))
		x, y = 0, 1
		DrawCells(cvs, cells, x, y)

		diff := real - d
		var opts []cell.Option
		if diff > time.Second {
			opts = append(opts, cell.FgColor(BAD_COLOR))
		} else if diff > 500*time.Millisecond {
			opts = append(opts, cell.FgColor(cell.ColorYellow))
		}
		cells = buffer.NewCells(
			fmt.Sprintf("Diff: %s", Timestamp(&diff)),
			opts...,
		)
		x, y = 0, 2
		DrawCells(cvs, cells, x, y)
	}

	lowerMidY := w.area.Dy() / 4 * 3
	if isPlaying {
		d := samplesToDuration(sampleRate, playbackPosition)
		cells = buffer.NewCells(fmt.Sprintf("%s", Timestamp(&d)))
		x, y = (w.area.Dx()/2)-(len(cells)/2), lowerMidY
		DrawCells(cvs, cells, x, y)
	}

	if w.selectionActive {
		d := w.selected.Duration()
		cells = buffer.NewCells(fmt.Sprintf("%s", Timestamp(&d)))
		x, y = (w.area.Dx()/2)-(len(cells)/2), lowerMidY+1
		DrawCells(cvs, cells, x, y)
	}

	if w.showDebug {
		cells = buffer.NewCells(fmt.Sprintf("%v (%d) [%d:%d] %v", w.selected, len(currentSession.Audio), start, end, w.area))
		x, y = (w.area.Dx()/2)-(len(cells)/2), 0
		DrawCells(cvs, cells, x, y)

		cells = buffer.NewCells(fmt.Sprintf("%d <= %d <= %d", durationToSamples(sampleRate, w.selected.Start), start, durationToSamples(sampleRate, w.selected.End)))
		x, y = (w.area.Dx()/2)-(len(cells)/2), 1
		DrawCells(cvs, cells, x, y)
	}

	if isPlaying {
		playbackMarkerX := -1
		chunk_length := len(samples) / cvs.Area().Dx()
		for x := 0; x < cvs.Area().Dx(); x++ {
			cStart, cEnd := start+(x*chunk_length), start+((x+1)*chunk_length)
			if cStart <= playbackPosition && playbackPosition <= cEnd {
				playbackMarkerX = x
			}
		}
		cvs.SetAreaCellOpts(
			image.Rect(playbackMarkerX, 0, playbackMarkerX+1, cvs.Area().Dy()),
			cell.BgColor(cell.ColorNumber(56)),
		)
	}

	if currentSession.Doc.SyncOffset >= w.window.Start && currentSession.Doc.SyncOffset <= w.window.End {
		syncOffsetX := timestampOffsetToX(currentSession.Doc.SyncOffset, w.area, w.window)

		if w.showDebug {
			cells = buffer.NewCells(fmt.Sprintf("sync offset: %v (x=%d)", currentSession.Doc.SyncOffset, syncOffsetX))
			x, y = (w.area.Dx()/2)-(len(cells)/2), 2
			DrawCells(cvs, cells, x, y)
		}

		cvs.SetAreaCellOpts(
			image.Rect(syncOffsetX, 2, syncOffsetX+1, cvs.Area().Dy()-4),
			cell.BgColor(SYNC_OFFSET_COLOR),
		)
	}

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
			if w.dragging {
				startPoint := mousePointToTimestampOffset(w.lastClickStart, w.area, w.window)
				dragPoint := mousePointToTimestampOffset(m.Position, w.area, w.window)
				if m.Position.X < w.lastClickStart.X {
					w.selected.Start = dragPoint
					w.selected.End = startPoint
				} else {
					w.selected.End = dragPoint
				}
			} else {
				log.Printf("clearing selection")
				w.selectionActive = false
			}
		} else if !w.dragging {
			w.selectionActive = true
			w.dragging = true
			w.selected = TimeSpan{
				Start: mousePointToTimestampOffset(m.Position, w.area, w.window),
			}
			log.Printf("drag select start %s", w.selected.Start)
			w.lastClickStart = m.Position
		}
	} else if m.Button == mouse.ButtonWheelDown {
		x := w.window.Duration() / 10
		w.window.Start -= x
		w.window.End += x
		if w.window.Start < 0 {
			w.window.Start = 0
		}
		if w.window.End > samplesToDuration(sampleRate, len(currentSession.Audio)) {
			w.window.End = samplesToDuration(sampleRate, len(currentSession.Audio))
		}
	} else if m.Button == mouse.ButtonWheelUp {
		x := w.window.Duration() / 10
		w.window.Start += x
		w.window.End -= x
	} else if m.Button == mouse.ButtonMiddle {
		if !w.dragging {
			w.dragging = true
			w.lastClickStart = m.Position
		}

		if w.lastClickStart.X != m.Position.X {
			diff := clamp(m.Position.X-w.lastClickStart.X, -1, 1)
			pixel_length := w.window.Duration() / time.Duration(w.area.Dx())
			time_diff := -time.Duration(diff) * pixel_length
			if w.window.Start+time_diff > 0 && w.window.End+time_diff < samplesToDuration(sampleRate, len(currentSession.Audio)) {
				w.window.Start += time_diff
				w.window.End += time_diff
			}
			w.lastClickStart = m.Position
		}

	} else if m.Button == mouse.ButtonRelease {
		if w.selectionActive && w.dragging {
			w.dragging = false
			if m.Position == w.lastClickStart {
				w.selectionActive = false
			}
			// w.selected.End = mousePointToTimestampOffset(m.Position, w.area, w.window)
			log.Printf("drag select end %s", w.selected.End)
		} else if w.dragging {
			w.dragging = false
		}
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

func timestampOffsetToX(timestamp time.Duration, area image.Rectangle, window TimeSpan) int {
	return int(timestamp-window.Start) * area.Dx() / int(window.Duration())
}

func (w *AudioDisplayWidget) Deselect() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.selectionActive = false
}
