package main

import (
	"io/ioutil"
	"strings"
)

type TakeMark uint8

const (
	Unmarked TakeMark = 0
	Good     TakeMark = 1
	Bad      TakeMark = 2
)

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

type Header struct {
	Chunks []Chunk
	Text   string
}

type Chunk struct {
	Content string
	Takes   []Take
}

func parseDoc(md string) Document {
	headers := []Header{}
	lines := strings.Split(md, "\n")
	h := Header{}
	c := Chunk{}
	for _, line := range lines {
		if line != "" && line[:1] == "#" {
			if h.Text == "" {
				h.Text = line
			} else {
				if c.Content != "" {
					c.Content = strings.TrimSpace(c.Content)
					h.Chunks = append(h.Chunks, c)
				}
				headers = append(headers, h)
				h = Header{
					Text: line,
				}
				c = Chunk{}
			}
			continue
		}
		if line == "" {
			if c.Content == "" {
				continue
			}
			c.Content = strings.TrimSpace(c.Content)
			h.Chunks = append(h.Chunks, c)
			c = Chunk{}
			continue
		}
		line = strings.TrimSpace(line)
		if c.Content != "" {
			if strings.HasPrefix(line, "```") {
				c.Content += "\n"
			} else {
				switch line[0] {
				case '-':
					c.Content += "\n"
				default:
					c.Content += " "
				}
			}
		}
		c.Content += line
		if line == "```" {
			c.Content += "\n"
		}
	}
	h.Chunks = append(h.Chunks, c)
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

	for _, header := range currentSession.Doc {
		renderable = append(renderable, header)
		for _, chunk := range header.Chunks {
			renderable = append(renderable, chunk)
		}
	}

	return renderable
}
