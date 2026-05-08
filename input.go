// This file handles raw input events from the terminal (keyboard and mouse)
// and translates them into game actions, UI interactions, and state mutations.
package main

import (
	"github.com/ttasc/ttbox"
)

func handleInput(gs *GameState, evt ttbox.Event) {
	switch evt.Type {
	case ttbox.EventKey:
		handleKeyboardEvent(gs, evt)
	case ttbox.EventMouse:
		handleMouseEvent(gs, evt)
	}
}

func handleKeyboardEvent(gs *GameState, evt ttbox.Event) {
	isGameOver := gs.Status != StatusPlaying

	if isGameOver {
		if evt.Ch == 'r' || evt.Ch == 'R' {
			gs.ResetGame(gs.Diff)
		}
		return
	}

	if gs.CursorX < 0 || gs.CursorY < 0 {
		gs.CursorX = gs.Diff.Width / 2
		gs.CursorY = gs.Diff.Height / 2
	}

	// Arrows
	switch evt.Key {
	case ttbox.KeyEnter:
		if gs.Board[gs.CursorY][gs.CursorX].IsRevealed {
			chord(gs, gs.CursorX, gs.CursorY)
		} else {
			reveal(gs, gs.CursorX, gs.CursorY)
		}
	case ttbox.KeyArrowUp:
		gs.CursorY = clamp(gs.CursorY-1, 0, gs.Diff.Height-1)
	case ttbox.KeyArrowDown:
		gs.CursorY = clamp(gs.CursorY+1, 0, gs.Diff.Height-1)
	case ttbox.KeyArrowLeft:
		gs.CursorX = clamp(gs.CursorX-1, 0, gs.Diff.Width-1)
	case ttbox.KeyArrowRight:
		gs.CursorX = clamp(gs.CursorX+1, 0, gs.Diff.Width-1)
	}

	// Vim-keys & Action keys
	switch evt.Ch {
	case ' ':
		if gs.Board[gs.CursorY][gs.CursorX].IsRevealed {
			chord(gs, gs.CursorX, gs.CursorY)
		} else {
			reveal(gs, gs.CursorX, gs.CursorY)
		}
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
}

func handleMouseEvent(gs *GameState, evt ttbox.Event) {
	if !evt.Press {
		return
	}

	termW, termH := ttbox.Size()
	diffBtn, resetBtn := getUIElements(gs, termW, termH)

	if evt.Button == ttbox.MouseLeft {
		if evt.X >= diffBtn.X && evt.X < diffBtn.X+diffBtn.W && evt.Y == diffBtn.Y {
			cycleDifficulty(gs)
			return
		}
		if evt.X >= resetBtn.X && evt.X < resetBtn.X+resetBtn.W && evt.Y == resetBtn.Y {
			gs.ResetGame(gs.Diff)
			return
		}
	}

	if gs.Status != StatusPlaying {
		return
	}

	boardX, boardY, ok := screenToBoard(gs, evt.X, evt.Y)
	if ok {
		switch evt.Button {
		case ttbox.MouseLeft:
			if gs.CursorX == boardX && gs.CursorY == boardY {
				if gs.Board[boardY][boardX].IsRevealed {
					chord(gs, boardX, boardY)
				} else {
					reveal(gs, boardX, boardY)
				}
			} else {
				gs.CursorX = boardX
				gs.CursorY = boardY
			}
		case ttbox.MouseRight:
			gs.CursorX = boardX
			gs.CursorY = boardY
			toggleFlag(gs, boardX, boardY)
		}
	} else {
		gs.CursorX = -1
		gs.CursorY = -1
	}
}

// cycleDifficulty rotates the current game difficulty through Easy, Medium, and Hard.
// It triggers an immediate game reset upon changing.
func cycleDifficulty(gs *GameState) {
	nextIndex := (gs.DiffIndex + 1) % len(Difficulties)
	gs.ResetGame(Difficulties[nextIndex])
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
