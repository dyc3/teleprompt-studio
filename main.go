package main

import (
	"context"
	"log"

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

	builder := grid.New()
	builder.Add(
		grid.ColWidthPerc(80,
			grid.RowHeightPerc(50,
				grid.Widget(helloWidget,
					container.Border(linestyle.Light),
					container.BorderTitle("Script"),
				),
			),
			grid.RowHeightPerc(45,
				grid.Widget(helloWidget,
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
			grid.Widget(helloWidget,
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

func main() {
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		log.Fatal(err)
	}
	c := buildLayout(t)

	ctx, cancel := context.WithCancel(context.Background())

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			t.Close()
			cancel()
		}
	}
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		log.Fatalf("%s", err)
	}
}
