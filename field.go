package main

import (
	"math/rand/v2"

	"github.com/gdamore/tcell/v2"
)

const CELL_STATE_OPEN int = 1
const CELL_STATE_HIDDEN int = 0
const CELL_STATE_FLAG int = -1

const CELL_VALUE_MINE int = 9

type fieldCell struct {
	// 0-8 mine count 9 is mine
	Value int
	// -1 is flag, 0 is hidden, 1 is open
	State int
}

func createField(width, height, mineCount int) [][]fieldCell {
	field := make([][]fieldCell, height)
	for i := range height {
		field[i] = make([]fieldCell, width)
	}

	emptySpaces := len(field) * len(field[0])
	for mineCount > 0 && emptySpaces > 0 {
		y := rand.IntN(len(field))
		x := rand.IntN(len(field[0]))

		if field[y][x].Value == CELL_VALUE_MINE {
			continue
		}

		field[y][x].Value = CELL_VALUE_MINE
		mineCount--
		emptySpaces--
	}

	for y := range field {
		for x := range field[0] {
			if field[y][x].Value != CELL_VALUE_MINE {
				mine_count := 0
				for cx := -1; cx <= 1; cx++ {
					if x+cx < 0 || x+cx >= len(field[0]) {
						continue
					}
					for cy := -1; cy <= 1; cy++ {
						if y+cy < 0 || y+cy >= len(field) {
							continue
						}
						if field[y+cy][x+cx].Value == CELL_VALUE_MINE {
							mine_count++
						}
					}
				}
				field[y][x].Value = mine_count
			}
		}
	}

	return field
}

func (a *app) drawField(cursorX, cursorY, scrollX, scrollY int, field [][]fieldCell) {
	xOffset, yOffset, fieldScreenWidth, fieldScreenHeight := a.getFieldScreenSize()
	for fieldScreenY := range fieldScreenHeight {
		fieldY := fieldScreenY + scrollY
		if fieldY >= len(field) {
			break
		}
		for fieldScreenX := range fieldScreenWidth {
			fieldX := fieldScreenX + scrollX
			if fieldX >= len(field[0]) {
				break
			}

			rune, style := a.cellToStyle(field[fieldY][fieldX])
			a.screen.SetContent(fieldScreenX+xOffset, fieldScreenY+yOffset, rune, nil, style)

		}
	}

	screenX := cursorX - scrollX + xOffset
	screenY := cursorY - scrollY + yOffset
	if screenX >= xOffset && screenY >= yOffset && screenX < fieldScreenWidth+xOffset && screenY < fieldScreenHeight+yOffset {
		rune, style := a.cellToStyle(field[cursorY][cursorX])
		a.screen.SetContent(screenX, screenY, rune, nil, style.Reverse(true))
	}
}

func openFieldMines(field [][]fieldCell) {
	for y := range field {
		for x := range field[y] {
			if field[y][x].Value == CELL_VALUE_MINE && field[y][x].State == CELL_STATE_HIDDEN {
				field[y][x].State = CELL_STATE_OPEN
			}
		}
	}
}

func closeFieldCopy(field [][]fieldCell) [][]fieldCell {
	closedField := make([][]fieldCell, len(field))
	for y := range closedField {
		closedField[y] = make([]fieldCell, len(field[y]))
		for x := range closedField[y] {
			closedField[y][x].Value = field[y][x].Value
			closedField[y][x].State = CELL_STATE_HIDDEN
		}
	}

	return closedField
}

func closeField(field [][]fieldCell) {
	for y := range field {
		for x := range field[y] {
			field[y][x].State = CELL_STATE_HIDDEN
		}
	}
}

func openField(field [][]fieldCell, x, y int) string {
	switch field[y][x].State {
	case CELL_STATE_HIDDEN:
		if field[y][x].Value == CELL_VALUE_MINE {
			return "LOST"
		} else {
			field[y][x].State = CELL_STATE_OPEN
			fieldClearZeroesAt(field, x, y)
			if isWon(field) {
				return "WON"
			}
		}
	case CELL_STATE_OPEN:
		if fieldFlagsAroundCell(field, x, y) == field[y][x].Value {
			if !fieldCheckAroundNumber(field, x, y) {
				return "LOST"
			}
			openAroundNumber(field, x, y)
			if isWon(field) {
				return "WON"
			}
		}
	}

	return "NONE"
}

func flagField(field [][]fieldCell, x, y int) {
	switch field[y][x].State {
	case CELL_STATE_HIDDEN:
		field[y][x].State = CELL_STATE_FLAG
	case CELL_STATE_FLAG:
		field[y][x].State = CELL_STATE_HIDDEN
	}
}

func moveField(field [][]fieldCell, x, y int, bigMove bool, direction string, amount int) (int, int) {
	if bigMove {
		switch direction {
		case "UP":
			y = moveBigUp(field, x, y)
		case "DOWN":
			y = moveBigDown(field, x, y)
		case "LEFT":
			x = moveBigLeft(field, x, y)
		case "RIGHT":
			x = moveBigRight(field, x, y)
		}

		return x, y
	}

	switch direction {
	case "UP":
		y -= amount
		if y < 0 {
			y = 0
		}
	case "DOWN":
		y += amount
		if y >= len(field) {
			y = len(field) - 1
		}
	case "LEFT":
		x -= amount
		if x < 0 {
			x = 0
		}
	case "RIGHT":
		x += amount
		if x >= len(field[0]) {
			x = len(field[0]) - 1
		}
	}

	return x, y
}

func (a *app) scrollField(bigScroll bool, direction string, fieldX, fieldY, scrollX, scrollY int) (int, int) {
	_, _, fieldScreenWidth, fieldScreenHeight := a.getFieldScreenSize()
	scrolloffX := a.getFieldScrolloffX()
	scrolloffY := a.getFieldScrolloffY()

	if !bigScroll {
		switch direction {
		case "UP":
			if scrollY > 0 {
				scrollY--
			}
		case "DOWN":
			scrollY++
		case "LEFT":
			if scrollX > 0 {
				scrollX--
			}
		case "RIGHT":
			scrollX++
		}
		return scrollX, scrollY
	}

	switch direction {
	case "UP":
		scrollY = max(fieldY-fieldScreenHeight+scrolloffY+1, 0)
	case "DOWN":
		scrollY = max(fieldY-scrolloffY, 0)
	case "LEFT":
		scrollX = max(fieldX-fieldScreenWidth+scrolloffX+1, 0)
	case "RIGHT":
		scrollX = max(fieldX-scrolloffX, 0)
	}

	return scrollX, scrollY
}

func (a *app) alignField(fieldX, fieldY, scrollX, scrollY int) (newScrollX int, newScrollY int) {
	_, _, fieldScreenWidth, fieldScreenHeight := a.getFieldScreenSize()
	scrolloffX := a.getFieldScrolloffX()
	scrolloffY := a.getFieldScrolloffY()

	if fieldX-scrollX < scrolloffX {
		scrollX = max(fieldX-scrolloffX, 0)
	} else if fieldX-scrollX > fieldScreenWidth-scrolloffX-1 {
		scrollX = max(fieldX-fieldScreenWidth+scrolloffX+1, 0)
	}

	if fieldY-scrollY < scrolloffY {
		scrollY = max(fieldY-scrolloffY, 0)
	} else if fieldY-scrollY > fieldScreenHeight-scrolloffY-1 {
		scrollY = max(fieldY-fieldScreenHeight+scrolloffY+1, 0)
	}

	return scrollX, scrollY
}

func fieldClearZeroesAt(field [][]fieldCell, x, y int) {
	if field[y][x].Value == 0 {
		for cx := -1; cx <= 1; cx++ {
			if x+cx < 0 || x+cx >= len(field[0]) {
				continue
			}
			for cy := -1; cy <= 1; cy++ {
				if y+cy < 0 || y+cy >= len(field) {
					continue
				}
				if field[y+cy][x+cx].State == CELL_STATE_HIDDEN {
					field[y+cy][x+cx].State = CELL_STATE_OPEN
					if field[y+cy][x+cx].Value == 0 {
						fieldClearZeroesAt(field, x+cx, y+cy)
					}
				}
			}
		}
	}
}

func fieldFlagsAroundCell(field [][]fieldCell, x, y int) int {
	flagCount := 0

	for cx := -1; cx <= 1; cx++ {
		if x+cx < 0 || x+cx >= len(field[0]) {
			continue
		}
		for cy := -1; cy <= 1; cy++ {
			if y+cy < 0 || y+cy >= len(field) {
				continue
			}
			if field[y+cy][x+cx].State == CELL_STATE_FLAG {
				flagCount++
			}
		}
	}

	return flagCount
}

func fieldCheckAroundNumber(field [][]fieldCell, x, y int) bool {
	for cx := -1; cx <= 1; cx++ {
		if x+cx < 0 || x+cx >= len(field[0]) {
			continue
		}
		for cy := -1; cy <= 1; cy++ {
			if y+cy < 0 || y+cy >= len(field) {
				continue
			}
			if field[y+cy][x+cx].State == CELL_STATE_FLAG {
				if field[y+cy][x+cx].Value != CELL_VALUE_MINE {
					return false
				}
			} else {
				if field[y+cy][x+cx].Value == CELL_VALUE_MINE {
					return false
				}
			}
		}
	}

	return true
}

func openAroundNumber(field [][]fieldCell, x, y int) {
	for cx := -1; cx <= 1; cx++ {
		if x+cx < 0 || x+cx >= len(field[0]) {
			continue
		}
		for cy := -1; cy <= 1; cy++ {
			if y+cy < 0 || y+cy >= len(field) {
				continue
			}
			if field[y+cy][x+cx].State != CELL_STATE_FLAG {
				field[y+cy][x+cx].State = CELL_STATE_OPEN
				if field[y+cy][x+cx].Value == 0 {
					fieldClearZeroesAt(field, x+cx, y+cy)
				}
			}
		}
	}
}

func isWon(field [][]fieldCell) bool {
	for y := range field {
		for x := range field[0] {
			if field[y][x].State == CELL_STATE_HIDDEN || field[y][x].State == CELL_STATE_FLAG {
				if field[y][x].Value != CELL_VALUE_MINE {
					return false
				}
			}
		}
	}

	return true
}

func totalMineCount(field [][]fieldCell) int {
	mineCount := 0
	for _, v := range field {
		for _, v := range v {
			if v.Value == CELL_VALUE_MINE {
				mineCount++
			}
		}
	}

	return mineCount
}

func minesLeft(field [][]fieldCell) int {
	mineCount := 0
	for _, v := range field {
		for _, v := range v {
			if v.Value == CELL_VALUE_MINE {
				mineCount++
			}
			if v.Value == CELL_VALUE_MINE && v.State == CELL_STATE_OPEN {
				mineCount--
			}
			if v.State == CELL_STATE_FLAG {
				mineCount--
			}
		}
	}

	return mineCount
}

func moveBigUp(field [][]fieldCell, x, y int) int {
	if y <= 0 {
		return 0
	}

	onHiddenCell := field[y][x].State == CELL_STATE_HIDDEN
	if !onHiddenCell {
		y--

		for y > 0 && field[y][x].State != CELL_STATE_HIDDEN {
			y--
		}
		return y
	}

	aboveIsHiddenCell := field[y-1][x].State == CELL_STATE_HIDDEN
	if !aboveIsHiddenCell {
		y--
		return y
	}

	for y-1 >= 0 && field[y-1][x].State == CELL_STATE_HIDDEN {
		y--
	}

	return y
}

func moveBigDown(field [][]fieldCell, x, y int) int {
	if y >= len(field)-1 {
		return len(field) - 1
	}

	onHiddenCell := field[y][x].State == CELL_STATE_HIDDEN
	if !onHiddenCell {
		y++
		for y < len(field)-1 && field[y][x].State != CELL_STATE_HIDDEN {
			y++
		}
		return y
	}

	belowIsHiddenCell := field[y+1][x].State == CELL_STATE_HIDDEN
	if !belowIsHiddenCell {
		y++
		return y
	}

	for y+1 < len(field) && field[y+1][x].State == CELL_STATE_HIDDEN {
		y++
	}

	return y
}

func moveBigRight(field [][]fieldCell, x, y int) int {
	if x >= len(field[0])-1 {
		return len(field[0]) - 1
	}

	onHiddenCell := field[y][x].State == CELL_STATE_HIDDEN
	if !onHiddenCell {
		x++
		for x < len(field[0])-1 && field[y][x].State != CELL_STATE_HIDDEN {
			x++
		}
		return x
	}

	rightIsHiddenCell := field[y][x+1].State == CELL_STATE_HIDDEN
	if !rightIsHiddenCell {
		x++
		return x
	}

	for x+1 < len(field[0]) && field[y][x+1].State == CELL_STATE_HIDDEN {
		x++
	}

	return x
}

func moveBigLeft(field [][]fieldCell, x, y int) int {
	if x <= 0 {
		return 0
	}

	onHiddenCell := field[y][x].State == CELL_STATE_HIDDEN
	if !onHiddenCell {
		x--
		for x > 0 && field[y][x].State != CELL_STATE_HIDDEN {
			x--
		}
		return x
	}

	leftIsHiddenCell := field[y][x-1].State == CELL_STATE_HIDDEN
	if !leftIsHiddenCell {
		x--
		return x
	}

	for x-1 >= 0 && field[y][x-1].State == CELL_STATE_HIDDEN {
		x--
	}

	return x
}

func (a *app) cellToStyle(cell fieldCell) (rune, tcell.Style) {
	switch cell.State {
	case CELL_STATE_HIDDEN:
		return '-', a.defStyle
	case CELL_STATE_FLAG:
		if a.settings.Theme == "MONO" {
			return 'F', a.defStyle
		} else {
			return 'F', a.defStyle.Background(tcell.ColorRed)
		}
	}

	style := a.defStyle
	rune := ' '

	switch cell.Value {
	case 0:
		style = a.defStyle
		rune = '0'
	case 1:
		style = a.defStyle.Foreground(tcell.Color27)
		rune = '1'
	case 2:
		style = a.defStyle.Foreground(tcell.Color34)
		rune = '2'
	case 3:
		style = a.defStyle.Foreground(tcell.ColorRed)
		rune = '3'
	case 4:
		style = a.defStyle.Foreground(tcell.Color205)
		rune = '4'
	case 5:
		style = a.defStyle.Foreground(tcell.Color124)
		rune = '5'
	case 6:
		style = a.defStyle.Foreground(tcell.Color45)
		rune = '6'
	case 7:
		style = a.defStyle.Foreground(tcell.Color92)
		rune = '7'
	case 8:
		style = a.defStyle.Foreground(tcell.ColorGray)
		rune = '8'
	case 9:
		style = a.defStyle.Background(tcell.ColorRed)
		rune = 'M'
	}

	if a.settings.Theme == "MONO" {
		style = a.defStyle
	}

	return rune, style
}
