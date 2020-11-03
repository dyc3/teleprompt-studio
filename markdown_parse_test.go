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
}
