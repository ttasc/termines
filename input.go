// This file handles raw input events from the terminal (keyboard and mouse)
// and translates them into game actions, UI interactions, and state mutations.
package main

import (
	"github.com/ttasc/ttbox"
)

// handleInput processes a single terminal event and updates the GameState.
// It routes keyboard and mouse inputs to the appropriate game logic.
func handleInput(gs *GameState, evt ttbox.Event) {

	isGameOver := gs.Status != StatusPlaying

	switch evt.Type {
	case ttbox.EventKey:
		// Restrict input to reset commands if the game has ended.
		if isGameOver {
			if evt.Ch == 'r' || evt.Ch == 'R' {
				gs.ResetGame(gs.Diff)
			}
			return
		}

		// Center the cursor on first interaction if uninitialized (-1).
		if gs.CursorX < 0 || gs.CursorY < 0 {
			gs.CursorX = gs.Diff.Width / 2
			gs.CursorY = gs.Diff.Height / 2
		}

		// Handle directional keys for cursor movement.
		switch evt.Key {
		case ttbox.KeyEnter:
			reveal(gs, gs.CursorX, gs.CursorY)
		case ttbox.KeyArrowUp:
			gs.CursorY = clamp(gs.CursorY-1, 0, gs.Diff.Height-1)
		case ttbox.KeyArrowDown:
			gs.CursorY = clamp(gs.CursorY+1, 0, gs.Diff.Height-1)
		case ttbox.KeyArrowLeft:
			gs.CursorX = clamp(gs.CursorX-1, 0, gs.Diff.Width-1)
		case ttbox.KeyArrowRight:
			gs.CursorX = clamp(gs.CursorX+1, 0, gs.Diff.Width-1)
		}

		// Handle vim-keys for movement and specific action characters.
		switch evt.Ch {
		case ' ':
			reveal(gs, gs.CursorX, gs.CursorY)
		case 'h', 'H':
			gs.CursorX = clamp(gs.CursorX-1, 0, gs.Diff.Width-1)
		case 'l', 'L':
			gs.CursorX = clamp(gs.CursorX+1, 0, gs.Diff.Width-1)
		case 'k', 'K':
			gs.CursorY = clamp(gs.CursorY-1, 0, gs.Diff.Height-1)
		case 'j', 'J':
			gs.CursorY = clamp(gs.CursorY+1, 0, gs.Diff.Height-1)
		case 'f', 'F':
			toggleFlag(gs, gs.CursorX, gs.CursorY)
		case 'r', 'R':
			gs.ResetGame(gs.Diff)
		case '1':
			gs.ResetGame(DiffEasy)
		case '2':
			gs.ResetGame(DiffMedium)
		case '3':
			gs.ResetGame(DiffHard)
		}

	case ttbox.EventMouse:
		if evt.Press {
			termW, termH := ttbox.Size()

			// Calculate UI bounding boxes dynamically to handle terminal resizing.
			diffBtn, resetBtn := getUIElements(gs, termW, termH)

			if evt.Button == ttbox.MouseLeft {
				// Check for difficulty toggle button click.
				if evt.X >= diffBtn.X && evt.X < diffBtn.X+diffBtn.W && evt.Y == diffBtn.Y {
					cycleDifficulty(gs)
					return
				}
				// Check for manual reset button click.
				if evt.X >= resetBtn.X && evt.X < resetBtn.X+resetBtn.W && evt.Y == resetBtn.Y {
					gs.ResetGame(gs.Diff)
					return
				}
			}

			// Ignore board interaction if the game has ended.
			if isGameOver {
				return
			}

			// Map absolute terminal coordinates to 2D board indices.
			boardX, boardY, ok := screenToBoard(gs, evt.X, evt.Y)
			if ok {
				switch evt.Button {
				case ttbox.MouseLeft:
					// Require a second click (or pressing space/enter) to reveal a cell.
					// The first click only snaps the cursor to the target cell.
					if gs.CursorX == boardX && gs.CursorY == boardY {
						reveal(gs, boardX, boardY)
					} else {
						gs.CursorX = boardX
						gs.CursorY = boardY
					}
				case ttbox.MouseRight:
					// Right click moves the cursor and instantly toggles the flag.
					gs.CursorX = boardX
					gs.CursorY = boardY
					toggleFlag(gs, boardX, boardY)
				}
			} else {
				// Clicked outside board bounds; hide cursor.
				gs.CursorX = -1
				gs.CursorY = -1
			}
		}
	}
}

// cycleDifficulty rotates the current game difficulty through Easy, Medium, and Hard.
// It triggers an immediate game reset upon changing.
func cycleDifficulty(gs *GameState) {
	switch gs.Diff.Name {
	case DiffEasy.Name:
		gs.ResetGame(DiffMedium)
	case DiffMedium.Name:
		gs.ResetGame(DiffHard)
	default:
		gs.ResetGame(DiffEasy)
	}
}

// clamp restricts a value to the inclusive range [min, max].
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// screenToBoard translates absolute terminal coordinates into grid array indices.
// Returns the x, y board coordinates and a boolean indicating if the click was within bounds.
func screenToBoard(gs *GameState, screenX, screenY int) (int, int, bool) {
	termW, termH := ttbox.Size()
	offsetX, offsetY := getBoardOffsets(gs, termW, termH)

	// Cells are rendered 3 characters wide in the terminal (e.g., "[ ]").
	boardPixelWidth := gs.Diff.Width * 3

	if screenX >= offsetX && screenX < offsetX+boardPixelWidth &&
		screenY >= offsetY && screenY < offsetY+gs.Diff.Height {
		// Divide relative X by 3 to map the 3-character UI width back to a single array index.
		return (screenX - offsetX) / 3, screenY - offsetY, true
	}

	return 0, 0, false
}
