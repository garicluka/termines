package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

type menu struct {
	// SELECT,PLAY,SAVED_GAMES,SETTINGS
	menuState string

	// PLAY,SAVED_GAMES,SETTINGS
	selectState string

	// WIDTH, HEIGHT, MINE_COUNT
	playState     string
	playWidth     int
	playHeight    int
	playMineCount int

	// PREPARE, FIND
	savedGamesState string
	savedGames      []gameInfo
	// SORT_BY, FIELD
	savedGamesPrepareState string
	// LATEST, OLDEST, BEST, WORST
	savedGamesPrepareSortByState string
	// ALL, CUSTOM
	savedGamesPrepareFieldState string
	// WIDTH, HEIGHT, MINE_COUNT
	savedGamesPrepareFieldCustomState     string
	savedGamesPrepareFieldCustomWidth     int
	savedGamesPrepareFieldCustomHeight    int
	savedGamesPrepareFieldCustomMineCount int
	savedGamesFindCurr                    int
	savedGamesFindScreenOffset            int
	savedGamesFindLastMPress              time.Time
	savedGamesFindLastMPressIndex         int

	// THEME, MAX_SCROLLOFF
	settingsState string
	// DEFAULT, LIGHT, DARK, MONO
	settingsThemeState   string
	settingsMaxScrolloff int
}

func (a *app) createMenu() {
	a.menu = menu{
		menuState: "SELECT",

		selectState: "PLAY",

		playState:     "WIDTH",
		playWidth:     0,
		playHeight:    0,
		playMineCount: 0,

		savedGamesState:                       "PREPARE",
		savedGames:                            []gameInfo{},
		savedGamesPrepareState:                "SORT_BY",
		savedGamesPrepareSortByState:          "LATEST",
		savedGamesPrepareFieldState:           "ALL",
		savedGamesPrepareFieldCustomState:     "WIDTH",
		savedGamesPrepareFieldCustomWidth:     0,
		savedGamesPrepareFieldCustomHeight:    0,
		savedGamesPrepareFieldCustomMineCount: 0,
		savedGamesFindCurr:                    0,
		savedGamesFindScreenOffset:            0,
		savedGamesFindLastMPress:              time.Now().Add(-time.Minute),
		savedGamesFindLastMPressIndex:         -1,

		settingsState:        "THEME",
		settingsThemeState:   a.settings.Theme,
		settingsMaxScrolloff: a.settings.MaxScrolloff,
	}
}

func (a *app) drawMenu() {
	switch a.menu.menuState {
	case "SELECT":
		a.drawMenuSelect()
	case "PLAY":
		a.drawMenuPlay()
	case "SAVED_GAMES":
		a.drawMenuSavedGames()
	case "SETTINGS":
		a.drawMenuSettings()
	}
}

func (a *app) eventKeyMenu(ev *tcell.EventKey) {
	key := ev.Key()
	rune := ev.Rune()

	if key == tcell.KeyEscape || rune == 'q' {
		a.cancel()
		return
	}

	switch a.menu.menuState {
	case "SELECT":
		a.eventKeyMenuSelect(key, rune)
	case "PLAY":
		a.eventKeyMenuPlay(key, rune)
	case "SAVED_GAMES":
		a.eventKeyMenuSavedGames(key, rune)
	case "SETTINGS":
		a.eventKeyMenuSettings(key, rune)
	}
}

func (a *app) drawMenuSelect() {
	a.setContentString(0, 0, a.defStyle, "Termines")
	a.setContentString(0, 1, a.defStyle, "Play")
	a.setContentString(0, 2, a.defStyle, "Saved Games")
	a.setContentString(0, 3, a.defStyle, "Settings")

	switch a.menu.selectState {
	case "PLAY":
		a.setContentString(0, 1, a.defStyle.Reverse(true), "Play")
	case "SAVED_GAMES":
		a.setContentString(0, 2, a.defStyle.Reverse(true), "Saved Games")
	case "SETTINGS":
		a.setContentString(0, 3, a.defStyle.Reverse(true), "Settings")
	}
}

func (a *app) drawMenuPlay() {
	currStart := 0

	playStr := "Play"
	a.setContentString(currStart, 0, a.defStyle, playStr)
	currStart += len(playStr) + 1

	isValid := isValidPlay(a.menu.playWidth, a.menu.playHeight, a.menu.playMineCount)
	var validStr string
	if isValid {
		validStr = "VALID"
	} else {
		validStr = "INVALID"
	}
	a.setContentString(currStart, 0, a.defStyle, validStr)
	currStart += len(validStr) + 1

	if isValid {
		mineDensity := getMineDensity(a.menu.playWidth, a.menu.playHeight, a.menu.playMineCount)
		mineDensityStr := fmt.Sprintf("%.2f%%", mineDensity)
		a.setContentString(currStart, 0, a.defStyle, mineDensityStr)
	}

	widthStr := ""
	if a.menu.playWidth > 0 {
		widthStr = strconv.Itoa(a.menu.playWidth)
	}
	widthStr = "Width:" + widthStr
	a.setContentString(0, 1, a.defStyle, widthStr)

	heightStr := ""
	if a.menu.playHeight > 0 {
		heightStr = strconv.Itoa(a.menu.playHeight)
	}
	heightStr = "Height:" + heightStr
	a.setContentString(0, 2, a.defStyle, heightStr)

	mineCountStr := ""
	if a.menu.playMineCount > 0 {
		mineCountStr = strconv.Itoa(a.menu.playMineCount)
	}
	mineCountStr = "Mine Count:" + mineCountStr
	a.setContentString(0, 3, a.defStyle, mineCountStr)

	switch a.menu.playState {
	case "WIDTH":
		a.screen.SetContent(len(widthStr), 1, ' ', nil, a.defStyle.Reverse(true))
	case "HEIGHT":
		a.screen.SetContent(len(heightStr), 2, ' ', nil, a.defStyle.Reverse(true))
	case "MINE_COUNT":
		a.screen.SetContent(len(mineCountStr), 3, ' ', nil, a.defStyle.Reverse(true))
	}
}

func (a *app) drawMenuSavedGames() {
	if a.menu.savedGamesState == "PREPARE" {
		a.drawMenuSavedGamesPrepare()
	} else {
		a.drawMenuSavedGamesFind()
	}
}

func (a *app) drawMenuSavedGamesPrepare() {
	a.setContentString(0, 0, a.defStyle, "Saved Games")

	// SORT BY
	a.setContentString(0, 1, a.defStyle, "Sort By:")
	if a.menu.savedGamesPrepareState == "SORT_BY" {
		a.setContentString(0, 1, a.defStyle.Reverse(true), "Sort By:")
	}

	currStart := 0

	latestStr := "Latest"
	a.setContentString(currStart, 2, a.defStyle, latestStr)
	if a.menu.savedGamesPrepareSortByState == "LATEST" {
		a.setContentString(currStart, 2, a.defStyle.Reverse(true), latestStr)
	}
	currStart += len(latestStr) + 1

	oldestStr := "Oldest"
	a.setContentString(currStart, 2, a.defStyle, oldestStr)
	if a.menu.savedGamesPrepareSortByState == "OLDEST" {
		a.setContentString(currStart, 2, a.defStyle.Reverse(true), oldestStr)
	}
	currStart += len(oldestStr) + 1

	bestStr := "Best"
	a.setContentString(currStart, 2, a.defStyle, bestStr)
	if a.menu.savedGamesPrepareSortByState == "BEST" {
		a.setContentString(currStart, 2, a.defStyle.Reverse(true), bestStr)
	}
	currStart += len(bestStr) + 1

	worstStr := "Worst"
	a.setContentString(currStart, 2, a.defStyle, worstStr)
	if a.menu.savedGamesPrepareSortByState == "WORST" {
		a.setContentString(currStart, 2, a.defStyle.Reverse(true), worstStr)

	}
	currStart += len(worstStr) + 1

	// FIELD
	a.setContentString(0, 3, a.defStyle, "Field:")
	if a.menu.savedGamesPrepareState == "FIELD" {
		a.setContentString(0, 3, a.defStyle.Reverse(true), "Field:")
	}

	currStart = 0

	allStr := "All"
	a.setContentString(currStart, 4, a.defStyle, allStr)
	if a.menu.savedGamesPrepareFieldState == "ALL" {
		a.setContentString(currStart, 4, a.defStyle.Reverse(true), allStr)
	}
	currStart += len(allStr) + 1

	customStr := "Custom"
	a.setContentString(currStart, 4, a.defStyle, customStr)
	if a.menu.savedGamesPrepareFieldState == "CUSTOM" {
		a.setContentString(currStart, 4, a.defStyle.Reverse(true), customStr)
	}
	currStart += len(customStr) + 1

	if a.menu.savedGamesPrepareFieldState == "CUSTOM" {
		a.setContentString(4, 4, a.defStyle.Reverse(true), "Custom")

		widthStr := ""
		if a.menu.savedGamesPrepareFieldCustomWidth > 0 {
			widthStr = strconv.Itoa(a.menu.savedGamesPrepareFieldCustomWidth)
		}
		widthStr = "Width:" + widthStr
		a.setContentString(0, 5, a.defStyle, widthStr)

		heightStr := ""
		if a.menu.savedGamesPrepareFieldCustomHeight > 0 {
			heightStr = strconv.Itoa(a.menu.savedGamesPrepareFieldCustomHeight)
		}
		heightStr = "Height:" + heightStr
		a.setContentString(0, 6, a.defStyle, heightStr)

		mineCountStr := ""
		if a.menu.savedGamesPrepareFieldCustomMineCount > 0 {
			mineCountStr = strconv.Itoa(a.menu.savedGamesPrepareFieldCustomMineCount)
		}
		mineCountStr = "Mine Count:" + mineCountStr
		a.setContentString(0, 7, a.defStyle, mineCountStr)

		if a.menu.savedGamesPrepareState == "FIELD" {
			switch a.menu.savedGamesPrepareFieldCustomState {
			case "WIDTH":
				a.screen.SetContent(len(widthStr), 5, ' ', nil, a.defStyle.Reverse(true))
			case "HEIGHT":
				a.screen.SetContent(len(heightStr), 6, ' ', nil, a.defStyle.Reverse(true))
			case "MINE_COUNT":
				a.screen.SetContent(len(mineCountStr), 7, ' ', nil, a.defStyle.Reverse(true))
			}
		}
	}
}

func (a *app) drawMenuSavedGamesFind() {
	infoStr := ""
	if a.menu.savedGamesPrepareFieldState == "ALL" {
		infoStr += "ALL"
	} else {
		width := strconv.Itoa(a.menu.savedGamesPrepareFieldCustomWidth)
		height := strconv.Itoa(a.menu.savedGamesPrepareFieldCustomHeight)
		mineCount := strconv.Itoa(a.menu.savedGamesPrepareFieldCustomMineCount)
		infoStr += width + "x" + height + "(" + mineCount + ")"
	}
	infoStr += " "
	infoStr += a.menu.savedGamesPrepareSortByState
	a.setContentString(0, 0, a.defStyle, "Saved Games"+" "+infoStr)
	_, screenHeight := a.screen.Size()
	screenHeight -= 1

	for i := range screenHeight {
		idx := i + a.menu.savedGamesFindScreenOffset
		if idx < len(a.menu.savedGames) {
			v := a.menu.savedGames[idx]

			num := strconv.Itoa(idx + 1)
			width := strconv.Itoa(v.FieldWidth)
			height := strconv.Itoa(v.FieldHeight)
			mineCount := strconv.Itoa(v.MineCount)
			time := strconv.Itoa(int(v.GameDuration.Seconds()))

			date := v.CreatedAt.Format("2006-01-02 15:04:05")

			str := num + "." + v.Result + " " + width + "x" + height + "(" + mineCount + ")" + " " + time + " " + date

			if idx == a.menu.savedGamesFindCurr {
				a.setContentString(0, i+1, a.defStyle.Reverse(true), str)
			} else {
				a.setContentString(0, i+1, a.defStyle, str)
			}
		}
	}
}

func (a *app) drawMenuSettings() {
	a.setContentString(0, 0, a.defStyle, "Settings")

	a.setContentString(0, 1, a.defStyle, "Theme")
	a.setContentString(0, 3, a.defStyle, "MAX_SCROLLOFF")
	switch a.menu.settingsState {
	case "THEME":
		a.setContentString(0, 1, a.defStyle.Reverse(true), "Theme")
	case "MAX_SCROLLOFF":
		a.setContentString(0, 3, a.defStyle.Reverse(true), "MAX_SCROLLOFF")
	}

	currStart := 0

	defaultThemeStr := "Default"
	a.setContentString(currStart, 2, a.defStyle, defaultThemeStr)
	if a.menu.settingsThemeState == "DEFAULT" {
		a.setContentString(currStart, 2, a.defStyle.Reverse(true), defaultThemeStr)
	}
	currStart += len(defaultThemeStr) + 1

	lightThemeStr := "Light"
	a.setContentString(currStart, 2, a.defStyle, lightThemeStr)
	if a.menu.settingsThemeState == "LIGHT" {
		a.setContentString(currStart, 2, a.defStyle.Reverse(true), lightThemeStr)
	}
	currStart += len(lightThemeStr) + 1

	darkThemeStr := "Dark"
	a.setContentString(currStart, 2, a.defStyle, darkThemeStr)
	if a.menu.settingsThemeState == "DARK" {
		a.setContentString(currStart, 2, a.defStyle.Reverse(true), darkThemeStr)
	}
	currStart += len(darkThemeStr) + 1

	monoThemeStr := "Mono"
	a.setContentString(currStart, 2, a.defStyle, monoThemeStr)
	if a.menu.settingsThemeState == "MONO" {
		a.setContentString(currStart, 2, a.defStyle.Reverse(true), monoThemeStr)
	}
	currStart += len(monoThemeStr) + 1

	maxScrolloffStr := strconv.Itoa(a.menu.settingsMaxScrolloff)
	a.setContentString(0, 4, a.defStyle, maxScrolloffStr)
}

func (a *app) eventKeyMenuSelect(key tcell.Key, rune rune) {
	if key == tcell.KeyEnter || key == tcell.KeyTab || rune == ' ' || rune == 'd' {
		switch a.menu.selectState {
		case "PLAY":
			a.menu.menuState = "PLAY"
		case "SAVED_GAMES":
			a.menu.savedGamesState = "PREPARE"
			a.menu.menuState = "SAVED_GAMES"
		case "SETTINGS":
			a.menu.menuState = "SETTINGS"
		}
	}

	if rune == 'j' || key == tcell.KeyDown {
		switch a.menu.selectState {
		case "PLAY":
			a.menu.selectState = "SAVED_GAMES"
		case "SAVED_GAMES":
			a.menu.selectState = "SETTINGS"
		case "SETTINGS":
			a.menu.selectState = "PLAY"
		}
	}

	if rune == 'k' || key == tcell.KeyUp {
		switch a.menu.selectState {
		case "PLAY":
			a.menu.selectState = "SETTINGS"
		case "SAVED_GAMES":
			a.menu.selectState = "PLAY"
		case "SETTINGS":
			a.menu.selectState = "SAVED_GAMES"
		}
	}
}

func (a *app) eventKeyMenuPlay(key tcell.Key, rune rune) {
	if key == tcell.KeyBackspace || key == tcell.KeyBackspace2 {
		switch a.menu.playState {
		case "WIDTH":
			a.menu.playWidth = intRemoveLast(a.menu.playWidth)
		case "HEIGHT":
			a.menu.playHeight = intRemoveLast(a.menu.playHeight)
		case "MINE_COUNT":
			a.menu.playMineCount = intRemoveLast(a.menu.playMineCount)
		}
	}

	if rune == 'b' {
		a.menu.menuState = "SELECT"
		return
	}

	if key == tcell.KeyEnter || key == tcell.KeyTab || rune == ' ' || rune == 'd' {
		validPlay := isValidPlay(a.menu.playWidth, a.menu.playHeight, a.menu.playMineCount)
		switch a.menu.playState {
		case "WIDTH":
			if validPlay {
				a.play = createPlay(a.menu.playWidth, a.menu.playHeight, a.menu.playMineCount)
				a.state = "PLAY"
				break
			}
			if a.menu.playWidth == 0 {
				break
			}

			if a.menu.playHeight == 0 {
				a.menu.playState = "HEIGHT"
			} else if a.menu.playMineCount == 0 {
				a.menu.playState = "MINE_COUNT"
			}
		case "HEIGHT":
			if validPlay {
				a.play = createPlay(a.menu.playWidth, a.menu.playHeight, a.menu.playMineCount)
				a.state = "PLAY"
				break
			}
			if a.menu.playHeight == 0 {
				break
			}
			if a.menu.playMineCount == 0 {
				a.menu.playState = "MINE_COUNT"
			} else if a.menu.playWidth == 0 {
				a.menu.playState = "WIDTH"
			}
		case "MINE_COUNT":
			if validPlay {
				a.play = createPlay(a.menu.playWidth, a.menu.playHeight, a.menu.playMineCount)
				a.state = "PLAY"
				break
			}
			if a.menu.playMineCount == 0 {
				break
			}
			if a.menu.playWidth == 0 {
				a.menu.playState = "WIDTH"
			} else if a.menu.playHeight == 0 {
				a.menu.playState = "HEIGHT"
			}
		}
	}

	if rune == 'j' || key == tcell.KeyDown {
		switch a.menu.playState {
		case "WIDTH":
			a.menu.playState = "HEIGHT"
		case "HEIGHT":
			a.menu.playState = "MINE_COUNT"
		case "MINE_COUNT":
			a.menu.playState = "WIDTH"
		}
	}

	if rune == 'k' || key == tcell.KeyUp {
		switch a.menu.playState {
		case "WIDTH":
			a.menu.playState = "MINE_COUNT"
		case "HEIGHT":
			a.menu.playState = "WIDTH"
		case "MINE_COUNT":
			a.menu.playState = "HEIGHT"
		}
	}

	switch rune {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		switch a.menu.playState {
		case "WIDTH":
			a.menu.playWidth = intToIntAdd(a.menu.playWidth, rune)
		case "HEIGHT":
			a.menu.playHeight = intToIntAdd(a.menu.playHeight, rune)
		case "MINE_COUNT":
			a.menu.playMineCount = intToIntAdd(a.menu.playMineCount, rune)
		}
	}
}

func (a *app) eventKeyMenuSavedGames(key tcell.Key, rune rune) {
	if a.menu.savedGamesState == "PREPARE" {
		a.eventKeyMenuSavedGamesPrepare(key, rune)
	} else {
		a.eventKeyMenuSavedGamesFind(key, rune)
	}
}

func (a *app) eventKeyMenuSavedGamesPrepare(key tcell.Key, rune rune) {
	if key == tcell.KeyBackspace || key == tcell.KeyBackspace2 {
		if a.menu.savedGamesPrepareState == "FIELD" && a.menu.savedGamesPrepareFieldState == "CUSTOM" {
			switch a.menu.savedGamesPrepareFieldCustomState {
			case "WIDTH":
				a.menu.savedGamesPrepareFieldCustomWidth = intRemoveLast(a.menu.savedGamesPrepareFieldCustomWidth)
			case "HEIGHT":
				a.menu.savedGamesPrepareFieldCustomHeight = intRemoveLast(a.menu.savedGamesPrepareFieldCustomHeight)
			case "MINE_COUNT":
				a.menu.savedGamesPrepareFieldCustomMineCount = intRemoveLast(a.menu.savedGamesPrepareFieldCustomMineCount)
			}

		}
	}

	if rune == 'b' {
		a.menu.menuState = "SELECT"
		return
	}

	if key == tcell.KeyEnter || key == tcell.KeyTab || rune == ' ' || rune == 'd' {
		validInfo :=
			(a.menu.savedGamesPrepareFieldCustomWidth != 0 &&
				a.menu.savedGamesPrepareFieldCustomHeight != 0 &&
				a.menu.savedGamesPrepareFieldCustomMineCount != 0) ||
				a.menu.savedGamesPrepareFieldState == "ALL"

		if validInfo {
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
			a.menu.savedGamesState = "FIND"
			return
		}

		if a.menu.savedGamesPrepareState != "FIELD" {
			if a.menu.savedGamesPrepareFieldCustomWidth == 0 {
				a.menu.savedGamesPrepareFieldCustomState = "WIDTH"
				a.menu.savedGamesPrepareState = "FIELD"
			} else if a.menu.savedGamesPrepareFieldCustomHeight == 0 {
				a.menu.savedGamesPrepareFieldCustomState = "HEIGHT"
				a.menu.savedGamesPrepareState = "FIELD"
			} else {
				a.menu.savedGamesPrepareFieldCustomState = "MINE_COUNT"
				a.menu.savedGamesPrepareState = "FIELD"
			}
		} else {
			switch a.menu.savedGamesPrepareFieldCustomState {
			case "WIDTH":
				if a.menu.savedGamesPrepareFieldCustomWidth == 0 {
					break
				}
				if a.menu.savedGamesPrepareFieldCustomHeight == 0 {
					a.menu.savedGamesPrepareFieldCustomState = "HEIGHT"
					a.menu.savedGamesPrepareState = "FIELD"
				} else {
					a.menu.savedGamesPrepareFieldCustomState = "MINE_COUNT"
					a.menu.savedGamesPrepareState = "FIELD"
				}
			case "HEIGHT":
				if a.menu.savedGamesPrepareFieldCustomHeight == 0 {
					break
				}
				if a.menu.savedGamesPrepareFieldCustomWidth == 0 {
					a.menu.savedGamesPrepareFieldCustomState = "WIDTH"
					a.menu.savedGamesPrepareState = "FIELD"
				} else {
					a.menu.savedGamesPrepareFieldCustomState = "MINE_COUNT"
					a.menu.savedGamesPrepareState = "FIELD"
				}
			case "MINE_COUNT":
				if a.menu.savedGamesPrepareFieldCustomMineCount == 0 {
					break
				}
				if a.menu.savedGamesPrepareFieldCustomWidth == 0 {
					a.menu.savedGamesPrepareFieldCustomState = "WIDTH"
					a.menu.savedGamesPrepareState = "FIELD"
				} else {
					a.menu.savedGamesPrepareFieldCustomState = "HEIGHT"
					a.menu.savedGamesPrepareState = "FIELD"
				}
			}
		}
	}

	if rune == 'k' || key == tcell.KeyUp {
		switch a.menu.savedGamesPrepareState {
		case "SORT_BY":
			a.menu.savedGamesPrepareFieldCustomState = "MINE_COUNT"
			a.menu.savedGamesPrepareState = "FIELD"
		case "FIELD":
			if a.menu.savedGamesPrepareFieldState == "CUSTOM" {
				switch a.menu.savedGamesPrepareFieldCustomState {
				case "WIDTH":
					a.menu.savedGamesPrepareState = "SORT_BY"
				case "HEIGHT":
					a.menu.savedGamesPrepareFieldCustomState = "WIDTH"
				case "MINE_COUNT":
					a.menu.savedGamesPrepareFieldCustomState = "HEIGHT"
				}
			} else {
				a.menu.savedGamesPrepareState = "SORT_BY"
			}
		}
	}

	if rune == 'j' || key == tcell.KeyDown {
		switch a.menu.savedGamesPrepareState {
		case "SORT_BY":
			a.menu.savedGamesPrepareFieldCustomState = "WIDTH"
			a.menu.savedGamesPrepareState = "FIELD"
		case "FIELD":
			if a.menu.savedGamesPrepareFieldState == "CUSTOM" {
				switch a.menu.savedGamesPrepareFieldCustomState {
				case "WIDTH":
					a.menu.savedGamesPrepareFieldCustomState = "HEIGHT"
				case "HEIGHT":
					a.menu.savedGamesPrepareFieldCustomState = "MINE_COUNT"
				case "MINE_COUNT":
					a.menu.savedGamesPrepareState = "SORT_BY"
				}
			} else {
				a.menu.savedGamesPrepareState = "SORT_BY"
			}
		}
	}

	if rune == 'l' || key == tcell.KeyRight {
		switch a.menu.savedGamesPrepareState {
		case "SORT_BY":
			switch a.menu.savedGamesPrepareSortByState {
			case "LATEST":
				a.menu.savedGamesPrepareSortByState = "OLDEST"
			case "OLDEST":
				a.menu.savedGamesPrepareSortByState = "BEST"
			case "BEST":
				a.menu.savedGamesPrepareSortByState = "WORST"
			case "WORST":
				a.menu.savedGamesPrepareSortByState = "LATEST"
			}
		case "FIELD":
			switch a.menu.savedGamesPrepareFieldState {
			case "ALL":
				a.menu.savedGamesPrepareFieldCustomState = "WIDTH"
				a.menu.savedGamesPrepareFieldState = "CUSTOM"
			case "CUSTOM":
				a.menu.savedGamesPrepareFieldState = "ALL"
			}
		}
	}

	if rune == 'h' || key == tcell.KeyLeft {
		switch a.menu.savedGamesPrepareState {
		case "SORT_BY":
			switch a.menu.savedGamesPrepareSortByState {
			case "LATEST":
				a.menu.savedGamesPrepareSortByState = "WORST"
			case "OLDEST":
				a.menu.savedGamesPrepareSortByState = "LATEST"
			case "BEST":
				a.menu.savedGamesPrepareSortByState = "OLDEST"
			case "WORST":
				a.menu.savedGamesPrepareSortByState = "BEST"
			}
		case "FIELD":
			switch a.menu.savedGamesPrepareFieldState {
			case "ALL":
				a.menu.savedGamesPrepareFieldCustomState = "WIDTH"
				a.menu.savedGamesPrepareFieldState = "CUSTOM"
			case "CUSTOM":
				a.menu.savedGamesPrepareFieldState = "ALL"
			}
		}
	}

	switch rune {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if a.menu.savedGamesPrepareState == "FIELD" && a.menu.savedGamesPrepareFieldState == "CUSTOM" {
			switch a.menu.savedGamesPrepareFieldCustomState {
			case "WIDTH":
				a.menu.savedGamesPrepareFieldCustomWidth = intToIntAdd(a.menu.savedGamesPrepareFieldCustomWidth, rune)
			case "HEIGHT":
				a.menu.savedGamesPrepareFieldCustomHeight = intToIntAdd(a.menu.savedGamesPrepareFieldCustomHeight, rune)
			case "MINE_COUNT":
				a.menu.savedGamesPrepareFieldCustomMineCount = intToIntAdd(a.menu.savedGamesPrepareFieldCustomMineCount, rune)
			}
		}
	}
}

func (a *app) eventKeyMenuSavedGamesFind(key tcell.Key, rune rune) {
	if key == tcell.KeyEnter || key == tcell.KeyTab || rune == ' ' || rune == 'd' {
		if len(a.menu.savedGames) > 0 {
			i, d, err := a.loadGameInfoAndData(a.menu.savedGames[a.menu.savedGamesFindCurr].Id)
			if err != nil {
				a.log(err)
				a.cancel()
			}
			a.replay.gInfo = i
			a.replay.gData = d

			a.replay.rInfo = a.createReplayInfo(a.replay.gData)
			a.state = "REPLAY"
		}
	}
	if rune == 'm' {
		if time.Since(a.menu.savedGamesFindLastMPress).Abs() < time.Second/2 &&
			a.menu.savedGamesFindLastMPressIndex == a.menu.savedGamesFindCurr {
			if len(a.menu.savedGames) == 0 {
				return
			}

			err := a.deleteGame(a.menu.savedGames[a.menu.savedGamesFindCurr].Id)
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
			a.menu.savedGamesFindCurr = min(a.menu.savedGamesFindCurr, len(savedGames)-1)
			a.menu.savedGamesFindCurr = max(a.menu.savedGamesFindCurr, 0)
		} else {
			a.menu.savedGamesFindLastMPress = time.Now()
			a.menu.savedGamesFindLastMPressIndex = a.menu.savedGamesFindCurr
		}
	}

	if rune == 'b' {
		a.menu.savedGamesState = "PREPARE"
		return
	}

	if rune == 'j' || key == tcell.KeyDown {
		_, screenHeight := a.screen.Size()
		screenHeight -= 1

		if len(a.menu.savedGames) == 0 {
			return
		}

		if a.menu.savedGamesFindCurr == len(a.menu.savedGames)-1 {
			a.menu.savedGamesFindCurr = 0
			a.menu.savedGamesFindScreenOffset = 0
		} else {
			a.menu.savedGamesFindCurr++
			screenIdx := a.menu.savedGamesFindCurr - a.menu.savedGamesFindScreenOffset
			if screenIdx >= screenHeight {
				a.menu.savedGamesFindScreenOffset++
			}
		}
	}

	if rune == 'k' || key == tcell.KeyUp {
		_, screenHeight := a.screen.Size()
		screenHeight -= 1
		if len(a.menu.savedGames) == 0 {
			return
		}

		if a.menu.savedGamesFindCurr == 0 {
			if len(a.menu.savedGames) >= screenHeight {
				a.menu.savedGamesFindCurr = len(a.menu.savedGames) - 1
				a.menu.savedGamesFindScreenOffset = len(a.menu.savedGames) - screenHeight
			} else {
				a.menu.savedGamesFindCurr = len(a.menu.savedGames) - 1
				a.menu.savedGamesFindScreenOffset = 0
			}
		} else {
			a.menu.savedGamesFindCurr--
			screenIdx := a.menu.savedGamesFindCurr - a.menu.savedGamesFindScreenOffset
			if screenIdx < 0 {
				a.menu.savedGamesFindScreenOffset--
			}
		}
	}
}

func (a *app) eventKeyMenuSettings(key tcell.Key, rune rune) {
	if key == tcell.KeyEnter || key == tcell.KeyTab || rune == ' ' || rune == 'd' {
		newSettings := settings{
			Theme:        a.menu.settingsThemeState,
			MaxScrolloff: a.menu.settingsMaxScrolloff,
		}

		err := a.updateSettings(newSettings)
		if err != nil {
			a.log(err)
			a.cancel()
			return
		}
		a.settings = newSettings

		defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
		switch newSettings.Theme {
		case "LIGHT":
			defStyle = tcell.StyleDefault.Background(tcell.Color254).Foreground(tcell.Color234)
		case "DARK":
			defStyle = tcell.StyleDefault.Background(tcell.Color234).Foreground(tcell.Color254)
		}
		a.defStyle = defStyle
		a.screen.SetStyle(defStyle)

		a.screen.Clear()
		a.draw()
		a.screen.Sync()

		a.menu.menuState = "SELECT"
		return
	}

	if rune == 'b' {
		a.menu.menuState = "SELECT"
		return
	}

	if rune == 'j' || key == tcell.KeyDown {
		switch a.menu.settingsState {
		case "THEME":
			a.menu.settingsState = "MAX_SCROLLOFF"
		case "MAX_SCROLLOFF":
			a.menu.settingsState = "THEME"
		}
	}

	if rune == 'k' || key == tcell.KeyUp {
		switch a.menu.settingsState {
		case "THEME":
			a.menu.settingsState = "MAX_SCROLLOFF"
		case "MAX_SCROLLOFF":
			a.menu.settingsState = "THEME"
		}
	}

	if rune == 'h' || key == tcell.KeyLeft {
		switch a.menu.settingsState {
		case "THEME":
			switch a.menu.settingsThemeState {
			case "DEFAULT":
				a.menu.settingsThemeState = "MONO"
			case "LIGHT":
				a.menu.settingsThemeState = "DEFAULT"
			case "DARK":
				a.menu.settingsThemeState = "LIGHT"
			case "MONO":
				a.menu.settingsThemeState = "DARK"
			}
		case "MAX_SCROLLOFF":
			if a.menu.settingsMaxScrolloff > 0 {
				a.menu.settingsMaxScrolloff--
			}
		}
	}

	if rune == 'l' || key == tcell.KeyRight {
		switch a.menu.settingsState {
		case "THEME":
			switch a.menu.settingsThemeState {
			case "DEFAULT":
				a.menu.settingsThemeState = "LIGHT"
			case "LIGHT":
				a.menu.settingsThemeState = "DARK"
			case "DARK":
				a.menu.settingsThemeState = "MONO"
			case "MONO":
				a.menu.settingsThemeState = "DEFAULT"
			}
		case "MAX_SCROLLOFF":
			a.menu.settingsMaxScrolloff++
		}
	}
}

func isValidPlay(width, height, mineCount int) bool {
	if width == 0 || height == 0 || mineCount == 0 {
		return false
	}

	if mineCount >= width*height {
		return false
	}

	return true
}

func intRemoveLast(original int) int {
	originalStr := strconv.Itoa(original)
	if len(originalStr) > 1 {
		originalStr = originalStr[0 : len(originalStr)-1]

		newInt, _ := strconv.ParseInt(originalStr, 10, 64)
		return int(newInt)
	} else {
		return 0
	}
}

func intToIntAdd(original int, newDigit rune) int {
	switch newDigit {
	case '0':
		if original > 0 {
			newInt, _ := strconv.ParseInt(strconv.Itoa(original)+string(newDigit), 10, 64)
			return int(newInt)
		}

	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		newInt, _ := strconv.ParseInt(strconv.Itoa(original)+string(newDigit), 10, 64)
		return int(newInt)

	}

	return original
}
