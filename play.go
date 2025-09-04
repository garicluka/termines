package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

type play struct {
	fieldCurrX       int
	fieldCurrY       int
	fieldCurrScrollX int
	fieldCurrScrollY int
	startingStats    bool
	lastQPress       time.Time
	history          []historyStep
	field            [][]fieldCell
	startTime        time.Time
	started          bool
	timeChan         chan struct{}
	lastSPress       time.Time
}

func createPlay(width, height, mineCount int) play {
	var field [][]fieldCell
	field = createField(width, height, mineCount)

	return play{
		fieldCurrX:       0,
		fieldCurrY:       0,
		fieldCurrScrollX: 0,
		fieldCurrScrollY: 0,
		lastQPress:       time.Now().Add(-time.Minute),
		history:          []historyStep{},
		field:            field,
		started:          false,
		lastSPress:       time.Now().Add(-time.Minute),
	}
}

func (a *app) drawPlay() {
	if a.play.startingStats {
		width := len(a.play.field[0])
		height := len(a.play.field)
		mineCount := totalMineCount(a.play.field)
		fieldWidthStr := strconv.Itoa(width)
		fieldHeightStr := strconv.Itoa(height)
		mineCountStr := strconv.Itoa(mineCount)
		mineDensityStr := fmt.Sprintf("%.2f%%", getMineDensity(width, height, mineCount))
		startingStatsStr := fieldWidthStr + "x" + fieldHeightStr + "(" + mineCountStr + ")" + " " + mineDensityStr
		a.setContentString(0, 0, a.defStyle, startingStatsStr)
	} else {
		currStart := 0
		minesLeftStr := "Mines Left:" + strconv.Itoa(minesLeft(a.play.field))
		a.setContentString(currStart, 0, a.defStyle, minesLeftStr)
		currStart += len(minesLeftStr) + 3

		var timePassed time.Duration
		if a.play.started {
			timePassed = time.Since(a.play.startTime)
		} else {
			timePassed = 0
		}

		secondsStr := "Time:" + strconv.Itoa(int(timePassed.Seconds()))
		a.setContentString(currStart, 0, a.defStyle, secondsStr)
		currStart += len(secondsStr) + 3
	}

	a.drawField(a.play.fieldCurrX, a.play.fieldCurrY, a.play.fieldCurrScrollX, a.play.fieldCurrScrollY, a.play.field)
}

func (a *app) eventKeyPlay(ev *tcell.EventKey) {
	rune := ev.Rune()
	key := ev.Key()

	if rune == 'q' {
		if time.Since(a.play.lastQPress).Abs() < time.Second/2 {
			if a.play.started {
				a.play.started = false
				close(a.play.timeChan)
			}
			a.state = "MENU"
		} else {
			a.play.lastQPress = time.Now()
		}
		return
	}

	if rune == 's' {
		a.play.lastSPress = time.Now()
		return
	}

	bigArrowMove := false
	if time.Since(a.play.lastSPress).Abs() < time.Second/2 {
		bigArrowMove = true
		a.play.lastSPress = time.Now().Add(-time.Minute)
	}

	originalFieldCurrX := a.play.fieldCurrX
	originalFieldCurrY := a.play.fieldCurrY

	if rune == '?' {
		a.play.startingStats = !a.play.startingStats
	}

	if rune == 'k' || (key == tcell.KeyUp && !bigArrowMove) {
		a.play.fieldCurrX, a.play.fieldCurrY = moveField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY, false, "UP", 1)
	}
	if rune == 'j' || (key == tcell.KeyDown && !bigArrowMove) {
		a.play.fieldCurrX, a.play.fieldCurrY = moveField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY, false, "DOWN", 1)
	}
	if rune == 'h' || (key == tcell.KeyLeft && !bigArrowMove) {
		a.play.fieldCurrX, a.play.fieldCurrY = moveField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY, false, "LEFT", 1)
	}
	if rune == 'l' || (key == tcell.KeyRight && !bigArrowMove) {
		a.play.fieldCurrX, a.play.fieldCurrY = moveField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY, false, "RIGHT", 1)
	}

	if rune == 'K' || (key == tcell.KeyUp && bigArrowMove) {
		a.play.fieldCurrX, a.play.fieldCurrY = moveField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY, true, "UP", 0)
	}
	if rune == 'J' || (key == tcell.KeyDown && bigArrowMove) {
		a.play.fieldCurrX, a.play.fieldCurrY = moveField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY, true, "DOWN", 0)
	}
	if rune == 'H' || (key == tcell.KeyLeft && bigArrowMove) {
		a.play.fieldCurrX, a.play.fieldCurrY = moveField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY, true, "LEFT", 0)
	}
	if rune == 'L' || (key == tcell.KeyRight && bigArrowMove) {
		a.play.fieldCurrX, a.play.fieldCurrY = moveField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY, true, "RIGHT", 0)
	}

	switch rune {
	case 'd', 'D':
		if !a.play.started {
			a.startGame()

			a.play.history = append(a.play.history, historyStep{
				CurrGameDuration: time.Duration(0),
				Kind:             "MOVE",
				MoveX:            a.play.fieldCurrX,
				MoveY:            a.play.fieldCurrY,
			})
		}

		result := openField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY)

		a.play.history = append(a.play.history, historyStep{
			CurrGameDuration: time.Since(a.play.startTime),
			Kind:             "OPEN",
			OpenResult:       result,
		})

		if result != "NONE" {
			a.play.started = false
			close(a.play.timeChan)
			gInfo := createGameInfo(result, a.play.history, a.play.field)

			a.replay.gInfo = gInfo
			a.replay.gData = gameData{
				Field:   closeFieldCopy(a.play.field),
				History: a.play.history,
			}
			a.replay.rInfo = a.createReplayInfo(a.replay.gData)

			a.state = "REPLAY"

			historyCopy := make([]historyStep, len(a.play.history))
			copy(historyCopy, a.play.history)
			fieldCopy := closeFieldCopy(a.play.field)

			safeGo(func() {
				a.saveGame(gInfo, gameData{
					Id:      gInfo.Id,
					Field:   fieldCopy,
					History: historyCopy,
				})
			}, a.screen)
		}
	case 'f', 'F':
		if !a.play.started {
			a.startGame()
			a.play.history = append(a.play.history, historyStep{
				CurrGameDuration: time.Duration(0),
				Kind:             "MOVE",
				MoveX:            a.play.fieldCurrX,
				MoveY:            a.play.fieldCurrY,
			})
		}

		a.play.history = append(a.play.history, historyStep{
			CurrGameDuration: time.Since(a.play.startTime),
			Kind:             "FLAG",
		})

		flagField(a.play.field, a.play.fieldCurrX, a.play.fieldCurrY)
	case 'i', 'I', 'u', 'U', 'y', 'Y', 'o', 'O':
		var movement string
		var bigScroll bool
		if rune == 'i' || rune == 'I' {
			movement = "UP"
		}
		if rune == 'u' || rune == 'U' {
			movement = "DOWN"
		}
		if rune == 'y' || rune == 'Y' {
			movement = "LEFT"
		}
		if rune == 'o' || rune == 'O' {
			movement = "RIGHT"
		}

		if rune == 'I' || rune == 'U' || rune == 'Y' || rune == 'O' {
			bigScroll = true
		} else {
			bigScroll = false
		}

		a.play.fieldCurrScrollX, a.play.fieldCurrScrollY = a.scrollField(bigScroll, movement, a.play.fieldCurrX, a.play.fieldCurrY, a.play.fieldCurrScrollX, a.play.fieldCurrScrollY)
	}

	if rune == 'k' || rune == 'K' || key == tcell.KeyUp ||
		rune == 'j' || rune == 'J' || key == tcell.KeyDown ||
		rune == 'h' || rune == 'H' || key == tcell.KeyLeft ||
		rune == 'l' || rune == 'L' || key == tcell.KeyRight {
		a.play.fieldCurrScrollX, a.play.fieldCurrScrollY = a.alignField(a.play.fieldCurrX, a.play.fieldCurrY, a.play.fieldCurrScrollX, a.play.fieldCurrScrollY)

		if a.play.started {
			if originalFieldCurrX != a.play.fieldCurrX || originalFieldCurrY != a.play.fieldCurrY {
				a.play.history = append(a.play.history, historyStep{
					CurrGameDuration: time.Since(a.play.startTime),
					Kind:             "MOVE",
					MoveX:            a.play.fieldCurrX,
					MoveY:            a.play.fieldCurrY,
				})
			}
		}
	}
}

func (a *app) startGame() {
	a.play.startTime = time.Now()
	ticker := time.NewTicker(time.Second)
	quit := make(chan struct{})
	a.play.started = true
	a.play.timeChan = quit

	a.wg.Add(1)
	safeGo(func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				return
			case <-ticker.C:
				a.screen.Clear()
				a.draw()
				a.screen.Show()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}, a.screen)
}
