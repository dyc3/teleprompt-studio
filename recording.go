package main

import (
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/zimmski/osutil"
)

// TODO: record audio with type that makes sense, not the type that fits into the line chart the easiest
var audioStream chan []float64 = make(chan []float64)

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
	samples := make([]float64, bufSize)
	for {
		err := stream.Read()
		if err != nil {
			log.Fatalf("Failed to read stream audio: %s", err)
		}

		for i, s := range in {
			samples[i] = float64(s)
		}

		audioStream <- samples
	}
}

var recordedAudio []float64 = make([]float64, 0, 44100)

func audioProcessor() {
	log.Print("Audio processing started")
	for {
		buffer := <-audioStream
		recordedAudio = append(recordedAudio, buffer...)

		displayBuffer := recordedAudio[clamp(len(recordedAudio)-44100, 0, len(recordedAudio)):]
		waveformWidget.Series("Waveform", displayBuffer)
	}
}
