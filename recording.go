package main

import (
	"log"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/zimmski/osutil"
)

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
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
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

var recordedAudio []int32 = make([]int32, 0, 44100)

func audioProcessor() {
	log.Print("Audio processing started")
	const displayBufferSize = 44100
	for {
		buffer := <-audioStream
		recordedAudio = append(recordedAudio, buffer...)

		dispOffset := clamp(len(recordedAudio)-displayBufferSize, 0, len(recordedAudio))
		displayBuffer := make([]float64, displayBufferSize)
		for i, s := range recordedAudio[dispOffset:] {
			displayBuffer[i] = float64(s)
		}
		waveformWidget.Series("Waveform", displayBuffer)
	}
}

func samplesToDuration(sampleRate int, nSamples int) time.Duration {
	return time.Duration(nSamples/sampleRate) * time.Second
}
