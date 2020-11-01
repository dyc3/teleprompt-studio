# teleprompt-studio

A terminal program that makes it easier and faster to record and manage voice over, particularly for video production, written in Go.

# Workflow

1. The user provides a script as a markdown file.
2. The contents of the markdown file are split into chunks.
3. A new session is started, and audio begins recording.
4. The chunks will be displayed to the user for them to read.
   1. The current chunk is highlighted, and can be changed using the arrow keys.
5. The user presses a key to start a take for the selected chunk.
6. The user records the take and presses the key again to stop the take.
7. The timestamps of the take are recorded, and stored with other takes for that chunk.
8. The user can optionally mark the previous take as good or bad.
9. When the user wishes to end the session, they can stop recording. Starting the recording again will require a new session.

# Features

- Simple UI
- Save audio to wav or flac
- Video/Audio Sync Marker
- Waveform visualization
- Take previewing
- Good/bad take markers
- Markdown support
