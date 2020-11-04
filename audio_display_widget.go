package main

import (
	"fmt"
	"image"
	"sync"
	"time"

	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/buffer"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

type AudioDisplayWidget struct {
	mu sync.Mutex

	// The time interval in which to display the waveform
	window TimeSpan

	area image.Rectangle

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
		a := cvs.Area()
		x, y := (a.Dx()/2)-(len(cells)/2), a.Dy()/2
		for i, c := range cells {
			cvs.SetCell(image.Point{
				X: x + i,
				Y: y,
			}, c.Rune, c.Opts)
		}

		return nil
	}

	cells := buffer.NewCells(fmt.Sprintf("%d", len(recordedAudio)))
	a := cvs.Area()
	x, y := (a.Dx()/2)-(len(cells)/2), a.Dy()/2
	for i, c := range cells {
		cvs.SetCell(image.Point{
			X: x + i,
			Y: y,
		}, c.Rune, c.Opts)
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

	return nil
}

func (w *AudioDisplayWidget) Options() widgetapi.Options {
	w.mu.Lock()
	defer w.mu.Unlock()

	return widgetapi.Options{}
}

func mousePointToTimestampOffset(p image.Point, area image.Rectangle, window TimeSpan) time.Duration {
	return window.Start + window.Duration()*time.Duration(float32(p.X)/float32(area.Dx()))
}
