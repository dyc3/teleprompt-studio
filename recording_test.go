package main

import (
	"testing"
	"time"
)

func TestSamplesToDuration(t *testing.T) {
	if samplesToDuration(44100, 44100) != 1*time.Second {
		t.Errorf("Incorrect samples to duration")
	}

	if samplesToDuration(44100, 44100*2) != 2*time.Second {
		t.Errorf("Incorrect samples to duration")
	}
}
