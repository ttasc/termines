// This file implements the core gameplay logic of Minesweeper, including
// board generation, state transitions (win/loss), and cell interaction mechanics.
package main

import (
	"math/rand"
	"time"
)

// abs returns the absolute value of an integer.
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// generateMines populates the board with mines and calculates adjacent mine counts.
// It ensures the first clicked cell (cx, cy) and its immediate neighbors are always safe.
func generateMines(gs *GameState, cx, cy int) {
	w, h := gs.Diff.Width, gs.Diff.Height

	// Prevent infinite loops by capping mines to available cells minus the 3x3 safe zone.
	maxPossibleMines := max((w*h)-9, 0)
	targetMines := min(gs.Diff.Mines, maxPossibleMines)

	placed := 0
	for placed < targetMines {
		rx := rand.Intn(w)
		ry := rand.Intn(h)

		if gs.Board[ry][rx].IsMine {
			continue
		}

		// Enforce a 3x3 safe zone around the initial click coordinates.
		if abs(rx-cx) <= 1 && abs(ry-cy) <= 1 {
			continue
		}

		gs.Board[ry][rx].IsMine = true
		placed++
	}

	// Pre-calculate adjacent mine counts for all non-mine cells.
	// This single O(N*M) pass prevents dynamic calculation overhead during rendering.
	for y := range h {
		for x := range w {
			if gs.Board[y][x].IsMine {
				continue
			}

			count := uint8(0)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy

					if nx >= 0 && nx < w && ny >= 0 && ny < h {
						if gs.Board[ny][nx].IsMine {
							count++
						}
					}
				}
			}
			gs.Board[y][x].AdjacentMines = count
		}
	}
}

// point represents a 2D coordinate used primarily for the BFS queue.
type point struct {
	x, y int
}

// checkWin verifies if all non-mine cells have been successfully revealed.
// If true, it transitions the game state to won and automatically flags remaining mines.
func checkWin(gs *GameState) {
	totalCells := gs.Diff.Width * gs.Diff.Height
	safeCells := totalCells - gs.Diff.Mines

	if gs.RevealedCount == safeCells {
		gs.Status = StatusWon

		for y := 0; y < gs.Diff.Height; y++ {
			for x := 0; x < gs.Diff.Width; x++ {
				cell := &gs.Board[y][x]
				if cell.IsMine && !cell.IsFlagged {
					cell.IsFlagged = true
					gs.FlagsPlaced++
				}
			}
		}
	}
}

// reveal uncovers a cell at (x, y) and updates the game state.
// If the cell is empty (0 adjacent mines), it triggers a flood-fill to reveal neighbors.
func reveal(gs *GameState, x, y int) {
	if gs.Status != StatusPlaying {
		return
	}

	cell := &gs.Board[y][x]

	// Ignore clicks on already revealed or flagged cells.
	if cell.IsRevealed || cell.IsFlagged {
		return
	}

	// Defer board generation until the first click to guarantee a safe start.
	if !gs.FirstClickDone {
		generateMines(gs, x, y)
		gs.FirstClickDone = true
		gs.StartTime = time.Now()
	}

	// Hitting a mine triggers the loss state and reveals all mines.
	if cell.IsMine {
		gs.Status = StatusLost
		cell.IsExploded = true

		for by := 0; by < gs.Diff.Height; by++ {
			for bx := 0; bx < gs.Diff.Width; bx++ {
				c := &gs.Board[by][bx]
				if c.IsMine {
					c.IsRevealed = true
				}
			}
		}
		return
	}

	// Breadth-First Search (BFS) flood-fill algorithm.
	// Iteratively reveals empty neighboring cells.
	queue :=[]point{{x, y}}
	w, h := gs.Diff.Width, gs.Diff.Height

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		cx, cy := curr.x, curr.y
		currCell := &gs.Board[cy][cx]

		// Skip already processed cells to prevent infinite loops.
		if currCell.IsRevealed {
			continue
		}

		currCell.IsRevealed = true
		gs.RevealedCount++

		// Stop expanding if this cell borders a mine.
		if currCell.AdjacentMines > 0 {
			continue
		}

		// Enqueue all valid, unrevealed, and unflagged neighbors.
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := cx+dx, cy+dy
				if nx >= 0 && nx < w && ny >= 0 && ny < h {
					neighbor := &gs.Board[ny][nx]
					if !neighbor.IsRevealed && !neighbor.IsFlagged {
						queue = append(queue, point{nx, ny})
					}
				}
			}
		}
	}

	checkWin(gs)
}

// toggleFlag places or removes a flag on an unrevealed cell.
func toggleFlag(gs *GameState, x, y int) {
	if gs.Status != StatusPlaying {
		return
	}

	gs.Board[y][x].IsFlagged = !gs.Board[y][x].IsFlagged
}
