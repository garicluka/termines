package main

import (
	"context"
	"log"

	"github.com/gdamore/tcell/v2"
)

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer safePanic(nil, s)

	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = initSettings()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	sett, err := getSettings()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	app := createApp(s, ctx, cancel, sett)

	app.createMenu()

	app.screen.Clear()
	app.draw()
	app.screen.Sync()

	app.wg.Add(1)
	safeGo(func() {
		defer app.wg.Done()
		app.eventReceiverLoop()
	}, app.screen)

	safeGo(app.eventSenderLoop, app.screen)

	app.wg.Wait()
}
