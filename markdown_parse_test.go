package main

import "testing"

func TestParseDoc(t *testing.T) {
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

	md = `# Ch. 1
chunk A

chunk B
# Ch. 2

chunk A2

chunk B2

chunk C2`
	doc = parseDoc(md)
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

	md = `# Title
# Test
chunk 1

chunk 2`
	doc = parseDoc(md)
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
}
