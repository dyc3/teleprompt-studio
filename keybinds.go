package main

import (
	"log"

	"github.com/mum4k/termdash/keyboard"
)

type keybind struct {
	key      keyboard.Key
	desc     string
	callback func()
}

func getAvailableKeybinds() []keybind {
	var keys []keybind
	if !isRecordingTake {
		keys = append(keys, []keybind{
			{
				key:      keyboard.KeyArrowDown,
				desc:     "Next Chunk",
				callback: keybindNextChunk,
			},
			{
				key:      keyboard.KeyArrowUp,
				desc:     "Previous Chunk",
				callback: keybindPreviousChunk,
			},
			{
				key:      ' ',
				desc:     "Start Take",
				callback: func() { startTake(false) },
			},
			{
				key:      's',
				desc:     "Start Sync Take",
				callback: func() { startTake(true) },
			},
			{
				key:      'r',
				desc:     "End Session",
				callback: keybindEndSession,
			},
		}...)

		if len(currentSession.Doc.headers) > 0 && len(currentSession.Doc.GetChunk(int(selectedChunk)).Takes) > 0 {
			keys = append(keys, []keybind{
				{
					key:      'g',
					desc:     "Mark Good",
					callback: keybindMarkGood,
				},
				{
					key:      'b',
					desc:     "Mark Bad",
					callback: keybindMarkBad,
				},
				{
					key:      'p',
					desc:     "Play Selected Take",
					callback: keybindPlayTake,
				},
			}...)
		}

		if ui.audio.selectionActive {
			keys = append(keys,
				keybind{
					key:      't',
					desc:     "New Take from selection",
					callback: keybindCreateTakeFromSelection,
				},
			)
		}
	} else {
		keys = append(keys, []keybind{
			{
				key:      ' ',
				desc:     "End Take",
				callback: func() { endTake() },
			},
			{
				key:      'g',
				desc:     "End Take & Mark Good",
				callback: keybindMarkGood,
			},
			{
				key:      'b',
				desc:     "End Take & Mark Bad",
				callback: keybindMarkBad,
			},
		}...)
	}
	keys = append(keys, []keybind{
		{
			key:      'f',
			desc:     "Toggle Stick Viewport To End",
			callback: func() { ui.audio.stickToEnd = !ui.audio.stickToEnd },
		},
		{
			key:      keyboard.KeyCtrlD,
			desc:     "Toggle Debug",
			callback: func() { ui.audio.showDebug = !ui.audio.showDebug },
		},
	}...)
	return keys
}

func keybindPreviousChunk() {
	if isRecordingTake {
		return
	}
	if selectedChunk > 0 {
		selectedChunk -= 1
	}
	chunk := currentSession.Doc.GetChunk(int(selectedChunk))
	selectedTake = len(chunk.Takes) - 1
}

func keybindNextChunk() {
	if isRecordingTake {
		return
	}
	if selectedChunk < uint(currentSession.Doc.CountChunks()-1) {
		selectedChunk += 1
	}
	chunk := currentSession.Doc.GetChunk(int(selectedChunk))
	selectedTake = len(chunk.Takes) - 1
}

func keybindMarkGood() {
	currentSession.Doc.GetChunk(int(selectedChunk)).Takes[selectedTake].Mark = Good
	if isRecordingTake {
		endTake()
	}
}

func keybindMarkBad() {
	currentSession.Doc.GetChunk(int(selectedChunk)).Takes[selectedTake].Mark = Bad
	if isRecordingTake {
		endTake()
	}
}

func keybindCreateTakeFromSelection() {
	if ui.audio.selectionActive {
		chunk := currentSession.Doc.GetChunk(int(selectedChunk))
		take := Take{}
		take.Start = ui.audio.selected.Start
		take.End = ui.audio.selected.End
		chunk.Takes = append(chunk.Takes, take)
		selectedTake = len(chunk.Takes) - 1
		ui.audio.Deselect()
	}
}

func keybindEndSession() {
	err := EndSession()
	if err != nil {
		log.Print(err)
	}
}

func keybindPlayTake() {
	take := currentSession.Doc.GetChunk(int(selectedChunk)).Takes[selectedTake]
	go playbackTake(take)
}
