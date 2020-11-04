package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mum4k/termdash/cell"

	"github.com/gordonklaus/portaudio"
	"github.com/zimmski/osutil"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

const ROOTID = "root"

var selectedChunk uint
var selectedTake int
var doc Document

var scriptWidget *ScriptDisplayWidget
var waveformWidget *linechart.LineChart
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

	waveformWidget, err = linechart.New(
		linechart.XAxisUnscaled(),
		linechart.YAxisFormattedValues(IgnoreValueFormatter),
	)
	if err != nil {
		log.Fatal(err)
	}

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

func record() {
	var err error
	osutil.CaptureWithCGo(func() {
		err = portaudio.Initialize()
	})
	if err != nil {
		log.Fatalf("Failed to initialize recording: %s", err)
	}
	defer portaudio.Terminate()
	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	if err != nil {
		log.Fatalf("Failed to open stream audio: %s", err)
	}
	defer stream.Close()

	err = stream.Start()
	if err != nil {
		log.Fatalf("Failed to start stream audio: %s", err)
	}

	samples := make([]float64, 64)
	for {
		err := stream.Read()
		if err != nil {
			log.Fatalf("Failed to read stream audio: %s", err)
		}

		for i, s := range in {
			samples[i] = float64(s)
		}

		waveformWidget.Series("Waveform", samples)
	}
}

func buildControlsDisplay() {
	keybinds := getKeybinds()
	for _, bind := range keybinds {
		controlsWidget.Write(fmt.Sprintf("%s", bind.key), text.WriteCellOpts(cell.BgColor(cell.ColorWhite), cell.FgColor(cell.ColorBlack)))
		controlsWidget.Write(fmt.Sprintf(" %s  ", bind.desc))
	}
}

type Take struct {
	Mark TakeMark
}

func main() {
	f, err := os.Create("debug.log")
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	scriptFile := flag.String("script", "", "Path to the markdown file to use as input.")
	flag.Parse()

	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()
	log.Print("Building layout")
	c := buildLayout(t)

	ctx, cancel := context.WithCancel(context.Background())

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			t.Close()
			cancel()
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

	buildControlsDisplay()

	log.Print("Reading script")
	err = readScript(*scriptFile)
	if err != nil {
		t.Close()
		log.Fatalf("Failed to open file %s: %s", *scriptFile, err)
	}

	go record()

	log.Print("Running termdash")
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(10*time.Millisecond)); err != nil {
		log.Fatalf("%s", err)
	}
}
