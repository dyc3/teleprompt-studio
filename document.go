package main

import (
	"errors"
	"io/ioutil"
	"strings"
)

type TakeMark uint8

const (
	Unmarked TakeMark = 0
	Good     TakeMark = 1
	Bad      TakeMark = 2
)

// Metadata prefixes used in scripts. They should be omited from selectable chunks.
func getMetaPrefixes() []string {
	return []string{"TODO", "REF", "NOTE", "BIT"}
}

func (m TakeMark) String() string {
	switch m {
	case Unmarked:
		return "unmarked"
	case Good:
		return "good"
	case Bad:
		return "bad"
	}
	return "unknown"
}

type Document []Header

func (doc *Document) CountChunks() int {
	c := 0
	for _, h := range *doc {
		c += len(h.Chunks)
	}
	return c
}

func (doc *Document) GetChunk(index int) *Chunk {
	for _, h := range *doc {
		if index-len(h.Chunks) < 0 {
			return &h.Chunks[index]
		}
		index -= len(h.Chunks)
	}
	return nil
}

type chunkOrder uint8

const (
	chunk_normal chunkOrder = 0
	chunk_meta   chunkOrder = 1
)

type Header struct {
	Text       string
	Chunks     []Chunk
	MetaChunks []MetaChunk
	chunkOrder []chunkOrder
}

type Chunk struct {
	Content string
	Takes   []Take
}

// A chunk that contains content that should not be selectable for takes.
type MetaChunk struct {
	Content string
}

func (h *Header) AddChunk(chunk interface{}) error {
	var ord chunkOrder
	switch c := chunk.(type) {
	case Chunk:
		ord = chunk_normal
		h.Chunks = append(h.Chunks, c)
	case MetaChunk:
		ord = chunk_meta
		h.MetaChunks = append(h.MetaChunks, c)
	default:
		return errors.New("Invalid type for chunk")
	}
	h.chunkOrder = append(h.chunkOrder, ord)

	return nil
}

func isMeta(line string) bool {
	for _, prefix := range getMetaPrefixes() {
		if strings.HasPrefix(line, prefix+":") {
			return true
		}
	}
	return false
}

func parseDoc(md string) Document {
	headers := []Header{}
	lines := strings.Split(md, "\n")
	h := Header{}
	text := ""
	doAddChunk := func() {
		if strings.Contains(h.Text, "Intro bit") {
			h.AddChunk(MetaChunk{
				Content: text,
			})
			return
		}

		if isMeta(text) {
			h.AddChunk(MetaChunk{
				Content: text,
			})
		} else {
			h.AddChunk(Chunk{
				Content: text,
			})
		}
	}
	for _, line := range lines {
		if line != "" && line[:1] == "#" {
			if h.Text == "" {
				h.Text = line
			} else {
				if text != "" {
					text = strings.TrimSpace(text)
					doAddChunk()
				}
				headers = append(headers, h)
				h = Header{
					Text: line,
				}
				text = ""
			}
			continue
		}
		if line == "" {
			if text == "" {
				continue
			}
			text = strings.TrimSpace(text)
			doAddChunk()
			text = ""
			continue
		}
		trimmed := strings.TrimSpace(line)
		if text != "" {
			if strings.HasPrefix(line, "```") {
				text += "\n"
			} else if isMeta(line) {
				text += "\n"
			} else {
				switch trimmed[0] {
				case '-':
					text += "\n"
				default:
					text += " "
				}
			}
		}
		text += line
		if line == "```" {
			text += "\n"
		}
	}
	doAddChunk()
	headers = append(headers, h)
	return headers
}

func readScript(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	md := string(b)
	currentSession.Doc = parseDoc(md)
	return nil
}

func (d *Document) GetAllTakes() []Take {
	var takes []Take
	for _, h := range *d {
		for _, c := range h.Chunks {
			takes = append(takes, c.Takes...)
		}
	}
	return takes
}

func (d *Document) GetRenderable() []interface{} {
	var renderable []interface{}

	for _, header := range *d {
		renderable = append(renderable, header)
		idxs := map[chunkOrder]int{
			chunk_normal: 0,
			chunk_meta:   0,
		}
		for _, ord := range header.chunkOrder {
			var chunk interface{}
			switch ord {
			case chunk_normal:
				chunk = header.Chunks[idxs[ord]]
			case chunk_meta:
				chunk = header.MetaChunks[idxs[ord]]
			}
			renderable = append(renderable, chunk)
			idxs[ord]++
		}
	}

	return renderable
}
