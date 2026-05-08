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

	maxPossibleMines := max((w*h)-9, 0)
	targetMines := min(gs.Diff.Mines, maxPossibleMines)

	// Collect all safe coordinates (except the 3x3 area around the first click).
	validCoords := make([]point, 0, w*h)
	for y := range h {
		for x := range w {
			// Ignore the 3x3 safe zone.
			if abs(x-cx) <= 1 && abs(y-cy) <= 1 {
				continue
			}
			validCoords = append(validCoords, point{x, y})
		}
	}

	// Shuffle a valid array of coordinates.
	rand.Shuffle(len(validCoords), func(i, j int) {
		validCoords[i], validCoords[j] = validCoords[j], validCoords[i]
	})

	// Place the mine at the first coordinate in `targetMines` after mixing.
	for i := range targetMines {
		p := validCoords[i]
		gs.Board[p.y][p.x].IsMine = true
	}

	// Pre-calculate adjacent mine counts for all non-mine cells.
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
	w, h := gs.Diff.Width, gs.Diff.Height
	// Allocate approximately 25% of the total number of cells in the table in terms of capacity.
	queue := make([]point, 0, (w*h)/4)
	queue = append(queue, point{x, y})

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

	cell := &gs.Board[y][x]

	if cell.IsRevealed {
		return
	}

	cell.IsFlagged = !cell.IsFlagged

	if cell.IsFlagged {
		gs.FlagsPlaced++
	} else {
		gs.FlagsPlaced--
	}
}

// chord tự động mở các ô xung quanh nếu số cờ xung quanh bằng đúng với số của ô hiện tại.
func chord(gs *GameState, x, y int) {
	if gs.Status != StatusPlaying {
		return
	}

	cell := &gs.Board[y][x]

	// Chording chỉ áp dụng cho những ô đã mở và có số (AdjacentMines > 0)
	if !cell.IsRevealed || cell.AdjacentMines == 0 {
		return
	}

	w, h := gs.Diff.Width, gs.Diff.Height
	flagCount := uint8(0)

	// Bước 1: Đếm số lượng cờ thực tế đang cắm xung quanh ô này (trong phạm vi 3x3)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < w && ny >= 0 && ny < h {
				if gs.Board[ny][nx].IsFlagged {
					flagCount++
				}
			}
		}
	}

	// Bước 2: Nếu số cờ bằng đúng với giá trị của ô số
	if flagCount == cell.AdjacentMines {
		// Duyệt lại vùng 3x3 một lần nữa để mở các ô chưa mở
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < w && ny >= 0 && ny < h {
					neighbor := &gs.Board[ny][nx]
					// Chỉ mở những ô chưa mở và chưa bị cắm cờ
					if !neighbor.IsRevealed && !neighbor.IsFlagged {
						reveal(gs, nx, ny)
					}
				}
			}
		}
	}
}
