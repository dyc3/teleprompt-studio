package main

import "testing"

func TestGetChunks(t *testing.T) {
	md := `# Test
chunk 1

chunk 2`
	chunks := getChunks(md)
	if len(chunks) != 2 {
		t.Errorf("Incorrect number of chunks extracted: %d", len(chunks))
	}
	if chunks[0].Content != "chunk 1" {
		t.Errorf("Incorrect chunk content: %s", chunks[0].Content)
	}
	if chunks[1].Content != "chunk 2" {
		t.Errorf("Incorrect chunk content: %s", chunks[1].Content)
	}
}
