package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const SessionsFolder = "sessions"

type Session struct {
	Audio []int32
	Doc   Document
	Id    int

	// Indicates whether the session has been saved to disk.
	hasBeenSaved bool
}

func (s *Session) ExtractAudio(timespan TimeSpan) []int32 {
	startIdx := durationToSamples(sampleRate, timespan.Start)
	endIdx := durationToSamples(sampleRate, timespan.End)
	return s.Audio[startIdx:endIdx]
}

func (s *Session) updateSyncOffset() {
	// take the audio from the first sync take, find peak, set Doc.syncOffset
	t := s.Doc.syncTakes[0].TimeSpan
	a := s.ExtractAudio(t)
	peakIdx := indexOfMaxInt32(a)
	relOffset := samplesToDuration(sampleRate, peakIdx)
	s.Doc.SyncOffset = t.Start + relOffset
}

// Derive the session ID by checking how many sessions exist in the directory.
func (s *Session) deriveId() {
	if s.hasBeenSaved {
		return
	}
	num := 0
	for {
		_, err := os.Stat(path.Join(SessionsFolder, fmt.Sprintf("%d", num)))
		if os.IsNotExist(err) {
			break
		}
		num++
	}
	s.Id = num
}

func (s *Session) FullSave() error {
	err := os.Mkdir(SessionsFolder, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	s.deriveId()

	dir := path.Join(SessionsFolder, fmt.Sprintf("%d", s.Id))
	err = os.Mkdir(dir, 0755)
	if err != nil {
		return err
	}
	audioFile, err := os.Create(path.Join(dir, "audio.wav"))
	if err != nil {
		return err
	}
	defer audioFile.Close()

	e := wav.NewEncoder(audioFile, sampleRate, 32, 1, 1)
	defer e.Close()
	buf := audio.PCMBuffer{
		Format:         audio.FormatMono44100,
		DataType:       audio.DataTypeI32,
		SourceBitDepth: 32,
		I32:            s.Audio,
	}
	err = e.Write(buf.AsIntBuffer())
	if err != nil {
		return err
	}

	// save metadata
	sessionMetadata, err := json.Marshal(
		struct {
			SyncOffset string `json:"SyncOffset"`
		}{
			Timestamp(&currentSession.Doc.SyncOffset),
		},
	)
	if err != nil {
		log.Print("Failed to marshal session metadata")
		return err
	}
	metadataFile, err := os.Create(path.Join(dir, "metadata.json"))
	if err != nil {
		log.Print("Failed to create metadata file")
		return err
	}
	defer metadataFile.Close()
	metadataFile.Write(sessionMetadata)

	// save takes
	takesFile, err := os.Create(path.Join(dir, "takes.csv"))
	if err != nil {
		log.Print("Failed to create takes file")
		return err
	}
	defer takesFile.Close()
	_, err = takesFile.WriteString("header,chunk_index,chunk_text,take_index,take_mark,take_start,take_end\n")
	if err != nil {
		log.Print("Failed to write takes header")
		return err
	}
	syncOffset := currentSession.Doc.SyncOffset
	for _, header := range currentSession.Doc.headers {
		for c, chunk := range header.Chunks {
			for t, take := range chunk.Takes {
				syncedStart := take.Start - syncOffset
				syncedEnd := take.End - syncOffset
				line := fmt.Sprintf("%s,%d,%s...,%d,%s,%s,%s\n", header.Text, c, chunk.Content[:clamp(32, 0, len(chunk.Content))], t, take.Mark, Timestamp(&syncedStart), Timestamp(&syncedEnd))
				_, err = takesFile.WriteString(line)
				if err != nil {
					log.Print("Failed to write takes")
					return err
				}
			}
		}
	}

	log.Printf("Current session successfully saved: %s", dir)

	s.hasBeenSaved = true

	return nil
}
