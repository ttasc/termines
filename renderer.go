// This file manages the terminal UI rendering pipeline.
// It translates the abstract GameState into character and color arrays,
// handling dynamic layout centering and visual state hierarchies.
package main

import (
	"fmt"
	"time"

	"github.com/ttasc/ttbox"
)

// Terminal characters used for rendering board states.
const (
	CharUnrevealed   = '■'
	CharFlag         = 'F'
	CharMine         = '*'
	CharWrongFlag    = 'X'
	CharLeftBracket  = '['
	CharRightBracket = ']'
)

// 256-color ANSI palette constants.
const (
	ColorBoardGrid = 239
	ColorBgActive  = 236
	ColorBtnBg     = 237

	ColorFlag        = 196
	ColorMine        = 196
	ColorMineExplode = 196

	ColorHoverValid   = 39
	ColorHoverInvalid = 196

	ColorWin  = 114
	ColorLose = 196

	ColorText    = 250
	ColorTextDim = 240
	ColorBgModal = 235
)

// NumberColors maps a cell's adjacent mine count (1-8) to specific ANSI colors.
var NumberColors =[]int{
	0, 39, 114, 220, 208, 196, 201, 135, 240,
}

// UIButton defines a clickable bounding box in absolute terminal coordinates.
type UIButton struct {
	X, Y, W, H int
}

// render executes the complete drawing pipeline for a single frame.
func render(gs *GameState) {
	ttbox.Clear()
	termW, termH := ttbox.Size()

	drawStatusline(gs, termW, termH)
	drawBoard(gs, termW, termH)
	drawResetButton(gs, termW, termH)

	if gs.Status != StatusPlaying {
		drawEndgameBanner(gs, termW, termH)
	} else {
		drawControlsGuide(termH)
	}

	ttbox.Present()
}

// getUIElements calculates the dynamic layout of top and bottom UI buttons.
// Recalculated every frame to handle real-time terminal window resizing.
func getUIElements(gs *GameState, termW, termH int) (diffBtn, resetBtn UIButton) {
	_, offsetY := getBoardOffsets(gs, termW, termH)
	centerX := termW / 2

	headerY := max(offsetY-2, 0)
	timerText := formatTime(gs)
	diffText := fmt.Sprintf(" [ %s ] ", gs.Diff.Name)

	diffBtn = UIButton{
		X: centerX - (len(timerText) / 2) - len(diffText),
		Y: headerY,
		W: len(diffText),
		H: 1,
	}

	resetY := offsetY + gs.Diff.Height + 1
	// Cap the reset button position to ensure it doesn't overlap the bottom 2 banner lines.
	if resetY >= termH-2 {
		resetY = max(termH-3, 0)
	}

	resetText := " [ RESET ] "
	resetBtn = UIButton{
		X: centerX - (len(resetText) / 2),
		Y: resetY,
		W: len(resetText),
		H: 1,
	}

	return diffBtn, resetBtn
}

// getBoardOffsets calculates the top-left coordinate needed to center the grid.
// Important: Cells are rendered 3 characters wide to achieve a square visual aspect ratio.
func getBoardOffsets(gs *GameState, termW, termH int) (int, int) {
	boardPixelWidth := gs.Diff.Width * 3
	offsetX := max((termW-boardPixelWidth)/2, 0)
	offsetY := max((termH-gs.Diff.Height)/2, 0)
	return offsetX, offsetY
}

// drawStatusline renders the timer, difficulty button, and remaining flag count.
func drawStatusline(gs *GameState, termW, termH int) {
	diffBtn, _ := getUIElements(gs, termW, termH)
	y := diffBtn.Y

	if y != 0 {
		ttbox.DrawTextCenter(1, " T E R M I N E S ", ColorText, ttbox.ColorDefault)
	}

	centerX := termW / 2

	timerText := formatTime(gs)
	ttbox.DrawTextCenter(y, timerText, 255, ttbox.ColorDefault)

	diffText := fmt.Sprintf(" [ %s ] ", gs.Diff.Name)
	drawButton(diffBtn.X, y, diffText, ColorHoverValid, ColorBtnBg)

	minesLeft := gs.Diff.Mines - gs.FlagsPlaced
	flagText := fmt.Sprintf(" %d Flags ", minesLeft)
	flagX := centerX + (len(timerText) / 2) + (len(timerText) % 2)

	for i, ch := range flagText {
		fg := 255
		if ch == CharFlag {
			fg = ColorFlag
		}
		ttbox.SetCell(flagX+i, y, ch, fg, ColorBgActive)
	}
}

// drawResetButton renders the reset toggle below the board.
func drawResetButton(gs *GameState, termW, termH int) {
	_, resetBtn := getUIElements(gs, termW, termH)
	drawButton(resetBtn.X, resetBtn.Y, " [ RESET ] ", 255, ColorBtnBg)
}

// drawControlsGuide renders static keybinding instructions at the bottom edge.
func drawControlsGuide(termH int) {
	guideText := " REVEAL/CHORD:l-click,space,enter | FLAG:r-click,f | MOVE:h,j,k,l;arrows | RESET:r | MODE:1,2,3 "
	y := termH - 1

	ttbox.DrawTextCenter(y, guideText, ColorText, ttbox.ColorDefault)
}

// drawEndgameBanner displays the game result prominently at the bottom of the screen without covering the board.
func drawEndgameBanner(gs *GameState, termW, termH int) {
	if termW == 0 || termH == 0 {
		return
	}

	msgFg := ColorWin
	msg := " * YOU WIN! * "
	if gs.Status == StatusLost {
		msgFg = ColorLose
		msg = " * GAME OVER * "
	}
	subMsg := " [R] Play Again   [ESC] Exit "

	// Draw the main result message.
	ttbox.SetAttr(true, false, false, false)
	ttbox.DrawTextCenter(termH-2, msg, msgFg, ttbox.ColorDefault)
	ttbox.ResetAttr()

	// Draw the sub-message for key actions.
	ttbox.DrawTextCenter(termH-1, subMsg, ColorTextDim, ttbox.ColorDefault)
}

// drawBoard iterates over the 2D game state and renders each cell.
func drawBoard(gs *GameState, termW, termH int) {
	offsetX, offsetY := getBoardOffsets(gs, termW, termH)

	for y := 0; y < gs.Diff.Height; y++ {
		for x := 0; x < gs.Diff.Width; x++ {

			ch, fg, bg, leftChar, rightChar, bracketFg := getCellStyle(gs, x, y)

			// Project grid coordinates (x, y) to absolute terminal coordinates.
			// Each grid cell occupies 3 horizontal terminal cells (e.g., "[1]").
			screenX := offsetX + (x * 3)
			screenY := offsetY + y

			ttbox.SetCell(screenX, screenY, leftChar, bracketFg, bg)
			ttbox.SetCell(screenX+1, screenY, ch, fg, bg)
			ttbox.SetCell(screenX+2, screenY, rightChar, bracketFg, bg)
		}
	}
}

// getCellStyle resolves the competing visual states of a single cell into characters and colors.
// Priority: Game Over validations > Exploded Mines > Flags > Revealed Numbers > Unrevealed.
func getCellStyle(gs *GameState, x, y int) (ch rune, fg int, bg int, leftChar rune, rightChar rune, bracketFg int) {
	cell := gs.Board[y][x]
	ch = CharUnrevealed
	fg = ColorBoardGrid
	bg = ttbox.ColorDefault
	leftChar, rightChar = ' ', ' '
	bracketFg = ColorHoverValid

	if gs.Status == StatusLost && cell.IsFlagged && !cell.IsMine {
		// Identify falsely placed flags upon game over.
		ch, fg, bg = CharWrongFlag, 255, ColorMineExplode
	} else if cell.IsFlagged {
		ch, fg = CharFlag, ColorFlag
	} else if cell.IsRevealed {
		if cell.IsMine {
			ch, fg = CharMine, ColorMine
			if cell.IsExploded {
				// Highlight the specific mine that triggered the game over.
				bg, fg = ColorMineExplode, 255
			}
		} else if cell.AdjacentMines > 0 {
			ch = rune('0' + cell.AdjacentMines)
			fg = NumberColors[cell.AdjacentMines]
		} else {
			ch = ' '
		}
	}

	// Apply cursor highlighting overlay (bracket chars) on top of the base cell state.
	isCursor := (gs.CursorX == x && gs.CursorY == y)
	if isCursor {
		leftChar, rightChar = CharLeftBracket, CharRightBracket
		if cell.IsRevealed {
			// Red brackets indicate an invalid action (cannot interact with revealed cells).
			bracketFg = ColorHoverInvalid
		}
		if !cell.IsRevealed && !cell.IsFlagged {
			// Brighten the core block character when hovered.
			fg = 255
		}
	}

	return ch, fg, bg, leftChar, rightChar, bracketFg
}

// drawButton writes a string to the terminal at (x,y) applying specific colors.
func drawButton(x, y int, text string, fg, bg int) {
	for i, ch := range text {
		ttbox.SetCell(x+i, y, ch, fg, bg)
	}
}

// formatTime formats elapsed seconds into HH:MM:SS.
func formatTime(gs *GameState) string {
	var h, m, s int
	if gs.FirstClickDone {
		elapsed := time.Since(gs.StartTime)
		h = int(elapsed.Hours())
		m = int(elapsed.Minutes()) % 60
		s = int(elapsed.Seconds()) % 60
	}
	return fmt.Sprintf("  %02d:%02d:%02d  ", h, m, s)
}
