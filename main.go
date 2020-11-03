package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

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
var loadedChunks []Chunk

var scriptWidget *ScriptDisplayWidget
var waveformWidget *linechart.LineChart
var chunksWidget *text.Text

func IgnoreValueFormatter(value float64) string {
	return ""
}

func buildLayout(t *termbox.Terminal) *container.Container {
	root, err := container.New(t, container.ID(ROOTID))
	if err != nil {
		log.Fatal(err)
	}

	helloWidget, err := text.New(
		text.WrapAtWords(),
	)
	if err != nil {
		log.Fatal(err)
	}
	helloWidget.Write("Hello")

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

	chunksWidget, err = text.New()
	if err != nil {
		log.Fatal(err)
	}

	builder := grid.New()
	builder.Add(
		grid.ColWidthPerc(80,
			grid.RowHeightPerc(50,
				grid.Widget(scriptWidget,
					container.Border(linestyle.Light),
					container.BorderTitle("Script"),
				),
			),
			grid.RowHeightPerc(45,
				grid.Widget(waveformWidget,
					container.Border(linestyle.Light),
					container.BorderTitle("Audio"),
				),
			),
			grid.RowHeightFixed(3,
				grid.Widget(helloWidget,
					container.Border(linestyle.Light),
					container.BorderTitle("Controls"),
				),
			),
		),
	)

	builder.Add(
		grid.ColWidthPerc(20,
			grid.Widget(chunksWidget,
				container.Border(linestyle.Light),
				container.BorderTitle("Chunks"),
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

type ChunkMark uint8

const (
	Unmarked ChunkMark = 0
	Good     ChunkMark = 1
	Bad      ChunkMark = 2
)

type Chunk struct {
	Content string
	Mark    ChunkMark
}

func getChunks(md string) []Chunk {
	chunks := []Chunk{}
	lines := strings.Split(md, "\n")
	c := Chunk{}
	for _, line := range lines {
		if line == "" || line[:1] == "#" {
			if c.Content == "" {
				continue
			}
			chunks = append(chunks, c)
			c = Chunk{}
			continue
		}
		c.Content += line
	}
	chunks = append(chunks, c)
	return chunks
}

func readScript(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	md := string(b)
	loadedChunks = getChunks(md)
	scriptWidget.SetChunks(loadedChunks)
	return nil
}

func updateSelectedChunk() {
	scriptWidget.SelectChunk(selectedChunk)
	chunksWidget.Write(fmt.Sprintf("%d\n", selectedChunk))
}

func main() {
	scriptFile := flag.String("script", "", "Path to the markdown file to use as input.")
	flag.Parse()

	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()
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
			updateSelectedChunk()
		} else if k.Key == keyboard.KeyArrowDown {
			if selectedChunk < uint(len(loadedChunks)) {
				selectedChunk += 1
			}
			updateSelectedChunk()
		}
	}

	err = readScript(*scriptFile)
	if err != nil {
		t.Close()
		log.Fatalf("Failed to open file %s: %s", *scriptFile, err)
	}

	go record()

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(10*time.Millisecond)); err != nil {
		log.Fatalf("%s", err)
	}
}
