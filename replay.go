package main

import (
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
)

type gameReplay struct {
	gInfo gameInfo
	gData gameData
	rInfo replayInfo
}

type gameInfo struct {
	Id           uuid.UUID
	GameDuration time.Duration
	Result       string
	MineCount    int
	FieldWidth   int
	FieldHeight  int
	CreatedAt    time.Time
}

type gameData struct {
	Id      uuid.UUID
	Field   [][]fieldCell
	History []historyStep
}

type replayInfo struct {
	currStepIdx int

	currX       int
	currY       int
	currScrollX int
	currScrollY int

	lastMPress     time.Time
	autoplayActive bool
	stopAutoplay   chan struct{}
}

// modifies gData.Field
func (a *app) createReplayInfo(gData gameData) replayInfo {
	closeField(gData.Field)

	currX, currY, currScrollX, currScrollY := a.stepsFromZeroTo(gData.Field, gData.History, len(gData.History)-1)

	return replayInfo{
		currStepIdx: len(gData.History) - 1,

		currX:       currX,
		currY:       currY,
		currScrollX: currScrollX,
		currScrollY: currScrollY,

		lastMPress:     time.Now().Add(-time.Minute),
		stopAutoplay:   nil,
		autoplayActive: false,
	}
}

func createGameInfo(result string, history []historyStep, field [][]fieldCell) gameInfo {
	id := uuid.New()

	return gameInfo{
		Id:           id,
		Result:       result,
		GameDuration: history[len(history)-1].CurrGameDuration,
		MineCount:    totalMineCount(field),
		FieldWidth:   len(field[0]),
		FieldHeight:  len(field),
		CreatedAt:    time.Now(),
	}
}

func (a *app) drawReplay() {
	currStart := 0
	resultStr := a.replay.gInfo.Result
	a.setContentString(currStart, 0, a.defStyle, resultStr)
	currStart += len(resultStr) + 3

	currStepIdx := max(a.replay.rInfo.currStepIdx, 0)
	currTimeStr := strconv.Itoa(int(a.replay.gData.History[currStepIdx].CurrGameDuration.Seconds()))
	gameDurationStr := strconv.Itoa(int(a.replay.gData.History[len(a.replay.gData.History)-1].CurrGameDuration.Seconds()))
	timeStr := "Time:" + currTimeStr + "/" + gameDurationStr
	a.setContentString(currStart, 0, a.defStyle, timeStr)
	currStart += len(timeStr) + 3

	widthStr := strconv.Itoa(a.replay.gInfo.FieldWidth)
	heightStr := strconv.Itoa(a.replay.gInfo.FieldHeight)
	mineCountStr := strconv.Itoa(a.replay.gInfo.MineCount)
	minesLeftStr := strconv.Itoa(minesLeft(a.replay.gData.Field))
	fieldInfoStr := widthStr + "x" + heightStr + "(" + minesLeftStr + "/" + mineCountStr + ")"
	a.setContentString(currStart, 0, a.defStyle, fieldInfoStr)
	currStart += len(fieldInfoStr) + 3

	currStepStr := strconv.Itoa(a.replay.rInfo.currStepIdx + 1)
	maxStepStr := strconv.Itoa(len(a.replay.gData.History))
	stepStr := "Step:" + currStepStr + "/" + maxStepStr
	a.setContentString(currStart, 0, a.defStyle, stepStr)
	currStart += len(stepStr) + 3

	dateStr := a.replay.gInfo.CreatedAt.Format("2006-01-02 15:04:05")
	a.setContentString(currStart, 0, a.defStyle, dateStr)
	currStart += len(dateStr) + 3

	a.drawField(a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY, a.replay.gData.Field)
}

func (a *app) eventKeyReplay(ev *tcell.EventKey) {
	key := ev.Key()
	rune := ev.Rune()

	if a.replay.rInfo.autoplayActive {
		if key == tcell.KeyEscape || rune == 'q' {
			a.replay.rInfo.autoplayActive = false
			close(a.replay.rInfo.stopAutoplay)

			a.cancel()
		}

		if rune == 'p' {
			a.replay.rInfo.autoplayActive = false
			close(a.replay.rInfo.stopAutoplay)
		}

		if rune == 'r' {
			a.replay.rInfo.autoplayActive = false
			close(a.replay.rInfo.stopAutoplay)

			a.play = createPlay(a.replay.gInfo.FieldWidth, a.replay.gInfo.FieldHeight, a.replay.gInfo.MineCount)
			a.state = "PLAY"
		}

		if rune == 'b' {
			a.replay.rInfo.autoplayActive = false
			close(a.replay.rInfo.stopAutoplay)

			if a.menu.menuState == "SAVED_GAMES" && a.menu.savedGamesState == "FIND" {
				savedGames, err := a.loadGameInfos(
					a.menu.savedGamesPrepareSortByState,
					a.menu.savedGamesPrepareFieldState == "ALL",
					a.menu.savedGamesPrepareFieldCustomWidth,
					a.menu.savedGamesPrepareFieldCustomHeight,
					a.menu.savedGamesPrepareFieldCustomMineCount,
				)
				if err != nil {
					a.log(err)
					a.cancel()
					return
				}

				a.menu.savedGames = savedGames
				a.menu.savedGamesFindCurr = 0
				a.menu.savedGamesFindScreenOffset = 0
			}

			a.state = "MENU"
		}

		return
	}

	if key == tcell.KeyEscape || rune == 'q' {
		a.cancel()
		return
	}

	if rune == 'm' {
		if time.Since(a.replay.rInfo.lastMPress).Abs() < time.Second/2 {
			err := a.deleteGame(a.replay.gInfo.Id)
			if err != nil {
				a.log(err)
				a.cancel()
				return

			}

			savedGames, err := a.loadGameInfos(
				a.menu.savedGamesPrepareSortByState,
				a.menu.savedGamesPrepareFieldState == "ALL",
				a.menu.savedGamesPrepareFieldCustomWidth,
				a.menu.savedGamesPrepareFieldCustomHeight,
				a.menu.savedGamesPrepareFieldCustomMineCount,
			)
			if err != nil {
				a.log(err)
				a.cancel()
				return
			}

			a.menu.savedGames = savedGames
			a.menu.savedGamesFindCurr = 0
			a.menu.savedGamesFindScreenOffset = 0
			a.state = "MENU"
		} else {
			a.replay.rInfo.lastMPress = time.Now()
		}
	}

	if rune == 'p' {
		a.replay.rInfo.stopAutoplay = make(chan struct{})
		a.replay.rInfo.autoplayActive = true
		safeGo(a.autoplayReplay, a.screen)
	}

	if rune == 'l' || key == tcell.KeyRight {
		a.replay.rInfo.currStepIdx++
		if a.replay.rInfo.currStepIdx >= len(a.replay.gData.History) {
			a.replay.rInfo.currStepIdx = -1
		}

		a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.stepsFromZeroTo(a.replay.gData.Field, a.replay.gData.History, a.replay.rInfo.currStepIdx)
	}

	if rune == 'h' || key == tcell.KeyLeft {
		a.replay.rInfo.currStepIdx--
		if a.replay.rInfo.currStepIdx < -1 {
			a.replay.rInfo.currStepIdx = len(a.replay.gData.History) - 1
		}

		a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.stepsFromZeroTo(a.replay.gData.Field, a.replay.gData.History, a.replay.rInfo.currStepIdx)
	}
	if rune == 'j' || key == tcell.KeyDown {
		a.replay.rInfo.currStepIdx = -1
		a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.stepsFromZeroTo(a.replay.gData.Field, a.replay.gData.History, a.replay.rInfo.currStepIdx)
	}
	if rune == 'k' || key == tcell.KeyUp {
		a.replay.rInfo.currStepIdx = len(a.replay.gData.History) - 1
		a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.stepsFromZeroTo(a.replay.gData.Field, a.replay.gData.History, a.replay.rInfo.currStepIdx)
	}

	if rune == 'i' {
		if a.replay.rInfo.currScrollY > 0 {
			a.replay.rInfo.currScrollY--
		}
	}
	if rune == 'u' {
		a.replay.rInfo.currScrollY++
	}
	if rune == 'y' {
		if a.replay.rInfo.currScrollX > 0 {
			a.replay.rInfo.currScrollX--
		}
	}
	if rune == 'o' {
		a.replay.rInfo.currScrollX++
	}

	if rune == 'I' {
		a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.scrollField(true, "UP", a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY)
	}
	if rune == 'U' {
		a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.scrollField(true, "DOWN", a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY)
	}
	if rune == 'Y' {
		a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.scrollField(true, "LEFT", a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY)
	}
	if rune == 'O' {
		a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.scrollField(true, "RIGHT", a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY)
	}

	switch rune {
	case 'r':
		a.play = createPlay(a.replay.gInfo.FieldWidth, a.replay.gInfo.FieldHeight, a.replay.gInfo.MineCount)
		a.state = "PLAY"
	case 'b':
		if a.menu.menuState == "SAVED_GAMES" && a.menu.savedGamesState == "FIND" {
			savedGames, err := a.loadGameInfos(
				a.menu.savedGamesPrepareSortByState,
				a.menu.savedGamesPrepareFieldState == "ALL",
				a.menu.savedGamesPrepareFieldCustomWidth,
				a.menu.savedGamesPrepareFieldCustomHeight,
				a.menu.savedGamesPrepareFieldCustomMineCount,
			)
			if err != nil {
				a.log(err)
				a.cancel()
				return
			}

			a.menu.savedGames = savedGames
			a.menu.savedGamesFindCurr = 0
			a.menu.savedGamesFindScreenOffset = 0
		}
		a.state = "MENU"
	}
}

func (a *app) autoplayReplay() {
	a.wg.Add(1)
	defer a.wg.Done()
	if a.replay.rInfo.currStepIdx == len(a.replay.gData.History)-1 {
		a.replay.rInfo.currStepIdx = -1
	}

	if a.replay.rInfo.currStepIdx < 0 {
		a.replay.rInfo.currStepIdx = 0
		a.screen.Clear()
		a.draw()
		a.screen.Show()
	}

	for step := a.replay.rInfo.currStepIdx; step <= len(a.replay.gData.History)-1; step++ {
		select {
		case <-a.ctx.Done():
			return
		case <-a.replay.rInfo.stopAutoplay:
			return
		default:
			var timer *time.Timer
			if step > 0 {
				timer = time.NewTimer((a.replay.gData.History[step].CurrGameDuration - a.replay.gData.History[step-1].CurrGameDuration).Abs())
			} else {
				timer = time.NewTimer(time.Duration(0))
			}

			select {
			case <-a.ctx.Done():
				return
			case <-a.replay.rInfo.stopAutoplay:
				return
			case <-timer.C:
				a.replay.rInfo.currStepIdx = step
				a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.stepsFromZeroTo(a.replay.gData.Field, a.replay.gData.History, a.replay.rInfo.currStepIdx)

				a.screen.Clear()
				a.draw()
				a.screen.Show()
			}
		}
	}

	a.replay.rInfo.autoplayActive = false
}

// modifies field
func (a *app) stepsFromZeroTo(field [][]fieldCell, history []historyStep, idx int) (x int, y int, scrollX int, scrollY int) {
	closeField(field)
	for i := 0; i <= idx; i++ {
		x, y, scrollX, scrollY = a.nextStep(field, x, y, scrollX, scrollY, history[i])
	}

	return
}

// modifies field
func (a *app) nextStep(field [][]fieldCell, fieldCurrX, fieldCurrY, fieldCurrScrollX, fieldCurrScrollY int, step historyStep) (x int, y int, scrollX int, scrollY int) {
	switch step.Kind {
	case "MOVE":
		fieldCurrScrollX, fieldCurrScrollY = a.alignField(step.MoveX, step.MoveY, fieldCurrScrollX, fieldCurrScrollY)
		return step.MoveX, step.MoveY, fieldCurrScrollX, fieldCurrScrollY
	case "OPEN":
		result := openField(field, fieldCurrX, fieldCurrY)
		if result != step.OpenResult {
			a.log("ERROR: result is not equal to step.OpenResult. This should never happen!!!")
			a.cancel()
		}

		if step.OpenResult != "NONE" {
			openFieldMines(field)
		}
	case "FLAG":
		flagField(field, fieldCurrX, fieldCurrY)
	}

	return fieldCurrX, fieldCurrY, fieldCurrScrollX, fieldCurrScrollY
}
