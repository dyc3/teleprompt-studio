package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/go-audio/audio"

	"github.com/go-audio/wav"

	"github.com/gordonklaus/portaudio"
	"github.com/zimmski/osutil"
)

const sampleRate = 44100

var isRecording bool = false
var audioStream chan []int32 = make(chan []int32, 10)
var currentSession Session
var isPlaying bool = false

// Playback position in samples
var playbackPosition int = 0

type Session struct {
	Audio []int32
	Doc   Document
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

var portaudioInitialized bool = false

func initPortAudio() {
	var err error
	o, _ := osutil.CaptureWithCGo(func() {
		err = portaudio.Initialize()
	})
	log.Printf(string(o))
	if err != nil {
		log.Fatalf("Failed to initialize recording: %s", err)
	}
	portaudioInitialized = true
}

func record() {
	const bufSize = 1024

	// This is based on the record example shown in the portaudio repo.
	// It's unclear whether or not framesPerBuffer should match the buffer size
	// or be zero (where portaudio will provide variable length buffers).
	in := make([]int32, bufSize)
	stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, len(in), in)
	if err != nil {
		log.Fatalf("Failed to open stream audio: %s", err)
	}
	defer stream.Close()

	err = stream.Start()
	if err != nil {
		log.Fatalf("Failed to start stream audio: %s", err)
	}

	isRecording = true
	log.Print("Recording started")
	for {
		// wait for enough audio to fill the buffer
		for avail := 0; avail < len(in); avail, _ = stream.AvailableToRead() {
			time.Sleep(time.Second / sampleRate * time.Duration(len(in)-avail) / 2)
		}

		err := stream.Read()
		if err != nil {
			log.Fatalf("Failed to read stream audio: %s", err)
		}
		audioStream <- in
		if !isRecording {
			break
		}
	}
	log.Print("Recording stopped")
}

func StartSession() {
	if !isRecording {
		go record()
	}
}

func EndSession() error {
	if isRecording {
		isRecording = false
	}
	err := currentSession.Save()
	if err != nil {
		log.Printf("Failed to save session: %s", err)
		return err
	}
	return nil
}

func audioProcessor() {
	log.Print("Audio processing started")
	const displayBufferSize = sampleRate
	for {
		buffer := <-audioStream
		if cap(audioStream)-len(audioStream) < 3 {
			log.Printf("WARNING: audioStream channel is being overloaded, buffered messages: %d/%d", len(audioStream), cap(audioStream))
		}
		currentSession.Audio = append(currentSession.Audio, buffer...)

		if isRecordingTake {
			if isRecordingSyncTake {
				currentSession.Doc.syncTakes[selectedTake].End = samplesToDuration(sampleRate, len(currentSession.Audio))
			} else {
				chunk := currentSession.Doc.GetChunk(int(selectedChunk))
				chunk.Takes[selectedTake].End = samplesToDuration(sampleRate, len(currentSession.Audio))
			}
		}
	}
}

func playbackTimespan(timespan TimeSpan) {
	isPlaying = true
	log.Printf("playing back %d samples...", len(currentSession.Audio))
	// This is based on the play example shown in the portaudio repo.
	const bufSize = 1024

	if !portaudioInitialized {
		initPortAudio()
	}

	out := make([]int32, bufSize)
	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, len(out), &out)
	if err != nil {
		log.Fatalf("Failed to open stream audio: %s", err)
	}
	defer stream.Close()
	log.Printf("stream open: %v", stream.Info())

	err = stream.Start()
	if err != nil {
		log.Fatalf("Failed to start stream audio: %s", err)
	}
	defer stream.Stop()
	log.Printf("stream started")

	start := durationToSamples(sampleRate, timespan.Start)
	end := durationToSamples(sampleRate, timespan.End)
	samples := currentSession.Audio[start:end]
	for b := 0; b < len(samples); b += len(out) {
		playbackPosition = start + b
		out = samples[b:clamp(b+bufSize, 0, len(samples))]
		err := stream.Write()
		if err != nil {
			log.Fatalf("Failed to write stream audio: %v", err)
		}
	}

	log.Printf("playback complete")
	isPlaying = false
}

func playbackTake(take Take) {
	playbackTimespan(TimeSpan{
		Start: take.Start,
		End:   take.End,
	})
}

func samplesToDuration(sampleRate int, nSamples int) time.Duration {
	return time.Duration(nSamples) * time.Second / time.Duration(sampleRate)
}

func durationToSamples(sampleRate int, d time.Duration) int {
	return int(d.Seconds() * float64(sampleRate))
}

type TimeSpan struct {
	Start time.Duration
	End   time.Duration
}

func (t *TimeSpan) Duration() time.Duration {
	return t.End - t.Start
}

type Take struct {
	TimeSpan
	Mark TakeMark
}

func startTake(sync bool) error {
	if isRecordingTake {
		return errors.New("Already recording take")
	}
	take := Take{}
	take.Start = samplesToDuration(sampleRate, len(currentSession.Audio))
	if sync {
		take.Mark = Sync
		currentSession.Doc.syncTakes = append(currentSession.Doc.syncTakes, take)
		selectedTake = len(currentSession.Doc.syncTakes) - 1
		isRecordingSyncTake = true
	} else {
		chunk := currentSession.Doc.GetChunk(int(selectedChunk))
		chunk.Takes = append(chunk.Takes, take)
		selectedTake = len(chunk.Takes) - 1
	}
	isRecordingTake = true
	return nil
}

func endTake() error {
	if !isRecordingTake {
		return errors.New("Not recording take")
	}
	if isRecordingSyncTake {
		currentSession.Doc.syncTakes[selectedTake].End = samplesToDuration(sampleRate, len(currentSession.Audio))
		isRecordingSyncTake = false
		if currentSession.Doc.SyncOffset == time.Duration(0) {
			currentSession.updateSyncOffset()
		}
	} else {
		chunk := currentSession.Doc.GetChunk(int(selectedChunk))
		chunk.Takes[selectedTake].End = samplesToDuration(sampleRate, len(currentSession.Audio))
	}
	isRecordingTake = false
	return nil
}

func (s *Session) Save() error {
	err := os.Mkdir("sessions", 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	num := 0
	for {
		_, err := os.Stat(path.Join("sessions", fmt.Sprintf("%d", num)))
		if os.IsNotExist(err) {
			break
		}
		num++
	}

	dir := path.Join("sessions", fmt.Sprintf("%d", num))
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
					log.Print("Failed to write takes header")
					return err
				}
			}
		}
	}

	log.Printf("Current session successfully saved: %s", dir)

	return nil
}
