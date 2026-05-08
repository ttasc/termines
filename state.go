// This file defines the core data structures and state management for the game.
// It acts as the single source of truth for the application, holding the board,
// metrics, and cursor positions.
package main

import (
	"time"
)

// GameStatus represents the current lifecycle state of the game.
type GameStatus int

const (
	StatusPlaying GameStatus = iota
	StatusWon
	StatusLost
)

// Difficulty defines the grid dimensions and mine count for a game mode.
type Difficulty struct {
	Name   string
	Width  int
	Height int
	Mines  int
}

var (
	DiffEasy   = Difficulty{"EASY", 9, 9, 10}
	DiffMedium = Difficulty{"MEDIUM", 16, 16, 40}
	DiffHard   = Difficulty{"HARD", 30, 16, 99}
)
var Difficulties =[]Difficulty{DiffEasy, DiffMedium, DiffHard}

// Cell represents a single tile on the Minesweeper board.
type Cell struct {
	IsMine        bool
	IsRevealed    bool
	IsFlagged     bool
	IsExploded    bool  // Indicates the specific mine that triggered a loss state
	AdjacentMines uint8 // Pre-calculated count of neighboring mines (0-8)
}

// GameState holds the entire mutable state of the application.
type GameState struct {
	Board          [][]Cell
	Diff           Difficulty
	DiffIndex      int
	Status         GameStatus
	FlagsPlaced    int
	RevealedCount  int
	FirstClickDone bool          // Ensures the first reveal is always safe and starts the timer
	StartTime      time.Time
	CursorX        int
	CursorY        int
}

// NewGameState initializes a new game session with the given difficulty.
func NewGameState(diff Difficulty) *GameState {
	gs := &GameState{}
	gs.ResetGame(diff)
	return gs
}

// ResetGame clears the current state, resets timers/counters, and reallocates the board.
func (gs *GameState) ResetGame(diff Difficulty) {
	gs.Diff = diff

	for i, d := range Difficulties {
		if d.Name == diff.Name {
			gs.DiffIndex = i
			break
		}
	}

	gs.Status = StatusPlaying
	gs.FlagsPlaced = 0
	gs.RevealedCount = 0
	gs.FirstClickDone = false

	// Set cursor to -1 to hide it until the user interacts via keyboard or mouse.
	gs.CursorX = -1
	gs.CursorY = -1

	// Only reuse if the array already exists and is the same size as the new difficulty level.
	if len(gs.Board) == diff.Height && len(gs.Board) > 0 && len(gs.Board[0]) == diff.Width {
		for y := range gs.Board {
			for x := range gs.Board[y] {
				gs.Board[y][x] = Cell{}
			}
		}
	} else {
		gs.Board = make([][]Cell, diff.Height)
		for i := range gs.Board {
			gs.Board[i] = make([]Cell, diff.Width)
		}
	}
}
