package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

type app struct {
	ctx    context.Context
	cancel context.CancelFunc
	// MENU,PLAY,REPLAY
	state    string
	screen   tcell.Screen
	settings settings
	defStyle tcell.Style

	menu   menu
	play   play
	replay gameReplay

	event chan tcell.Event
	wg    sync.WaitGroup
}

type historyStep struct {
	CurrGameDuration time.Duration
	// MOVE,FLAG,OPEN
	Kind  string
	MoveX int
	MoveY int
	// NONE,WON,LOST
	OpenResult string
}

func createApp(s tcell.Screen, ctx context.Context, cancel context.CancelFunc, sett settings) app {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	switch sett.Theme {
	case "LIGHT":
		defStyle = tcell.StyleDefault.Background(tcell.Color254).Foreground(tcell.Color234)
	case "DARK":
		defStyle = tcell.StyleDefault.Background(tcell.Color234).Foreground(tcell.Color254)
	}
	s.SetStyle(defStyle)

	return app{
		ctx:      ctx,
		cancel:   cancel,
		state:    "MENU",
		screen:   s,
		event:    make(chan tcell.Event, 64),
		settings: sett,
		defStyle: defStyle,
	}
}

func (a *app) draw() {
	switch a.state {
	case "PLAY":
		a.drawPlay()
	case "MENU":
		a.drawMenu()
	case "REPLAY":
		a.drawReplay()
	}
}

func (a *app) eventSenderLoop() {
	for {
		ev := a.screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				a.cancel()
				return
			}
		}

		select {
		case a.event <- ev:
		default:
		}
	}
}

func (a *app) eventReceiverLoop() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case ev := <-a.event:
			done := make(chan struct{})

			safeGo(func() {
				switch ev := ev.(type) {
				case *tcell.EventResize:
					if a.state == "PLAY" {
						a.play.fieldCurrScrollX, a.play.fieldCurrScrollY = a.alignField(a.play.fieldCurrX, a.play.fieldCurrY, a.play.fieldCurrScrollX, a.play.fieldCurrScrollY)
					}

					if a.state == "REPLAY" {
						a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY = a.alignField(a.replay.rInfo.currX, a.replay.rInfo.currY, a.replay.rInfo.currScrollX, a.replay.rInfo.currScrollY)
					}

					a.menu.savedGamesFindCurr = 0
					a.menu.savedGamesFindScreenOffset = 0
				case *tcell.EventKey:
					switch a.state {
					case "MENU":
						a.eventKeyMenu(ev)
					case "PLAY":
						a.eventKeyPlay(ev)
					case "REPLAY":
						a.eventKeyReplay(ev)
					}
				}
				done <- struct{}{}
			}, a.screen)

			select {
			case <-a.ctx.Done():
				return
			case <-done:
				switch ev.(type) {
				case *tcell.EventResize:
					a.screen.Clear()
					a.draw()
					a.screen.Sync()
				default:
					if !a.replay.rInfo.autoplayActive {
						a.screen.Clear()
						a.draw()
						a.screen.Show()
					}
				}
			}
		}
	}
}

func (a *app) getFieldScreenSize() (xOffset, yOffset, width, height int) {
	xOffset = 0
	yOffset = 1

	terminalWidth, terminalHeight := a.screen.Size()

	width = terminalWidth - xOffset
	height = terminalHeight - yOffset

	return
}

func (a *app) setContentString(x, y int, style tcell.Style, content string) {
	for i, r := range content {
		a.screen.SetContent(x+i, y, r, nil, style)
	}
}

func (a *app) getFieldScrolloffX() int {
	_, _, width, _ := a.getFieldScreenSize()

	if width >= a.settings.MaxScrolloff*2+1 {
		return a.settings.MaxScrolloff
	}

	if width%2 == 0 {
		return max(width/2-1, 0)
	} else {
		return width / 2
	}
}

func (a *app) getFieldScrolloffY() int {
	_, _, _, height := a.getFieldScreenSize()

	if height >= a.settings.MaxScrolloff*2+1 {
		return a.settings.MaxScrolloff
	}

	if height%2 == 0 {
		return max(height/2-1, 0)
	} else {
		return height / 2
	}
}

func getMineDensity(width, height, mineCount int) float64 {
	return 100 / (float64(width) * float64(height)) * float64(mineCount)
}

func getTerminesDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "./config"
	}

	terminesDir := filepath.Join(configDir, "termines")

	err = os.MkdirAll(terminesDir, 0o755)
	if err != nil {
		return "", err
	}

	return terminesDir, nil
}

func getDataFilePath() (string, error) {
	terminesDir, err := getTerminesDir()
	if err != nil {
		return "", err
	}

	dataFilePath := filepath.Join(terminesDir, "data.db")

	return dataFilePath, nil
}

func toGob[T any](d T) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(d); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func fromGob[T any](b []byte) (T, error) {
	var zero T

	var d T
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&d); err != nil {
		return zero, err
	}

	return d, nil
}
