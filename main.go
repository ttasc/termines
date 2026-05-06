// This file is the entry point for the application.
// It bootstraps the terminal UI, initializes the game state,
// and orchestrates the primary poll-update-render game loop.
package main

import (
	"fmt"
	"time"

	"github.com/ttasc/ttbox"
)

// main initializes the TUI environment and runs the continuous game loop.
func main() {

	// Initialize terminal backend and ensure cleanup on exit.
	if err := ttbox.Init(); err != nil {
		fmt.Printf("Error initializing TUI: %v\n", err)
		return
	}
	defer ttbox.Close()

	ttbox.HideCursorFunc()
	ttbox.EnableMouseFunc()

	state := NewGameState(DiffMedium)

	isRunning := true
	for isRunning {

		// Poll for events with a timeout rather than blocking indefinitely.
		// A 100ms timeout (approx. 10 FPS) keeps CPU usage minimal while ensuring
		// the elapsed game timer updates and renders even when the user is idle.
		evt, err := ttbox.PollEventTimeout(100 * time.Millisecond)

		if err == nil {

			// Global exit condition.
			if evt.Type == ttbox.EventKey && evt.Key == ttbox.KeyEscape {
				isRunning = false
			} else {

				// Route all other terminal events to the input processor.
				handleInput(state, evt)
			}
		}

		// Sync the elapsed game duration if currently playing.
		// state.UpdateTimer()

		// Clear and redraw the entire terminal frame based on the mutated state.
		render(state)
	}
}
