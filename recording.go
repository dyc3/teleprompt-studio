package main

import (
	"errors"
	"log"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/zimmski/osutil"
)

const sampleRate = 44100

var isRecording bool = false
var audioStream chan []int32 = make(chan []int32)
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

func EndSession() {
	if isRecording {
		isRecording = false
	}
	// TODO: save session
}

func audioProcessor() {
	log.Print("Audio processing started")
	const displayBufferSize = sampleRate
	for {
		buffer := <-audioStream
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
