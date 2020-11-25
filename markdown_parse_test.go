package main

import (
	"reflect"
	"testing"
)

func TestParseDoc(t *testing.T) {
	t.Run("1 header, 2 chunks", func(t *testing.T) {
		t.Parallel()
		md := `# Test
chunk 1

chunk 2`
		doc := parseDoc(md)
		if len(doc) != 1 {
			t.Errorf("Incorrect number of headers extracted: %d", len(doc))
		}
		if len(doc[0].Chunks) != 2 {
			t.Errorf("Incorrect number of chunks extracted: %d", len(doc[0].Chunks))
		}
		if doc[0].Text != "# Test" {
			t.Errorf("Incorrect header text: %s", doc[0].Text)
		}
		if doc[0].Chunks[0].Content != "chunk 1" {
			t.Errorf("Incorrect chunk content: %s", doc[0].Chunks[0].Content)
		}
		if doc[0].Chunks[1].Content != "chunk 2" {
			t.Errorf("Incorrect chunk content: %s", doc[0].Chunks[1].Content)
		}
	})

	t.Run("2 headers, 5 chunks", func(t *testing.T) {
		t.Parallel()
		md := `# Ch. 1
chunk A

chunk B
# Ch. 2

chunk A2

chunk B2

chunk C2`
		doc := parseDoc(md)
		if len(doc) != 2 {
			t.Errorf("Incorrect number of headers extracted: %d", len(doc))
		}
		if len(doc[0].Chunks) != 2 {
			t.Errorf("Incorrect number of chunks extracted: %d", len(doc[0].Chunks))
		}
		if doc[0].Text != "# Ch. 1" {
			t.Errorf("Incorrect header text: %s", doc[0].Text)
		}
		if doc[0].Chunks[0].Content != "chunk A" {
			t.Errorf("Incorrect chunk content: %s", doc[0].Chunks[0].Content)
		}
		if doc[0].Chunks[1].Content != "chunk B" {
			t.Errorf("Incorrect chunk content: %s", doc[0].Chunks[1].Content)
		}
		if len(doc[1].Chunks) != 3 {
			t.Errorf("Incorrect number of chunks extracted: %d", len(doc[1].Chunks))
		}
		if doc[1].Text != "# Ch. 2" {
			t.Errorf("Incorrect header text: %s", doc[1].Text)
		}
		if doc[1].Chunks[0].Content != "chunk A2" {
			t.Errorf("Incorrect chunk content: %s", doc[1].Chunks[0].Content)
		}
		if doc[1].Chunks[1].Content != "chunk B2" {
			t.Errorf("Incorrect chunk content: %s", doc[1].Chunks[1].Content)
		}
		if doc[1].Chunks[2].Content != "chunk C2" {
			t.Errorf("Incorrect chunk content: %s", doc[1].Chunks[2].Content)
		}
	})

	t.Run("2 headers, 2 chunks", func(t *testing.T) {
		t.Parallel()
		md := `# Title
# Test
chunk 1

chunk 2`
		doc := parseDoc(md)
		if len(doc) != 2 {
			t.Errorf("Incorrect number of headers extracted: %d", len(doc))
		}
		if len(doc[0].Chunks) != 0 {
			t.Errorf("Incorrect number of chunks extracted: %d", len(doc[0].Chunks))
		}
		if len(doc[1].Chunks) != 2 {
			t.Errorf("Incorrect number of chunks extracted: %d", len(doc[1].Chunks))
		}
		if doc[1].Text != "# Test" {
			t.Errorf("Incorrect header text: %s", doc[1].Text)
		}
		if doc[1].Chunks[0].Content != "chunk 1" {
			t.Errorf("Incorrect chunk content: %s", doc[1].Chunks[0].Content)
		}
		if doc[1].Chunks[1].Content != "chunk 2" {
			t.Errorf("Incorrect chunk content: %s", doc[1].Chunks[1].Content)
		}
	})

	t.Run("2 lines, 1 chunk", func(t *testing.T) {
		t.Parallel()
		md := `line A
line B`
		doc := parseDoc(md)
		if len(doc) != 1 {
			t.Errorf("Incorrect number of headers extracted: %d", len(doc))
		}
		if len(doc[0].Chunks) != 1 {
			t.Errorf("Incorrect number of chunks extracted: %d", len(doc[0].Chunks))
		}
		if doc[0].Chunks[0].Content != "line A line B" {
			t.Errorf("Incorrect chunk content: %s", doc[0].Chunks[0].Content)
		}
	})

	t.Run("markdown list", func(t *testing.T) {
		t.Parallel()
		md := `# Test
- item 1
- item 2`
		doc := parseDoc(md)
		if len(doc) != 1 {
			t.Errorf("Incorrect number of headers extracted: %d", len(doc))
		}
		if len(doc[0].Chunks) != 1 {
			t.Errorf("Incorrect number of chunks extracted: %d", len(doc[0].Chunks))
		}
		if doc[0].Text != "# Test" {
			t.Errorf("Incorrect header text: %s", doc[0].Text)
		}
		if doc[0].Chunks[0].Content != "- item 1\n- item 2" {
			t.Errorf("Incorrect chunk content: %s", doc[0].Chunks[0].Content)
		}
	})

	t.Run("markdown list with indented items", func(t *testing.T) {
		t.Parallel()
		md := `# Test
- item 1
  - subitem 1
- item 2`
		doc := parseDoc(md)
		if len(doc) != 1 {
			t.Errorf("Incorrect number of headers extracted: %d", len(doc))
		}
		if len(doc[0].Chunks) != 1 {
			t.Errorf("Incorrect number of chunks extracted: %d", len(doc[0].Chunks))
		}
		if doc[0].Text != "# Test" {
			t.Errorf("Incorrect header text: %s", doc[0].Text)
		}
		if doc[0].Chunks[0].Content != "- item 1\n  - subitem 1\n- item 2" {
			t.Errorf("Incorrect chunk content: %s", doc[0].Chunks[0].Content)
		}
	})
}

func TestDocRenderable(t *testing.T) {
	doc := Document{
		Header{
			Chunks: []Chunk{
				Chunk{Content: "a"},
			},
			MetaChunks: []MetaChunk{
				MetaChunk{Content: "b"},
			},
			chunkOrder: []chunkOrder{chunk_normal, chunk_meta},
		},
	}
	expect := []interface{}{
		doc[0],
		Chunk{Content: "a"},
		MetaChunk{Content: "b"},
	}

	renderable := doc.GetRenderable()
	if !reflect.DeepEqual(renderable, expect) {
		t.Errorf("Renderable doc does not match expected: %v != %v", renderable, expect)
	}
}
