package main

import (
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

type Session struct {
	Audio []int32
	Doc   Document
}

func record() {
	const bufSize = 1024

	var err error
	osutil.CaptureWithCGo(func() {
		err = portaudio.Initialize()
	})
	if err != nil {
		log.Fatalf("Failed to initialize recording: %s", err)
	}
	defer portaudio.Terminate()
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
		err := stream.Read()
		if err != nil {
			log.Fatalf("Failed to read stream audio: %s", err)
		}
		audioStream <- in
		if !isRecording {
			break
		}
	}
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
			chunk := currentSession.Doc.GetChunk(int(selectedChunk))
			chunk.Takes[selectedTake].End = samplesToDuration(sampleRate, len(currentSession.Audio))
		}
	}
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

func startTake() error {
	if isRecordingTake {
		return errors.New("Already recording take")
	}
	chunk := currentSession.Doc.GetChunk(int(selectedChunk))
	take := Take{}
	take.Start = samplesToDuration(sampleRate, len(currentSession.Audio))
	chunk.Takes = append(chunk.Takes, take)
	selectedTake = len(chunk.Takes) - 1
	isRecordingTake = true
	return nil
}

func endTake() error {
	if !isRecordingTake {
		return errors.New("Not recording take")
	}
	chunk := currentSession.Doc.GetChunk(int(selectedChunk))
	chunk.Takes[selectedTake].End = samplesToDuration(sampleRate, len(currentSession.Audio))
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
	buf := audio.IntBuffer{
		Format:         audio.FormatMono44100,
		SourceBitDepth: 32,
	}
	for _, sample := range s.Audio {
		buf.Data = append(buf.Data, int(sample))
	}
	err = e.Write(&buf)
	if err != nil {
		return err
	}

	return nil
}
