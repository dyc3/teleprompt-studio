package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mum4k/termdash/cell"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

const ROOTID = "root"

var selectedChunk uint
var selectedTake int
var doc Document

var scriptWidget *ScriptDisplayWidget
var waveformWidget *AudioDisplayWidget
var chunksWidget *ChunkListWidget
var controlsWidget *text.Text
var takesWidget *TakeListWidget

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

type keybind struct {
	key  keyboard.Key
	desc string
}

func getKeybinds() []keybind {
	return []keybind{
		{
			key:  keyboard.KeyArrowDown,
			desc: "Next Chunk",
		},
		{
			key:  keyboard.KeyArrowUp,
			desc: "Previous Chunk",
		},
		{
			key:  ' ',
			desc: "Start Take",
		},
		{
			key:  'g',
			desc: "Mark Good",
		},
		{
			key:  'b',
			desc: "Mark Bad",
		},
	}
}

func IgnoreValueFormatter(value float64) string {
	return ""
}

func buildLayout(t *termbox.Terminal) *container.Container {
	root, err := container.New(t, container.ID(ROOTID))
	if err != nil {
		log.Fatal(err)
	}

	scriptWidget = &ScriptDisplayWidget{}
	if err != nil {
		log.Fatal(err)
	}

	waveformWidget = &AudioDisplayWidget{}
	waveformWidget.stickToEnd = true
	go waveformWidget.animateWaiting()

	chunksWidget = &ChunkListWidget{}
	if err != nil {
		log.Fatal(err)
	}

	controlsWidget, err = text.New()
	if err != nil {
		log.Fatal(err)
	}

	takesWidget = &TakeListWidget{}

	builder := grid.New()
	builder.Add(
		grid.ColWidthPerc(80,
			grid.RowHeightPerc(50,
				grid.Widget(scriptWidget,
					container.Border(linestyle.Light),
					container.BorderTitle("Script"),
				),
			),
			grid.RowHeightPerc(40,
				grid.Widget(waveformWidget,
					container.Border(linestyle.Light),
					container.BorderTitle("Audio"),
				),
			),
			grid.RowHeightFixed(3,
				grid.Widget(controlsWidget),
			),
		),
	)

	builder.Add(
		grid.ColWidthPerc(20,
			grid.RowHeightPerc(50,
				grid.Widget(chunksWidget,
					container.Border(linestyle.Light),
					container.BorderTitle("Chunks"),
				),
			),
			grid.RowHeightPerc(50,
				grid.Widget(takesWidget,
					container.Border(linestyle.Light),
					container.BorderTitle("Takes"),
				),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	if err := root.Update(ROOTID, gridOpts...); err != nil {
		log.Fatal(err)
	}

	return root
}

func updateControlsDisplay() {
	controlsWidget.Reset()
	keybinds := getKeybinds()
	for _, bind := range keybinds {
		controlsWidget.Write(fmt.Sprintf("%s", bind.key), text.WriteCellOpts(cell.BgColor(cell.ColorWhite), cell.FgColor(cell.ColorBlack)))
		controlsWidget.Write(fmt.Sprintf(" %s  ", bind.desc))
	}
}

type Take struct {
	TimeSpan
	Mark TakeMark
}

func globalKeyboardHandler(k *terminalapi.Keyboard) {
	if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
		terminal.Close()
		cancelGlobal()
	} else if k.Key == keyboard.KeyArrowUp {
		if selectedChunk > 0 {
			selectedChunk -= 1
		}
		chunk := doc.GetChunk(int(selectedChunk))
		selectedTake = len(chunk.Takes) - 1
	} else if k.Key == keyboard.KeyArrowDown {
		if selectedChunk < uint(doc.CountChunks()-1) {
			selectedChunk += 1
		}
		chunk := doc.GetChunk(int(selectedChunk))
		selectedTake = len(chunk.Takes) - 1
	} else if k.Key == ' ' {
		chunk := doc.GetChunk(int(selectedChunk))
		chunk.Takes = append(chunk.Takes, Take{})
		selectedTake = len(chunk.Takes) - 1
	} else if k.Key == 'g' {
		doc.GetChunk(int(selectedChunk)).Takes[selectedTake].Mark = Good
	} else if k.Key == 'b' {
		doc.GetChunk(int(selectedChunk)).Takes[selectedTake].Mark = Bad
	} else {
		log.Printf("Unknown key pressed: %v", k)
	}
}

var terminal *termbox.Terminal
var ctxGlobal context.Context
var cancelGlobal context.CancelFunc

func main() {
	f, err := os.Create("debug.log")
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	scriptFile := flag.String("script", "", "Path to the markdown file to use as input.")
	flag.Parse()

	terminal, err = termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		log.Fatal(err)
	}
	defer terminal.Close()
	log.Print("Building layout")
	c := buildLayout(terminal)

	ctxGlobal, cancelGlobal = context.WithCancel(context.Background())

	updateControlsDisplay()

	log.Print("Reading script")
	err = readScript(*scriptFile)
	if err != nil {
		terminal.Close()
		log.Fatalf("Failed to open file %s: %s", *scriptFile, err)
	}

	go record()
	go audioProcessor()

	log.Print("Running termdash")
	if err := termdash.Run(ctxGlobal, terminal, c, termdash.KeyboardSubscriber(globalKeyboardHandler), termdash.RedrawInterval(10*time.Millisecond)); err != nil {
		log.Fatalf("%s", err)
	}
}
