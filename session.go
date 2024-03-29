package main

import (
	"encoding/csv"
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
	hasBeenSaved     bool
	streamFileHandle *os.File
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

func (s *Session) getSessionDir() (string, error) {
	os.Mkdir(SessionsFolder, 0755)

	dir := path.Join(SessionsFolder, fmt.Sprintf("%d", s.Id))
	err := os.Mkdir(dir, 0755)
	if os.IsExist(err) {
		return dir, nil
	}
	if err != nil {
		return "", err
	}
	return dir, nil
}

// Deprecated: audio is now streamed to disk as it is being recorded.
func (s *Session) saveAudio() error {
	dir, err := s.getSessionDir()
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

	return nil
}

func (s *Session) saveTakes() error {
	dir, err := s.getSessionDir()
	if err != nil {
		return err
	}

	takesFile, err := os.Create(path.Join(dir, "takes.csv"))
	if err != nil {
		log.Print("Failed to create takes file")
		return err
	}
	defer takesFile.Close()
	w := csv.NewWriter(takesFile)
	defer w.Flush()
	err = w.Write([]string{"header", "chunk_index", "chunk_text", "take_index", "take_mark", "take_start", "take_end"})
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
				err = w.Write([]string{
					header.Text,
					fmt.Sprintf("%d", c),
					chunk.Content[:clamp(32, 0, len(chunk.Content))] + "...",
					fmt.Sprintf("%d", t),
					fmt.Sprintf("%s", take.Mark),
					fmt.Sprintf("%s", Timestamp(&syncedStart)),
					fmt.Sprintf("%s", Timestamp(&syncedEnd)),
				})
				if err != nil {
					log.Print("Failed to write takes")
					return err
				}
			}
		}
	}
	return nil
}

func (s *Session) saveMetadata() error {
	dir, err := s.getSessionDir()
	if err != nil {
		return err
	}

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

	return nil
}

func (s *Session) FullSave() error {
	err := os.Mkdir(SessionsFolder, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	s.deriveId()

	err = s.saveMetadata()
	if err != nil {
		log.Print("Failed to save metadata")
		return err
	}

	err = s.saveTakes()
	if err != nil {
		log.Print("Failed to save takes")
		return err
	}

	dir, err := s.getSessionDir()
	log.Printf("Current session successfully saved: %s", dir)

	s.hasBeenSaved = true

	return nil
}

func (s *Session) StartStreamingToDisk() (*wav.Encoder, error) {
	dir, err := s.getSessionDir()
	if err != nil {
		return nil, err
	}

	var audioFile *os.File
	audioFile, err = os.Create(path.Join(dir, "audio.wav"))
	if err != nil {
		return nil, err
	}
	s.streamFileHandle = audioFile

	e := wav.NewEncoder(audioFile, sampleRate, 32, 1, 1)
	s.hasBeenSaved = true
	return e, nil
}

func (s *Session) StopStreamingToDisk(e *wav.Encoder) error {
	err := e.Close()
	if err != nil {
		return err
	}
	s.streamFileHandle.Close()
	return nil
}
