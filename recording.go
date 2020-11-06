package main

import (
	"errors"
	"log"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/zimmski/osutil"
)

const sampleRate = 44100

var audioStream chan []int32 = make(chan []int32)

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

	log.Print("Recording started")
	for {
		err := stream.Read()
		if err != nil {
			log.Fatalf("Failed to read stream audio: %s", err)
		}
		audioStream <- in
	}
}

var recordedAudio []int32 = make([]int32, 0, sampleRate)

func audioProcessor() {
	log.Print("Audio processing started")
	const displayBufferSize = sampleRate
	for {
		buffer := <-audioStream
		recordedAudio = append(recordedAudio, buffer...)

		if isRecordingTake {
			chunk := doc.GetChunk(int(selectedChunk))
			chunk.Takes[selectedTake].End = samplesToDuration(sampleRate, len(recordedAudio))
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
	chunk := doc.GetChunk(int(selectedChunk))
	take := Take{}
	take.Start = samplesToDuration(sampleRate, len(recordedAudio))
	chunk.Takes = append(chunk.Takes, take)
	selectedTake = len(chunk.Takes) - 1
	isRecordingTake = true
	return nil
}

func endTake() error {
	if !isRecordingTake {
		return errors.New("Not recording take")
	}
	chunk := doc.GetChunk(int(selectedChunk))
	chunk.Takes[selectedTake].End = samplesToDuration(sampleRate, len(recordedAudio))
	isRecordingTake = false
	return nil
}
