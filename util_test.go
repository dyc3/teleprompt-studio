package main

import (
	"reflect"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas/buffer"
)

func TestMarkdownFontModifiers(t *testing.T) {
	input := "hello **world**!"
	cells := markdownFontModifiers(buffer.NewCells(input))
	expect_cells := buffer.NewCells("hello ")
	expect_cells = append(expect_cells, buffer.NewCells("world", cell.Bold())...)
	expect_cells = append(expect_cells, buffer.NewCells("!")...)
	if !reflect.DeepEqual(cells, expect_cells) {
		t.Errorf("Cells were not equal: len %d != len %d", len(cells), len(expect_cells))
	}
}
